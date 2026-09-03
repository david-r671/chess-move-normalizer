// Package sanfmt normalizes chess moves written in loose, inconsistent
// algebraic notation into clean Standard Algebraic Notation (SAN).
//
// Moves collected from old books, hand transcriptions, or different
// database exports disagree on details that don't change the move
// itself: "0-0" vs "O-O", "N:e4" vs "Nxe4", "e8Q" vs "e8=Q", "Qh5ch" vs
// "Qh5+". This package resolves that surface variation into one form
// without knowing anything about legal chess positions.
package sanfmt

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ErrEmpty is returned when nothing is left to parse after stripping
// whitespace and any leading move number.
var ErrEmpty = errors.New("sanfmt: empty move")

var (
	moveNumberRe     = regexp.MustCompile(`^\d+\.+\s*`)
	moveNumberOnlyRe = regexp.MustCompile(`^\d+\.+$`)
	resultRe         = regexp.MustCompile(`^(1-0|0-1|1/2-1/2|\*)$`)
	castleRe         = regexp.MustCompile(`^[Oo0](-?[Oo0]){1,2}$`)
	promoRe          = regexp.MustCompile(`(?i)(?:[=/])?\(?([qrbns])\)?$`)
	baseRe           = regexp.MustCompile(`^([KQRBNkqrbnSs]?)([a-h]?[1-8]?)([xX:]?)([a-h][1-8])$`)
)

// suffixPatterns must stay in this order: "checkmate" has to be tried
// before "mate" or the shorter pattern would strip only the tail and
// leave "check" dangling on the move body.
var suffixPatterns = []struct {
	re     *regexp.Regexp
	symbol string
}{
	{regexp.MustCompile(`(?i)checkmate$`), "#"},
	{regexp.MustCompile(`(?i)mate$`), "#"},
	{regexp.MustCompile(`\+\+$`), "#"}, // old descriptive-era double check mark for mate
	{regexp.MustCompile(`#$`), "#"},
	{regexp.MustCompile(`(?i)ch\.?$`), "+"},
	{regexp.MustCompile(`\+$`), "+"},
}

// Normalize takes a single move in loosely formatted algebraic notation
// and returns it as clean SAN: piece letters uppercase (pawns carry no
// letter), captures written as "x", promotions as "=Q", castling as
// "O-O" / "O-O-O", and check/mate as "+" / "#".
//
// Normalize only cleans up notation - it has no board and cannot tell
// whether the move is legal. Descriptive notation ("P-K4") is out of
// scope; only algebraic input is accepted.
func Normalize(input string) (string, error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return "", ErrEmpty
	}

	s = moveNumberRe.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	if s == "" {
		return "", ErrEmpty
	}

	compact := strings.Join(strings.Fields(s), "")
	if compact == "" {
		return "", ErrEmpty
	}

	suffix, rest := extractSuffix(compact)
	if rest == "" {
		return "", fmt.Errorf("sanfmt: %q has no move body", input)
	}

	if kingside, ok := castleSide(rest); ok {
		if kingside {
			return "O-O" + suffix, nil
		}
		return "O-O-O" + suffix, nil
	}

	body, err := parseMove(rest)
	if err != nil {
		return "", fmt.Errorf("sanfmt: %q: %w", input, err)
	}
	return body + suffix, nil
}

// NormalizeMoveList splits a PGN-style movetext string - the part of a
// game record after the tag pairs, e.g. "1. e4 e5 2. Nf3 Nc6 1-0" - into
// individual moves and normalizes each one. Move number tokens ("1.",
// "12...") and game results ("1-0", "0-1", "1/2-1/2", "*") are recognized
// and dropped rather than treated as moves; a move number glued to the
// following move ("1.e4") is handled the same way by Normalize itself.
//
// Comments in braces, NAG codes ($1), and parenthesized variations are
// not supported - movetext must be a plain move sequence.
//
// NormalizeMoveList returns every move that normalized successfully, in
// order. If one or more tokens failed to parse, it also returns a
// non-nil error built with errors.Join describing all of them; the
// caller can still use the moves that did succeed.
func NormalizeMoveList(input string) ([]string, error) {
	var moves []string
	var errs []error
	for _, field := range strings.Fields(input) {
		if moveNumberOnlyRe.MatchString(field) || resultRe.MatchString(field) {
			continue
		}
		clean, err := Normalize(field)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		moves = append(moves, clean)
	}
	return moves, errors.Join(errs...)
}

func extractSuffix(s string) (suffix, rest string) {
	for _, p := range suffixPatterns {
		if loc := p.re.FindStringIndex(s); loc != nil {
			return p.symbol, s[:loc[0]]
		}
	}
	return "", s
}

func castleSide(s string) (kingside, ok bool) {
	if !castleRe.MatchString(s) {
		return false, false
	}
	n := strings.Count(strings.ToUpper(s), "O") + strings.Count(s, "0")
	return n == 2, true
}

func parseMove(s string) (string, error) {
	promo := ""
	if loc := promoRe.FindStringSubmatchIndex(s); loc != nil {
		letter := normalizePieceLetter(s[loc[2]:loc[3]])
		promo = "=" + letter
		s = s[:loc[0]]
	}

	m := baseRe.FindStringSubmatch(s)
	if m == nil {
		return "", fmt.Errorf("unrecognized move syntax %q", s)
	}

	piece := normalizePieceLetter(m[1])
	disambig := m[2]
	capture := ""
	if m[3] != "" {
		capture = "x"
	}
	dest := m[4]

	return piece + disambig + capture + dest + promo, nil
}

// normalizePieceLetter uppercases a piece letter and maps the German
// "S" (Springer) to "N" for knight. An empty string passes through
// unchanged, since pawns have no letter.
func normalizePieceLetter(letter string) string {
	upper := strings.ToUpper(letter)
	if upper == "S" {
		return "N"
	}
	return upper
}
