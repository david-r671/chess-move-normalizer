package sanfmt

import (
	"strings"
	"testing"
)

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

func TestParse(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  Move
	}{
		{"plain pawn push", "e4", Move{To: "e4"}},
		{"plain piece move", "Nf3", Move{Piece: 'N', To: "f3"}},
		{"lowercase piece letter", "nf3", Move{Piece: 'N', To: "f3"}},
		{"pawn capture keeps file disambig", "exd5", Move{From: "e", Capture: true, To: "d5"}},
		{"colon capture (old notation)", "N:e4", Move{Piece: 'N', Capture: true, To: "e4"}},
		{"rank disambiguation", "R1a3", Move{Piece: 'R', From: "1", To: "a3"}},
		{"file disambiguation", "Nbd7", Move{Piece: 'N', From: "b", To: "d7"}},
		{"full disambiguation", "Qh4e1", Move{Piece: 'Q', From: "h4", To: "e1"}},
		{"german knight letter", "Sf3", Move{Piece: 'N', To: "f3"}},
		{"promotion without equals", "e8Q", Move{To: "e8", Promotion: 'Q'}},
		{"capture and promotion combined", "exd8Q", Move{From: "e", Capture: true, To: "d8", Promotion: 'Q'}},
		{"check suffix kept", "e4+", Move{To: "e4", Suffix: "+"}},
		{"mate spelled out", "Qh5mate", Move{Piece: 'Q', To: "h5", Suffix: "#"}},
		{"castling kingside", "O-O", Move{Castle: Kingside}},
		{"castling queenside", "0-0-0", Move{Castle: Queenside}},
		{"castling with check suffix", "O-O+", Move{Castle: Kingside, Suffix: "+"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.input)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", tc.input, err)
			}
			if got != tc.want {
				t.Errorf("Parse(%q) = %+v, want %+v", tc.input, got, tc.want)
			}
			if got.String() != mustNormalize(t, tc.input) {
				t.Errorf("Parse(%q).String() = %q, want %q", tc.input, got.String(), mustNormalize(t, tc.input))
			}
		})
	}
}

func mustNormalize(t *testing.T, input string) string {
	t.Helper()
	got, err := Normalize(input)
	if err != nil {
		t.Fatalf("Normalize(%q) returned error: %v", input, err)
	}
	return got
}

func TestNormalizeMoveList(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  []string
	}{
		{
			"numbered pairs with spaces",
			"1. e4 e5 2. Nf3 Nc6",
			[]string{"e4", "e5", "Nf3", "Nc6"},
		},
		{
			"move number glued to move",
			"1.e4 e5 2.Nf3 Nc6",
			[]string{"e4", "e5", "Nf3", "Nc6"},
		},
		{
			"black move number with ellipsis",
			"1. e4 e5 2. Nf3 Nc6 3. Bb5 a6 12... Nf6",
			[]string{"e4", "e5", "Nf3", "Nc6", "Bb5", "a6", "Nf6"},
		},
		{
			"trailing decisive result is dropped",
			"1. e4 e5 2. Qh5 Nc6 3. Bc4 Nf6 4. Qxf7#  1-0",
			[]string{"e4", "e5", "Qh5", "Nc6", "Bc4", "Nf6", "Qxf7#"},
		},
		{
			"trailing black win result is dropped",
			"1. f3 e5 2. g4 Qh4mate 0-1",
			[]string{"f3", "e5", "g4", "Qh4#"},
		},
		{
			"trailing draw result is dropped",
			"1. e4 e5 1/2-1/2",
			[]string{"e4", "e5"},
		},
		{
			"in-progress marker is dropped",
			"1. e4 e5 *",
			[]string{"e4", "e5"},
		},
		{
			"loose notation mixed in",
			"1. e4 e5 2. N x f3 ch",
			[]string{"e4", "e5", "Nxf3+"},
		},
		{
			"empty input yields no moves",
			"",
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NormalizeMoveList(tc.input)
			if err != nil {
				t.Fatalf("NormalizeMoveList(%q) returned error: %v", tc.input, err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("NormalizeMoveList(%q) = %v, want %v", tc.input, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("NormalizeMoveList(%q)[%d] = %q, want %q", tc.input, i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestNormalizeMoveListErrors(t *testing.T) {
	got, err := NormalizeMoveList("1. e4 banana 2. Nf3 Nc6")
	if err == nil {
		t.Fatalf("NormalizeMoveList returned no error for a bad token")
	}
	if !strings.Contains(err.Error(), "banana") {
		t.Errorf("error %v does not mention the bad token", err)
	}
	want := []string{"e4", "Nf3", "Nc6"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
		}
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
