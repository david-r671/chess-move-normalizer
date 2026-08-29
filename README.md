# chess-move-normalizer

Chess move notation is standardized on paper but not in practice. Pull
moves out of an old book, a hand-copied scoresheet, or a database
exported by three different tools and you'll see the same move written
several different ways:

```
0-0        O-O
N:e4       Nxe4
e8Q        e8=Q
Qh5ch      Qh5+
1. e4      e4
```

None of that changes what the move *is*, but it breaks anything that
does string comparison or lookup on move text. `sanfmt` takes messy
algebraic notation in and gives clean SAN out.

It does not know chess rules. It has no board and cannot tell you
whether a move is legal - it only cleans up how a syntactically valid
algebraic move is written. Descriptive notation ("P-K4", "N-KB3") is
out of scope.

## What it normalizes

- Piece letters: lowercase or German `S` (Springer) -> uppercase, `S` -> `N`
- Captures: `x`, `X`, `:`, or spaced (`N x e4`) -> `x`
- Castling: `0-0`, `o-o`, `OO` -> `O-O`; same idea for queenside
- Promotion: `e8Q`, `e8/Q`, `e8(Q)` -> `e8=Q`
- Check / mate: `ch`, `ch.`, `++`, `mate`, `checkmate` -> `+` / `#`
- Leading move numbers: `1.`, `1. `, `12...` are stripped
- Stray whitespace anywhere in the token

## Usage

As a library:

```go
package main

import (
	"fmt"

	sanfmt "github.com/david-r671/chess-move-normalizer"
)

func main() {
	clean, err := sanfmt.Normalize("N x e4 ch")
	if err != nil {
		panic(err)
	}
	fmt.Println(clean) // "Nxe4+"
}
```

As a command line filter, one move per line:

```
$ printf '1. e4\n0-0\ne8Q\nQh5mate\n' | go run ./cmd/sanfmt
e4
O-O
e8=Q
Qh5#
```

Moves that don't parse are reported on stderr and the command exits
non-zero, without stopping processing of the rest of the input.

## Status

Single-move normalization only - no game/PGN parsing yet, and no
legality checking. See the test file for the exact set of input forms
currently handled.

## License

MIT, see LICENSE.
