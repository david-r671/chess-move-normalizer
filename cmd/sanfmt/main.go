// Command sanfmt normalizes chess moves from stdin (or command-line
// arguments) into clean SAN, one move per line.
package main

import (
	"bufio"
	"fmt"
	"os"

	sanfmt "github.com/david-r671/chess-move-normalizer"
)

func main() {
	if len(os.Args) > 1 {
		os.Exit(run(os.Args[1:]))
	}

	var moves []string
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
	os.Exit(run(moves))
}

func run(moves []string) int {
	failures := 0
	for _, move := range moves {
		clean, err := sanfmt.Normalize(move)
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
