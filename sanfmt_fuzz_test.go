package sanfmt

import "testing"

// FuzzNormalize seeds from the same inputs as TestNormalize and
// TestNormalizeErrors, then checks two things no table can exhaustively
// cover: Normalize must never panic on arbitrary text, and its output
// must be a fixed point - normalizing an already-clean move has to
// return that same move unchanged.
func FuzzNormalize(f *testing.F) {
	seeds := []string{
		"e4", "Nf3", "nf3", "N x e4", "N:e4", "exd5", "Pe4", "R1a3", "Nbd7", "Qh4e1", "Sf3",
		"O-O", "0-0", "o-o", "OO", "O-O-O", "0-0-0", "000",
		"e8=Q", "e8Q", "e8/Q", "e8(Q)", "e8q", "exd8Q", "e:d8=Q",
		"e4+", "e4ch", "e4ch.", "Qh5#", "Qh5++", "Qh5mate", "Qh5checkmate",
		"O-O+", "0-0-0#",
		"1.e4", "1. e4", "12...Nf6", "  e4  ", "  N x e4 ch ",
		"", "   ", "1.", "z9", "Nxe9", "banana",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		clean, err := Normalize(input)
		if err != nil {
			return
		}

		again, err := Normalize(clean)
		if err != nil {
			t.Fatalf("Normalize(%q) = %q, but re-normalizing that output failed: %v", input, clean, err)
		}
		if again != clean {
			t.Fatalf("Normalize is not idempotent: Normalize(%q) = %q, Normalize(%q) = %q", input, clean, clean, again)
		}
	})
}
