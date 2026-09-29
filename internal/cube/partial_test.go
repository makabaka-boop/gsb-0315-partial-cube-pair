package cube_test

import (
	"testing"

	"cube-auth/internal/cube"
)

// swapCornerPieces exchanges the corner pieces at two positions while
// keeping each piece's orientation at the new position.
func swapCornerPieces(faces cube.Faces, posA, posB int) {
	a, b := cube.CornerRefs[posA].Stickers, cube.CornerRefs[posB].Stickers
	for i := range a {
		ca := getSticker(faces, a[i])
		cb := getSticker(faces, b[i])
		setSticker(faces, a[i], cb)
		setSticker(faces, b[i], ca)
	}
}

// swapEdgePieces exchanges the edge pieces at two positions while keeping
// each piece's orientation at the new position.
func swapEdgePieces(faces cube.Faces, posA, posB int) {
	a, b := cube.EdgeRefs[posA].Stickers, cube.EdgeRefs[posB].Stickers
	for i := range a {
		ca := getSticker(faces, a[i])
		cb := getSticker(faces, b[i])
		setSticker(faces, a[i], cb)
		setSticker(faces, b[i], ca)
	}
}

// flipEdge exchanges the two stickers of one edge position, flipping the
// piece that sits there in place.
func flipEdge(faces cube.Faces, pos int) {
	s := cube.EdgeRefs[pos].Stickers
	c0 := getSticker(faces, s[0])
	setSticker(faces, s[0], getSticker(faces, s[1]))
	setSticker(faces, s[1], c0)
}

// twistCorner cycles the three stickers of one corner position, shifting the
// U/D color to the next side slot.
func twistCorner(faces cube.Faces, pos int) {
	s := cube.CornerRefs[pos].Stickers
	c0 := getSticker(faces, s[0])
	setSticker(faces, s[0], getSticker(faces, s[2]))
	setSticker(faces, s[2], getSticker(faces, s[1]))
	setSticker(faces, s[1], c0)
}

func getSticker(faces cube.Faces, ref cube.StickerRef) string {
	name := []string{"U", "R", "F", "D", "L", "B"}[ref.Face]
	return faces[name][ref.Index]
}

func setSticker(faces cube.Faces, ref cube.StickerRef, color string) {
	name := []string{"U", "R", "F", "D", "L", "B"}[ref.Face]
	faces[name][ref.Index] = color
}

func mask(refs []cube.StickerRef, faces cube.Faces) {
	for _, ref := range refs {
		setSticker(faces, ref, "?")
	}
}

func assertAgreesWithObservations(t *testing.T, partial, full cube.Faces) {
	t.Helper()
	for _, face := range []string{"U", "R", "F", "D", "L", "B"} {
		for i, observed := range partial[face] {
			if observed != "?" && full[face][i] != observed {
				t.Errorf("completion %s[%d] = %q, disagrees with observed %q",
					face, i, full[face][i], observed)
			}
			if full[face][i] == "?" {
				t.Errorf("completion still contains %q at %s[%d]", "?", face, i)
			}
		}
	}
}

// The inspector's record: two corners are visibly swapped and the four
// stickers of two edges are unreadable. Swapping those two edges completes
// the legal state, so a greedy solver that leaves the edges in place wrongly
// reports no_completion.
func TestPartialTwoEdgeTranspositionCompletesCornerTransposition(t *testing.T) {
	faces := solved()
	swapCornerPieces(faces, 0, 1) // URF <-> UFL
	swapEdgePieces(faces, 0, 1)   // UR  <-> UF
	if r := cube.Validate(faces); !r.Legal {
		t.Fatalf("full state should be legal: %#v", r.Violation)
	}

	masked := faces
	mask(append([]cube.StickerRef(nil), cube.EdgeRefs[0].Stickers[:]...), masked)
	mask(append([]cube.StickerRef(nil), cube.EdgeRefs[1].Stickers[:]...), masked)

	report := cube.ValidatePartial(masked)
	if !report.Legal {
		t.Fatalf("completable state rejected: %#v", report.Violation)
	}
	if report.Ambiguous == nil || !*report.Ambiguous {
		t.Errorf("ambiguous = %v, want true: the two edges may or may not be exchanged", report.Ambiguous)
	}
	if report.Completion == nil {
		t.Fatal("legal partial report lacks completion")
	}
	assertAgreesWithObservations(t, masked, report.Completion)
	if r := cube.Validate(report.Completion); !r.Legal {
		t.Errorf("reported completion is not reachable: %#v", r.Violation)
	}
}

// Unknown stickers spread across several corners must not make a reachable
// state look impossible.
func TestPartialUnknownStickersAcrossMultipleCorners(t *testing.T) {
	faces := solved()
	for _, move := range []byte{'R', 'R', 'U', 'F', 'F', 'L'} {
		faces = turn(faces, move)
	}
	if r := cube.Validate(faces); !r.Legal {
		t.Fatalf("scrambled state should be legal: %#v", r.Violation)
	}

	var hidden []cube.StickerRef
	for _, pos := range []int{0, 3} { // URF, UBR: whole pieces unreadable
		hidden = append(hidden, cube.CornerRefs[pos].Stickers[:]...)
	}
	mask(hidden, faces)

	report := cube.ValidatePartial(faces)
	if !report.Legal {
		t.Fatalf("completable scrambled state rejected: %#v", report.Violation)
	}
	assertAgreesWithObservations(t, faces, report.Completion)
}

// Unknown stickers spread across several edges must not make a reachable
// state look impossible.
func TestPartialUnknownStickersAcrossMultipleEdges(t *testing.T) {
	faces := solved()
	for _, move := range []byte{'U', 'R', 'R', 'B', 'D', 'F', 'F'} {
		faces = turn(faces, move)
	}
	if r := cube.Validate(faces); !r.Legal {
		t.Fatalf("scrambled state should be legal: %#v", r.Violation)
	}

	var hidden []cube.StickerRef
	for _, pos := range []int{2, 7} { // UL, DB: whole pieces unreadable
		hidden = append(hidden, cube.EdgeRefs[pos].Stickers[:]...)
	}
	mask(hidden, faces)

	report := cube.ValidatePartial(faces)
	if !report.Legal {
		t.Fatalf("completable scrambled state rejected: %#v", report.Violation)
	}
	assertAgreesWithObservations(t, faces, report.Completion)
}

// A fully unreadable edge may need to be flipped to close the orientation
// sum; always trying orientation zero first yields a false no_completion.
func TestPartialUnknownEdgeRecoversRequiredFlip(t *testing.T) {
	faces := solved()
	// Flip both UF and DF in place; two flips keep the orientation sum even.
	flipEdge(faces, 1)
	flipEdge(faces, 5)
	if r := cube.Validate(faces); !r.Legal {
		t.Fatalf("two flipped edges are legal: %#v", r.Violation)
	}

	mask(append([]cube.StickerRef(nil), cube.EdgeRefs[1].Stickers[:]...), faces)

	report := cube.ValidatePartial(faces)
	if !report.Legal {
		t.Fatalf("state with a flipped unreadable edge rejected: %#v", report.Violation)
	}
	if report.Ambiguous == nil || *report.Ambiguous {
		t.Errorf("ambiguous = %v, want false: the flip is forced", report.Ambiguous)
	}
	if report.Cube.EdgeOrientations[1] != 1 {
		t.Errorf("edge orientation at UF = %d, want 1 (flipped)", report.Cube.EdgeOrientations[1])
	}
	assertAgreesWithObservations(t, faces, report.Completion)
}

// A fully unreadable corner may need a nonzero twist to close the
// orientation sum; always trying twist zero first yields a false
// no_completion.
func TestPartialUnknownCornerRecoversRequiredTwist(t *testing.T) {
	faces := solved()
	twistCorner(faces, 0) // URF +1
	twistCorner(faces, 1)
	twistCorner(faces, 1) // UFL +2, total twist 0 mod 3
	if r := cube.Validate(faces); !r.Legal {
		t.Fatalf("two compensating corner twists are legal: %#v", r.Violation)
	}

	mask(append([]cube.StickerRef(nil), cube.CornerRefs[0].Stickers[:]...), faces)

	report := cube.ValidatePartial(faces)
	if !report.Legal {
		t.Fatalf("state with a twisted unreadable corner rejected: %#v", report.Violation)
	}
	if report.Ambiguous == nil || *report.Ambiguous {
		t.Errorf("ambiguous = %v, want false: the twist is forced", report.Ambiguous)
	}
	if report.Cube.CornerOrientations[0] != 1 {
		t.Errorf("corner orientation at URF = %d, want 1", report.Cube.CornerOrientations[0])
	}
	assertAgreesWithObservations(t, faces, report.Completion)
}

// With only the centers known, many legal completions exist; the report must
// say the record still needs another sticker photographed.
func TestPartialCentersOnlyIsAmbiguous(t *testing.T) {
	faces := solved()
	for _, face := range []string{"U", "R", "F", "D", "L", "B"} {
		for i := 0; i < 9; i++ {
			if i != 4 {
				faces[face][i] = "?"
			}
		}
	}

	report := cube.ValidatePartial(faces)
	if !report.Legal {
		t.Fatalf("centers-only scan rejected: %#v", report.Violation)
	}
	if report.Ambiguous == nil || !*report.Ambiguous {
		t.Errorf("ambiguous = %v, want true", report.Ambiguous)
	}
	if report.Completion == nil {
		t.Fatal("centers-only report lacks a completion")
	}
	assertAgreesWithObservations(t, faces, report.Completion)
}

// A single unreadable sticker in a fully determined state has exactly one
// completion, so the report must not cry ambiguity.
func TestPartialSingleUnknownStickerIsUnique(t *testing.T) {
	faces := solved()
	faces["U"][8] = "?"

	report := cube.ValidatePartial(faces)
	if !report.Legal {
		t.Fatalf("solvable state with one unknown sticker rejected: %#v", report.Violation)
	}
	if report.Ambiguous == nil || *report.Ambiguous {
		t.Errorf("ambiguous = %v, want false", report.Ambiguous)
	}
	assertAgreesWithObservations(t, faces, report.Completion)
	if got := report.Completion["U"][8]; got != "W" {
		t.Errorf("completion U[8] = %q, want W", got)
	}
}

// An odd corner transposition with no matching edge freedom is genuinely
// unreachable; hiding a sticker must not manufacture a completion.
func TestPartialOddCornerSwapHasNoCompletion(t *testing.T) {
	faces := solved()
	swapCornerPieces(faces, 0, 1) // URF <-> UFL, odd corner permutation
	// Hide one sticker of an unrelated, fully determined corner.
	mask([]cube.StickerRef{cube.CornerRefs[2].Stickers[0]}, faces)

	report := cube.ValidatePartial(faces)
	if report.Legal {
		t.Fatalf("odd corner transposition with no edge freedom was completed: %#v", report.Completion)
	}
	if report.Violation == nil || report.Violation.Code != "no_completion" {
		t.Fatalf("violation = %#v, want no_completion", report.Violation)
	}
}

// Unknown stickers cannot repair an impossible piece combination even when
// several pieces are unreadable.
func TestPartialImpossibleWithManyUnknowns(t *testing.T) {
	faces := solved()
	// Exchange a side sticker of UF and DR. This keeps every color at nine
	// but turns UF into a second UR and DR into a second DF: the observed
	// piece set cannot be a bijection even with an unrelated edge hidden.
	faces["F"][1], faces["R"][7] = faces["R"][7], faces["F"][1]
	mask(append([]cube.StickerRef(nil), cube.EdgeRefs[9].Stickers[:]...), faces)

	report := cube.ValidatePartial(faces)
	if report.Legal {
		t.Fatalf("state with a duplicated corner piece was completed: %#v", report.Completion)
	}
	if report.Violation == nil || report.Violation.Code != "no_completion" {
		t.Fatalf("violation = %#v, want no_completion", report.Violation)
	}
}

func TestPartialRejectsUnknownColor(t *testing.T) {
	faces := solved()
	faces["F"][0] = "Z"
	report := cube.ValidatePartial(faces)
	if report.Legal || report.Violation.Code != "colors" {
		t.Fatalf("got legal=%v violation=%#v, want colors", report.Legal, report.Violation)
	}
}
