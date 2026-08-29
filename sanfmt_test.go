package sanfmt

import "testing"

func TestNormalize(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"plain pawn push", "e4", "e4"},
		{"plain piece move", "Nf3", "Nf3"},
		{"lowercase piece letter", "nf3", "Nf3"},
		{"spaced capture", "N x e4", "Nxe4"},
		{"colon capture (old notation)", "N:e4", "Nxe4"},
		{"pawn capture keeps file disambig", "exd5", "exd5"},
		{"explicit pawn letter is dropped", "Pe4", "e4"},
		{"rank disambiguation", "R1a3", "R1a3"},
		{"file disambiguation", "Nbd7", "Nbd7"},
		{"full disambiguation", "Qh4e1", "Qh4e1"},
		{"german knight letter", "Sf3", "Nf3"},
		{"castling kingside, letter O", "O-O", "O-O"},
		{"castling kingside, digit zero", "0-0", "O-O"},
		{"castling kingside, lowercase o", "o-o", "O-O"},
		{"castling kingside, no hyphens", "OO", "O-O"},
		{"castling queenside, letter O", "O-O-O", "O-O-O"},
		{"castling queenside, digit zero", "0-0-0", "O-O-O"},
		{"castling queenside, no hyphens", "000", "O-O-O"},
		{"promotion with equals", "e8=Q", "e8=Q"},
		{"promotion without equals", "e8Q", "e8=Q"},
		{"promotion with slash", "e8/Q", "e8=Q"},
		{"promotion with parens", "e8(Q)", "e8=Q"},
		{"promotion, lowercase letter", "e8q", "e8=Q"},
		{"capture and promotion combined", "exd8Q", "exd8=Q"},
		{"colon capture with promotion", "e:d8=Q", "exd8=Q"},
		{"check suffix kept", "e4+", "e4+"},
		{"check spelled ch", "e4ch", "e4+"},
		{"check spelled ch with period", "e4ch.", "e4+"},
		{"mate symbol kept", "Qh5#", "Qh5#"},
		{"mate spelled double plus", "Qh5++", "Qh5#"},
		{"mate spelled out", "Qh5mate", "Qh5#"},
		{"checkmate spelled out fully", "Qh5checkmate", "Qh5#"},
		{"castling with check suffix", "O-O+", "O-O+"},
		{"castling with mate suffix", "0-0-0#", "O-O-O#"},
		{"leading move number with dot", "1.e4", "e4"},
		{"leading move number with space", "1. e4", "e4"},
		{"leading move number for black", "12...Nf6", "Nf6"},
		{"surrounding whitespace", "  e4  ", "e4"},
		{"messy everything at once", "  N x e4 ch ", "Nxe4+"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Normalize(tc.input)
			if err != nil {
				t.Fatalf("Normalize(%q) returned error: %v", tc.input, err)
			}
			if got != tc.want {
				t.Errorf("Normalize(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestNormalizeErrors(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"only whitespace", "   "},
		{"only a move number", "1."},
		{"square off the board, file", "z9"},
		{"square off the board, rank", "Nxe9"},
		{"garbage", "banana"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Normalize(tc.input)
			if err == nil {
				t.Fatalf("Normalize(%q) = %q, want error", tc.input, got)
			}
		})
	}
}
