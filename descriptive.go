package sanfmt

import (
	"fmt"
	"regexp"
	"strings"
)

// descMoveRe matches a descriptive move body after suffix, promotion, and
// castling have already been stripped off: an optional side-qualified
// piece letter, an optional dash, a side-qualified file, and a rank digit.
// The file token can never appear bare (just "B" or "N") because
// descriptive notation always names which side's bishop, knight, or rook
// file a square sits on - that's what makes the square unambiguous
// without a board.
var descMoveRe = regexp.MustCompile(`^(QN|KN|QB|KB|QR|KR|Q|K|N|B|R|P)-?(QR|QN|QB|KR|KN|KB|Q|K)([1-8])$`)

// descriptiveFiles maps a descriptive file name to its algebraic file,
// following the back-rank order every game starts from: rook, knight,
// bishop, queen, king, bishop, knight, rook.
var descriptiveFiles = map[string]byte{
	"QR": 'a', "QN": 'b', "QB": 'c', "Q": 'd',
	"K": 'e', "KB": 'f', "KN": 'g', "KR": 'h',
}

// descriptivePieces maps a side-qualified piece token to its plain
// algebraic piece letter. The side qualifier only exists to tell two
// pieces of the same type apart in the source text; it says nothing about
// which file the piece is currently on, so it can't be carried through as
// SAN disambiguation and is dropped.
var descriptivePieces = map[string]byte{
	"Q": 'Q', "K": 'K',
	"N": 'N', "QN": 'N', "KN": 'N',
	"B": 'B', "QB": 'B', "KB": 'B',
	"R": 'R', "QR": 'R', "KR": 'R',
}

// ParseDescriptive takes a single move in descriptive notation - "P-K4",
// "N-KB3", "QR-K1", "Q-KR5ch" - and breaks it down into a Move.
//
// Descriptive squares are named relative to the player making the move
// ("K4" is e4 for White but e5 for Black), so the caller must say whose
// move it is via white. Everything else - suffixes, castling, promotion -
// is handled the same way as Parse.
//
// Descriptive captures name the piece being captured rather than a
// destination square ("NxP", "QPxP"), which this package cannot resolve
// without knowing where that piece actually sits on the board, so those
// return an error instead of a guess - as does a rook, knight, or bishop
// square with no Q/K qualifier ("Q-R5"), since which wing it's on is
// likewise something only a board can settle.
func ParseDescriptive(input string, white bool) (Move, error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return Move{}, ErrEmpty
	}

	s = moveNumberRe.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	if s == "" {
		return Move{}, ErrEmpty
	}

	compact := strings.Join(strings.Fields(s), "")
	if compact == "" {
		return Move{}, ErrEmpty
	}

	suffix, rest := extractSuffix(compact)
	if rest == "" {
		return Move{}, fmt.Errorf("sanfmt: %q has no move body", input)
	}

	if kingside, ok := castleSide(rest); ok {
		side := Queenside
		if kingside {
			side = Kingside
		}
		return Move{Castle: side, Suffix: suffix}, nil
	}

	// "Kt" is the older British spelling for knight; fold it to "N"
	// before matching so both spellings share one code path.
	rest = strings.ToUpper(rest)
	rest = strings.ReplaceAll(rest, "KT", "N")

	if strings.ContainsAny(rest, "X:") {
		return Move{}, fmt.Errorf("sanfmt: %q: descriptive captures name the captured piece, not a square, and can't be resolved without a board", input)
	}

	var promo byte
	if loc := promoRe.FindStringSubmatchIndex(rest); loc != nil {
		promo = normalizePieceLetter(rest[loc[2]:loc[3]])[0]
		rest = rest[:loc[0]]
	}

	groups := descMoveRe.FindStringSubmatch(rest)
	if groups == nil {
		return Move{}, fmt.Errorf("sanfmt: %q: unrecognized descriptive move syntax %q", input, rest)
	}

	var piece byte
	if groups[1] != "P" {
		piece = descriptivePieces[groups[1]]
	}

	to := string(descriptiveFiles[groups[2]]) + string(descriptiveRank(groups[3][0], white))

	return Move{
		Piece:     piece,
		To:        to,
		Promotion: promo,
		Suffix:    suffix,
	}, nil
}

// NormalizeDescriptive takes a single move in descriptive notation and
// returns it as clean SAN. See ParseDescriptive for what is and isn't
// supported.
func NormalizeDescriptive(input string, white bool) (string, error) {
	m, err := ParseDescriptive(input, white)
	if err != nil {
		return "", err
	}
	return m.String(), nil
}

// descriptiveRank converts a descriptive rank digit, counted from the
// mover's own back rank, into the algebraic rank counted from White's
// back rank.
func descriptiveRank(digit byte, white bool) byte {
	if white {
		return digit
	}
	return '0' + (9 - (digit - '0'))
}
