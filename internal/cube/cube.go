// Package cube authenticates the sticker state of a 3x3x3 Rubik's cube.
package cube

import "sort"

// Face is a fixed face identifier. Identifiers also identify center colors
// after mapping a concrete input color to the face on which it is centered.
type Face int

const (
	U Face = iota
	R
	F
	D
	L
	B
)

// StickerRef identifies one sticker position on an input face.
type StickerRef struct {
	Face  Face `json:"face"`
	Index int  `json:"index"`
}

// CornerRef contains the three sticker slots for one corner position.
type CornerRef struct {
	Name     string
	Stickers [3]StickerRef
}

// EdgeRef contains the two sticker slots for one edge position.
type EdgeRef struct {
	Name     string
	Stickers [2]StickerRef
}

// Corner positions, in the documented fixed order.
var CornerRefs = [8]CornerRef{
	{Name: "URF", Stickers: [3]StickerRef{{U, 8}, {R, 0}, {F, 2}}},
	{Name: "UFL", Stickers: [3]StickerRef{{U, 6}, {F, 0}, {L, 2}}},
	{Name: "ULB", Stickers: [3]StickerRef{{U, 0}, {L, 0}, {B, 2}}},
	{Name: "UBR", Stickers: [3]StickerRef{{U, 2}, {B, 0}, {R, 2}}},
	{Name: "DFR", Stickers: [3]StickerRef{{D, 2}, {F, 8}, {R, 6}}},
	{Name: "DLF", Stickers: [3]StickerRef{{D, 0}, {L, 8}, {F, 6}}},
	{Name: "DBL", Stickers: [3]StickerRef{{D, 6}, {B, 8}, {L, 6}}},
	{Name: "DRB", Stickers: [3]StickerRef{{D, 8}, {R, 8}, {B, 6}}},
}

// Edge positions, in the documented fixed order.
var EdgeRefs = [12]EdgeRef{
	{Name: "UR", Stickers: [2]StickerRef{{U, 5}, {R, 1}}},
	{Name: "UF", Stickers: [2]StickerRef{{U, 7}, {F, 1}}},
	{Name: "UL", Stickers: [2]StickerRef{{U, 3}, {L, 1}}},
	{Name: "UB", Stickers: [2]StickerRef{{U, 1}, {B, 1}}},
	{Name: "DR", Stickers: [2]StickerRef{{D, 5}, {R, 7}}},
	{Name: "DF", Stickers: [2]StickerRef{{D, 1}, {F, 7}}},
	{Name: "DL", Stickers: [2]StickerRef{{D, 3}, {L, 7}}},
	{Name: "DB", Stickers: [2]StickerRef{{D, 7}, {B, 7}}},
	{Name: "FR", Stickers: [2]StickerRef{{F, 5}, {R, 3}}},
	{Name: "FL", Stickers: [2]StickerRef{{F, 3}, {L, 5}}},
	{Name: "BL", Stickers: [2]StickerRef{{B, 5}, {L, 3}}},
	{Name: "BR", Stickers: [2]StickerRef{{B, 3}, {R, 5}}},
}

// Faces is one array of nine sticker colors for each face.
type Faces map[string][]string

// ColorIssue identifies a center or palette/count failure.
type ColorIssue struct {
	Kind     string `json:"kind"`
	Face     string `json:"face,omitempty"`
	Color    string `json:"color,omitempty"`
	Count    int    `json:"count,omitempty"`
	Expected int    `json:"expected,omitempty"`
}

// Sticker describes one concrete sticker involved in a malformed piece.
type Sticker struct {
	Face  string `json:"face"`
	Index int    `json:"index"`
	Color string `json:"color"`
}

// PieceIssue identifies duplicate, missing, or invalid cubies.
type PieceIssue struct {
	Kind         string    `json:"kind"`
	Position     *int      `json:"position,omitempty"`
	PositionName string    `json:"position_name,omitempty"`
	Piece        string    `json:"piece,omitempty"`
	Stickers     []Sticker `json:"stickers,omitempty"`
}

// ParityIssue reports the corner and edge permutation cycles whose parities differ.
type ParityIssue struct {
	CornerParity int        `json:"corner_parity"`
	EdgeParity   int        `json:"edge_parity"`
	CornerCycles [][]string `json:"corner_cycles,omitempty"`
	EdgeCycles   [][]string `json:"edge_cycles,omitempty"`
	Pieces       []string   `json:"pieces"`
}

// Violation is the first failed invariant.
type Violation struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Colors  []ColorIssue `json:"colors,omitempty"`
	Pieces  []PieceIssue `json:"pieces,omitempty"`
	Parity  *ParityIssue `json:"parity,omitempty"`
}

// CubieCube is the standardized legal-cube representation.
type CubieCube struct {
	CornerPositions    [8]int  `json:"corner_positions"`
	CornerOrientations [8]int  `json:"corner_orientations"`
	EdgePositions      [12]int `json:"edge_positions"`
	EdgeOrientations   [12]int `json:"edge_orientations"`
}

// Report is returned for every semantically parsed cube state.
type Report struct {
	Legal      bool       `json:"legal"`
	Violation  *Violation `json:"violation,omitempty"`
	Cube       *CubieCube `json:"cube,omitempty"`
	Completion Faces      `json:"completion,omitempty"`
	Ambiguous  *bool      `json:"ambiguous,omitempty"`
}

type cubie struct {
	id    int
	faces []Face
	valid bool
}

// Validate checks colors, piece bijectivity, orientations, and permutation
// parity in that order. It returns the first violated invariant.
func Validate(in Faces) *Report {
	faceOrder := []Face{U, R, F, D, L, B}
	faceNames := map[Face]string{U: "U", R: "R", F: "F", D: "D", L: "L", B: "B"}

	centerColor := map[Face]string{}
	colorFace := map[string]Face{}
	var colorIssues []ColorIssue
	for _, f := range faceOrder {
		color := in[faceNames[f]][4]
		centerColor[f] = color
		if other, ok := colorFace[color]; ok {
			colorIssues = append(colorIssues,
				ColorIssue{Kind: "duplicate_center", Face: faceNames[other], Color: color},
				ColorIssue{Kind: "duplicate_center", Face: faceNames[f], Color: color},
			)
			continue
		}
		colorFace[color] = f
	}

	counts := map[string]int{}
	for _, f := range faceOrder {
		for _, color := range in[faceNames[f]] {
			counts[color]++
		}
	}
	allColors := make([]string, 0, len(counts))
	for color := range counts {
		allColors = append(allColors, color)
	}
	sort.Strings(allColors)
	for _, color := range allColors {
		if _, ok := colorFace[color]; !ok {
			colorIssues = append(colorIssues, ColorIssue{Kind: "unknown_color", Color: color, Count: counts[color]})
		}
	}
	// A color can be reported as both duplicate center and wrong count; the
	// first report is enough, but keep counts stable by iterating face order.
	for _, f := range faceOrder {
		color := centerColor[f]
		if counts[color] != 9 {
			colorIssues = append(colorIssues, ColorIssue{
				Kind: "wrong_count", Face: faceNames[f], Color: color,
				Count: counts[color], Expected: 9,
			})
		}
	}
	if len(colorIssues) > 0 {
		return violation("colors", "each face center must be unique and its color must occur exactly nine times", colorIssues, nil, nil)
	}

	corners := make([]cubie, 8)
	for i, ref := range CornerRefs {
		faces := physicalFaces(ref.Stickers[:], in, colorFace)
		id, ok := findCubie(faces, true)
		corners[i] = cubie{id: id, faces: faces, valid: ok}
	}

	edges := make([]cubie, 12)
	for i, ref := range EdgeRefs {
		faces := physicalFaces(ref.Stickers[:], in, colorFace)
		id, ok := findCubie(faces, false)
		edges[i] = cubie{id: id, faces: faces, valid: ok}
	}

	cornerSeen := map[int]int{}
	edgeSeen := map[int]int{}
	for _, c := range corners {
		if c.valid {
			cornerSeen[c.id]++
		}
	}
	for _, e := range edges {
		if e.valid {
			edgeSeen[e.id]++
		}
	}

	var pieceIssues []PieceIssue
	for i, c := range corners {
		stickers := issueStickers(CornerRefs[i].Stickers[:], in)
		if !c.valid {
			pieceIssues = append(pieceIssues, PieceIssue{
				Kind: "invalid_corner", Position: intPtr(i),
				PositionName: CornerRefs[i].Name, Stickers: stickers,
			})
		} else if cornerSeen[c.id] > 1 {
			pieceIssues = append(pieceIssues, PieceIssue{
				Kind: "duplicate_corner", Position: intPtr(i),
				PositionName: CornerRefs[i].Name, Piece: CornerRefs[c.id].Name,
				Stickers: stickers,
			})
		}
	}
	for id := 0; id < 8; id++ {
		if cornerSeen[id] == 0 {
			pieceIssues = append(pieceIssues, PieceIssue{
				Kind: "missing_corner", Piece: CornerRefs[id].Name,
			})
		}
	}
	for i, e := range edges {
		stickers := issueStickers(EdgeRefs[i].Stickers[:], in)
		if !e.valid {
			pieceIssues = append(pieceIssues, PieceIssue{
				Kind: "invalid_edge", Position: intPtr(i),
				PositionName: EdgeRefs[i].Name, Stickers: stickers,
			})
		} else if edgeSeen[e.id] > 1 {
			pieceIssues = append(pieceIssues, PieceIssue{
				Kind: "duplicate_edge", Position: intPtr(i),
				PositionName: EdgeRefs[i].Name, Piece: EdgeRefs[e.id].Name,
				Stickers: stickers,
			})
		}
	}
	for id := 0; id < 12; id++ {
		if edgeSeen[id] == 0 {
			pieceIssues = append(pieceIssues, PieceIssue{
				Kind: "missing_edge", Piece: EdgeRefs[id].Name,
			})
		}
	}
	if len(pieceIssues) > 0 {
		return violation("piece_count", "each corner and edge piece must occur exactly once", nil, pieceIssues, nil)
	}

	var cornerOrientationSum int
	var badCorners []PieceIssue
	for pos, c := range corners {
		co := indexOfAxis(c.faces, 0)
		cornerOrientationSum += co
		if co < 0 || co > 2 {
			panic("validated corner has no U/D color")
		}
		if co != 0 {
			badCorners = append(badCorners, PieceIssue{
				Kind: "twisted_corner", Position: intPtr(pos),
				PositionName: CornerRefs[pos].Name, Piece: CornerRefs[c.id].Name,
				Stickers: issueStickers(CornerRefs[pos].Stickers[:], in),
			})
		}
	}

	var edgeOrientationSum int
	var badEdges []PieceIssue
	for pos, e := range edges {
		canonicalFirst := EdgeRefs[e.id].Stickers[0].Face
		eo := 0
		if e.faces[0] != canonicalFirst {
			eo = 1
		}
		edgeOrientationSum += eo
		if eo != 0 {
			badEdges = append(badEdges, PieceIssue{
				Kind: "flipped_edge", Position: intPtr(pos),
				PositionName: EdgeRefs[pos].Name, Piece: EdgeRefs[e.id].Name,
				Stickers: issueStickers(EdgeRefs[pos].Stickers[:], in),
			})
		}
	}
	if cornerOrientationSum%3 != 0 {
		return violation("corner_orientation", "corner orientations must sum to zero modulo three", nil, badCorners, nil)
	}
	if edgeOrientationSum%2 != 0 {
		return violation("edge_orientation", "edge orientations must sum to zero modulo two", nil, badEdges, nil)
	}

	cp := [8]int{}
	co := [8]int{}
	for i, c := range corners {
		cp[i] = c.id
		co[i] = indexOfAxis(c.faces, 0)
	}
	ep := [12]int{}
	eo := [12]int{}
	for i, e := range edges {
		ep[i] = e.id
		if e.faces[0] != EdgeRefs[e.id].Stickers[0].Face {
			eo[i] = 1
		}
	}
	cornerParity, cornerCycles := permutationParity(cp[:])
	edgeParity, edgeCycles := permutationParity(ep[:])
	if cornerParity != edgeParity {
		return violation("permutation_parity", "corner and edge permutations must have equal parity", nil, nil, &ParityIssue{
			CornerParity: cornerParity,
			EdgeParity:   edgeParity,
			CornerCycles: nameCycles(cornerCycles, CornerNames()),
			EdgeCycles:   nameCycles(edgeCycles, EdgeNames()),
			Pieces:       append(cyclePieceNames(cornerCycles, CornerNames()), cyclePieceNames(edgeCycles, EdgeNames())...),
		})
	}

	return &Report{Legal: true, Cube: &CubieCube{
		CornerPositions:    cp,
		CornerOrientations: co,
		EdgePositions:      ep,
		EdgeOrientations:   eo,
	}}
}

func violation(code, message string, colors []ColorIssue, pieces []PieceIssue, parity *ParityIssue) *Report {
	return &Report{
		Legal: false,
		Violation: &Violation{
			Code:    code,
			Message: message,
			Colors:  colors,
			Pieces:  pieces,
			Parity:  parity,
		},
	}
}

func physicalFaces(refs []StickerRef, in Faces, colorFace map[string]Face) []Face {
	names := []string{"U", "R", "F", "D", "L", "B"}
	faces := make([]Face, len(refs))
	for i, ref := range refs {
		faces[i] = colorFace[in[names[ref.Face]][ref.Index]]
	}
	return faces
}

func issueStickers(refs []StickerRef, in Faces) []Sticker {
	names := []string{"U", "R", "F", "D", "L", "B"}
	out := make([]Sticker, len(refs))
	for i, ref := range refs {
		out[i] = Sticker{Face: names[ref.Face], Index: ref.Index, Color: in[names[ref.Face]][ref.Index]}
	}
	return out
}

func findCubie(faces []Face, corner bool) (int, bool) {
	if !distinctAxes(faces) {
		return -1, false
	}
	key := keyFor(faces)
	if corner {
		for id, ref := range CornerRefs {
			if keyFor(ref.faceIDs()) == key {
				return id, true
			}
		}
		return -1, false
	}
	for id, ref := range EdgeRefs {
		if keyFor(ref.faceIDs()) == key {
			return id, true
		}
	}
	return -1, false
}

func (r CornerRef) faceIDs() []Face {
	out := make([]Face, 3)
	for i, s := range r.Stickers {
		out[i] = s.Face
	}
	return out
}

func (r EdgeRef) faceIDs() []Face {
	out := make([]Face, 2)
	for i, s := range r.Stickers {
		out[i] = s.Face
	}
	return out
}

func distinctAxes(faces []Face) bool {
	seen := map[int]bool{}
	for _, f := range faces {
		axis := int(f) % 3
		if seen[axis] {
			return false
		}
		seen[axis] = true
	}
	return len(faces) == len(seen)
}

func keyFor(faces []Face) string {
	ids := make([]int, len(faces))
	for i, f := range faces {
		ids[i] = int(f)
	}
	sort.Ints(ids)
	out := ""
	for _, id := range ids {
		if out != "" {
			out += ","
		}
		out += string(rune('0' + id))
	}
	return out
}

func indexOfAxis(faces []Face, axis int) int {
	for i, f := range faces {
		if int(f)%3 == axis {
			return i
		}
	}
	return -1
}

func permutationParity(p []int) (int, [][]int) {
	seen := make([]bool, len(p))
	var cycles [][]int
	parity := 0
	for start := 0; start < len(p); start++ {
		if seen[start] {
			continue
		}
		var cycle []int
		at := start
		for !seen[at] {
			seen[at] = true
			cycle = append(cycle, at)
			at = p[at]
		}
		if len(cycle) > 1 {
			cycles = append(cycles, cycle)
			parity ^= (len(cycle) - 1) % 2
		}
	}
	return parity, cycles
}

func nameCycles(cycles [][]int, names []string) [][]string {
	out := make([][]string, len(cycles))
	for i, cycle := range cycles {
		out[i] = make([]string, len(cycle))
		for j, pos := range cycle {
			out[i][j] = names[pos]
		}
	}
	return out
}

func cyclePieceNames(cycles [][]int, names []string) []string {
	var out []string
	for _, cycle := range cycles {
		for _, pos := range cycle {
			out = append(out, names[pos])
		}
	}
	return out
}

func CornerNames() []string {
	out := make([]string, len(CornerRefs))
	for i, r := range CornerRefs {
		out[i] = r.Name
	}
	return out
}

func EdgeNames() []string {
	out := make([]string, len(EdgeRefs))
	for i, r := range EdgeRefs {
		out[i] = r.Name
	}
	return out
}

func intPtr(v int) *int { return &v }
