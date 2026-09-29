package cube_test

import (
	"testing"

	"cube-auth/internal/cube"
)

func TestSolvedCubeIsLegal(t *testing.T) {
	report := cube.Validate(solved())
	if !report.Legal {
		t.Fatalf("solved cube rejected: %#v", report.Violation)
	}
	if report.Cube.CornerPositions != [8]int{0, 1, 2, 3, 4, 5, 6, 7} {
		t.Errorf("corner positions = %v", report.Cube.CornerPositions)
	}
	if report.Cube.EdgePositions != [12]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11} {
		t.Errorf("edge positions = %v", report.Cube.EdgePositions)
	}
	if report.Cube.CornerOrientations != [8]int{} {
		t.Errorf("corner orientations = %v, want all zero", report.Cube.CornerOrientations)
	}
	if report.Cube.EdgeOrientations != [12]int{} {
		t.Errorf("edge orientations = %v, want all zero", report.Cube.EdgeOrientations)
	}
}

func TestLegalFaceTurnsFromSolved(t *testing.T) {
	t.Run("every single face turn", func(t *testing.T) {
		for _, move := range []byte("URFDLB") {
			t.Run(string(move), func(t *testing.T) {
				report := cube.Validate(turn(solved(), move))
				if !report.Legal {
					t.Fatalf("legal move %q rejected: %#v", move, report.Violation)
				}
				assertOrientationInvariants(t, report)
			})
		}
	})

	tests := []struct {
		name  string
		moves []byte
	}{
		{"each face once", []byte{'U', 'R', 'F', 'D', 'L', 'B'}},
		{"four repetitions return to solved", []byte{'R', 'R', 'R', 'R'}},
		{"three turns equal inverse", []byte{'F', 'F', 'F'}},
		{"mixed sequence", []byte{'U', 'U', 'R', 'R', 'R', 'F', 'L', 'L', 'D', 'B', 'B', 'B'}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			faces := solved()
			for _, move := range tt.moves {
				faces = turn(faces, move)
			}
			report := cube.Validate(faces)
			if !report.Legal {
				t.Fatalf("legal moves %s rejected: %#v", string(tt.moves), report.Violation)
			}
			assertOrientationInvariants(t, report)
		})
	}
}

func TestSingleCornerTwist(t *testing.T) {
	faces := solved()
	// Twist URF by cycling its U and R stickers.
	faces["U"][8], faces["R"][0] = faces["R"][0], faces["U"][8]

	report := cube.Validate(faces)
	if report.Legal {
		t.Fatal("single twisted corner was accepted")
	}
	if got, want := report.Violation.Code, "corner_orientation"; got != want {
		t.Fatalf("violation = %q, want %q (%#v)", got, want, report.Violation)
	}
	if len(report.Violation.Pieces) != 1 || report.Violation.Pieces[0].Piece != "URF" {
		t.Fatalf("involved pieces = %#v", report.Violation.Pieces)
	}
}

func TestSingleEdgeFlip(t *testing.T) {
	faces := solved()
	// Flip the UF edge by exchanging its two stickers.
	faces["U"][7], faces["F"][1] = faces["F"][1], faces["U"][7]

	report := cube.Validate(faces)
	if report.Legal {
		t.Fatal("single flipped edge was accepted")
	}
	if got, want := report.Violation.Code, "edge_orientation"; got != want {
		t.Fatalf("violation = %q, want %q (%#v)", got, want, report.Violation)
	}
	if len(report.Violation.Pieces) != 1 || report.Violation.Pieces[0].Piece != "UF" {
		t.Fatalf("involved pieces = %#v", report.Violation.Pieces)
	}
}

func TestTwoEdgeExchangeHasParityViolation(t *testing.T) {
	faces := solved()
	// Exchange UR and UF while keeping both edges unflipped.
	swapFaceStickers(faces,
		[2]sticker{{"U", 5}, {"R", 1}},
		[2]sticker{{"U", 7}, {"F", 1}},
	)

	report := cube.Validate(faces)
	if report.Legal {
		t.Fatal("a single transposition of two edges was accepted")
	}
	if got, want := report.Violation.Code, "permutation_parity"; got != want {
		t.Fatalf("violation = %q, want %q (%#v)", got, want, report.Violation)
	}
	assertContainsPiece(t, report.Violation.Parity.Pieces, "UR")
	assertContainsPiece(t, report.Violation.Parity.Pieces, "UF")
}

func TestDuplicatePiecesAreReportedBeforeOrientation(t *testing.T) {
	faces := solved()
	// Turn UF into a second UR and DR into a second DF by exchanging their
	// side stickers. This preserves the nine-sticker color counts, so color
	// counting alone cannot catch it.
	faces["F"][1], faces["R"][7] = faces["R"][7], faces["F"][1]

	report := cube.Validate(faces)
	if report.Legal {
		t.Fatal("duplicate edge pieces were accepted")
	}
	if got, want := report.Violation.Code, "piece_count"; got != want {
		t.Fatalf("violation = %q, want %q (%#v)", got, want, report.Violation)
	}
	assertIssueKindsContain(t, report.Violation.Pieces, "duplicate_edge", "missing_edge")
}

func TestColorCountsAreCheckedBeforePieces(t *testing.T) {
	faces := solved()
	faces["U"][0] = faces["D"][4]
	report := cube.Validate(faces)
	if report.Legal || report.Violation.Code != "colors" {
		t.Fatalf("got legal=%v violation=%#v", report.Legal, report.Violation)
	}
}

type sticker struct {
	face  string
	index int
}

func solved() cube.Faces {
	colors := map[string]string{
		"U": "W",
		"R": "G",
		"F": "O",
		"D": "Y",
		"L": "B",
		"B": "X",
	}
	faces := cube.Faces{}
	for face, color := range colors {
		faces[face] = make([]string, 9)
		for i := range faces[face] {
			faces[face][i] = color
		}
	}
	return faces
}

func swapFaceStickers(faces cube.Faces, a, b [2]sticker) {
	av := [2]string{faces[a[0].face][a[0].index], faces[a[1].face][a[1].index]}
	bv := [2]string{faces[b[0].face][b[0].index], faces[b[1].face][b[1].index]}
	faces[a[0].face][a[0].index], faces[a[1].face][a[1].index] = bv[0], bv[1]
	faces[b[0].face][b[0].index], faces[b[1].face][b[1].index] = av[0], av[1]
}

func assertOrientationInvariants(t *testing.T, report *cube.Report) {
	t.Helper()
	cornerSum := 0
	for i, co := range report.Cube.CornerOrientations {
		if co < 0 || co > 2 {
			t.Errorf("corner orientation at position %d = %d, want 0..2", i, co)
		}
		cornerSum += co
	}
	if cornerSum%3 != 0 {
		t.Errorf("corner orientation sum = %d, want 0 mod 3", cornerSum)
	}

	edgeSum := 0
	for i, eo := range report.Cube.EdgeOrientations {
		if eo != 0 && eo != 1 {
			t.Errorf("edge orientation at position %d = %d, want 0 or 1", i, eo)
		}
		edgeSum += eo
	}
	if edgeSum%2 != 0 {
		t.Errorf("edge orientation sum = %d, want 0 mod 2", edgeSum)
	}
}

func assertContainsPiece(t *testing.T, pieces []string, want string) {
	t.Helper()
	for _, got := range pieces {
		if got == want {
			return
		}
	}
	t.Errorf("pieces %v do not contain %q", pieces, want)
}

func assertIssueKindsContain(t *testing.T, issues []cube.PieceIssue, kinds ...string) {
	t.Helper()
	seen := map[string]bool{}
	for _, issue := range issues {
		seen[issue.Kind] = true
	}
	for _, kind := range kinds {
		if !seen[kind] {
			t.Errorf("piece issues %v do not contain kind %q", issues, kind)
		}
	}
}
