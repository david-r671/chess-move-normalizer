package sanfmt

import "testing"

func TestNormalizeDescriptive(t *testing.T) {
	cases := []struct {
		name  string
		input string
		white bool
		want  string
	}{
		{"white king pawn push", "P-K4", true, "e4"},
		{"black king pawn push", "P-K4", false, "e5"},
		{"white queen pawn push", "P-Q4", true, "d4"},
		{"black queen pawn push", "P-Q4", false, "d5"},
		{"knight to kingside bishop 3", "N-KB3", true, "Nf3"},
		{"knight to queenside bishop 3", "N-QB3", true, "Nc3"},
		{"old Kt spelling for knight", "Kt-KB3", true, "Nf3"},
		{"lowercase kt spelling", "kt-kb3", true, "Nf3"},
		{"disambiguated queen's rook", "QR-K1", true, "Re1"},
		{"disambiguated king's rook", "KR-K1", true, "Re1"},
		{"disambiguated queen's knight, old spelling", "QKt-QB3", true, "Nc3"},
		{"queen to kingside rook 5, white", "Q-KR5", true, "Qh5"},
		{"queen to kingside rook 5, black", "Q-KR5", false, "Qh4"},
		{"bishop to queenside knight 5", "B-QN5", true, "Bb5"},
		{"king move", "K-KB1", true, "Kf1"},
		{"no dash", "NKB3", true, "Nf3"},
		{"promotion with parens, white", "P-K8(Q)", true, "e8=Q"},
		{"promotion with parens, black", "P-K8(Q)", false, "e1=Q"},
		{"promotion with equals", "P-Q8=Q", true, "d8=Q"},
		{"check suffix", "Q-KR5ch", true, "Qh5+"},
		{"mate spelled out", "Q-KR5mate", true, "Qh5#"},
		{"castling kingside", "O-O", true, "O-O"},
		{"castling queenside, digit zero", "0-0-0", false, "O-O-O"},
		{"leading move number", "1. P-K4", true, "e4"},
		{"surrounding whitespace", "  P-K4  ", true, "e4"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NormalizeDescriptive(tc.input, tc.white)
			if err != nil {
				t.Fatalf("NormalizeDescriptive(%q, %v) returned error: %v", tc.input, tc.white, err)
			}
			if got != tc.want {
				t.Errorf("NormalizeDescriptive(%q, %v) = %q, want %q", tc.input, tc.white, got, tc.want)
			}
		})
	}
}

func TestParseDescriptive(t *testing.T) {
	cases := []struct {
		name  string
		input string
		white bool
		want  Move
	}{
		{"pawn push", "P-K4", true, Move{To: "e4"}},
		{"knight move", "N-KB3", true, Move{Piece: 'N', To: "f3"}},
		{"promotion", "P-K8(Q)", true, Move{To: "e8", Promotion: 'Q'}},
		{"check suffix", "Q-KR5ch", true, Move{Piece: 'Q', To: "h5", Suffix: "+"}},
		{"castling", "O-O", true, Move{Castle: Kingside}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseDescriptive(tc.input, tc.white)
			if err != nil {
				t.Fatalf("ParseDescriptive(%q, %v) returned error: %v", tc.input, tc.white, err)
			}
			if got != tc.want {
				t.Errorf("ParseDescriptive(%q, %v) = %+v, want %+v", tc.input, tc.white, got, tc.want)
			}
		})
	}
}

func TestNormalizeDescriptiveErrors(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"only whitespace", "   "},
		{"only a move number", "1."},
		{"capture names a piece, not a square", "NxP"},
		{"pawn capture names a piece, not a square", "QPxP"},
		{"colon capture", "N:P"},
		{"bare bishop file is not a real descriptive square", "B-N5"},
		{"unqualified rook file needs a board to resolve", "Q-R5"},
		{"garbage", "banana"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NormalizeDescriptive(tc.input, true)
			if err == nil {
				t.Fatalf("NormalizeDescriptive(%q) = %q, want error", tc.input, got)
			}
		})
	}
}
