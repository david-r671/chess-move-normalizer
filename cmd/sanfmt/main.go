// Command sanfmt normalizes chess moves from stdin (or command-line
// arguments) into clean SAN, one move per line.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	sanfmt "github.com/david-r671/chess-move-normalizer"
)

func main() {
	descriptive := flag.Bool("descriptive", false, `parse moves as descriptive notation ("P-K4") instead of algebraic`)
	black := flag.Bool("black", false, "the first move is Black's rather than White's; sides alternate after that")
	pgn := flag.Bool("pgn", false, "treat the input as PGN movetext: drop move numbers, results, comments, NAGs and variations")
	flag.Parse()

	if *pgn {
		if *descriptive || *black {
			fmt.Fprintln(os.Stderr, "sanfmt: -pgn cannot be combined with -descriptive or -black")
			os.Exit(2)
		}
		var text string
		if args := flag.Args(); len(args) > 0 {
			text = strings.Join(args, " ")
		} else {
			// Read everything at once: a brace comment can span lines,
			// so the text can't be split per line before it is parsed.
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintln(os.Stderr, "sanfmt: reading stdin:", err)
				os.Exit(1)
			}
			text = string(data)
		}
		os.Exit(runPGN(text))
	}

	var moves []string
	if args := flag.Args(); len(args) > 0 {
		moves = args
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			moves = append(moves, line)
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "sanfmt: reading stdin:", err)
			os.Exit(1)
		}
	}
	os.Exit(run(moves, *descriptive, !*black))
}

// runPGN normalizes a block of movetext and prints the mainline moves,
// one per line. Moves that did normalize are printed even when other
// tokens failed, matching what NormalizeMoveList returns.
func runPGN(text string) int {
	moves, err := sanfmt.NormalizeMoveList(text)
	for _, m := range moves {
		fmt.Println(m)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

// run normalizes each move in turn. In descriptive mode, white tracks
// whose move each successive entry is; it alternates regardless of
// whether a given move fails to parse, since ply order is a property
// of the game, not of the input's validity.
func run(moves []string, descriptive bool, white bool) int {
	failures := 0
	for _, move := range moves {
		var clean string
		var err error
		if descriptive {
			clean, err = sanfmt.NormalizeDescriptive(move, white)
			white = !white
		} else {
			clean, err = sanfmt.Normalize(move)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			failures++
			continue
		}
		fmt.Println(clean)
	}
	if failures > 0 {
		return 1
	}
	return 0
}
