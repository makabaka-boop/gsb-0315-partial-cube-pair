package cube

// Partial facelets use "?" for a sticker whose color was not observed.
// Centers always identify the six physical colors.

var faceLabels = [6]string{"U", "R", "F", "D", "L", "B"}

type placement struct {
	piece int
	twist int
}

func ValidatePartial(in Faces) *Report {
	unknown := false
	for _, face := range faceLabels {
		for _, color := range in[face] {
			unknown = unknown || color == "?"
		}
	}
	if !unknown {
		return Validate(in)
	}
	colors := make(map[string]bool)
	for _, face := range faceLabels {
		center := in[face][4]
		if center == "?" || colors[center] {
			return violation("colors", "centers must be six distinct known colors", nil, nil, nil)
		}
		colors[center] = true
	}
	counts := make(map[string]int)
	for _, face := range faceLabels {
		for _, color := range in[face] {
			if color == "?" {
				continue
			}
			if !colors[color] {
				return violation("colors", "an observed sticker has no matching center", nil, nil, nil)
			}
			counts[color]++
			if counts[color] > 9 {
				return violation("colors", "an observed color occurs more than nine times", nil, nil, nil)
			}
		}
	}
	completed := make(Faces)
	for _, face := range faceLabels {
		completed[face] = append([]string(nil), in[face]...)
	}
	var usedCorners, usedEdges uint16
	for pos, ref := range CornerRefs {
		found := false
		for piece := range CornerRefs {
			if usedCorners&(1<<piece) != 0 {
				continue
			}
			for twist := 0; twist < 3; twist++ {
				if matchesCorner(in, pos, piece, twist) {
					writeCorner(completed, ref, piece, twist, in)
					usedCorners |= 1 << piece
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return noCompletion()
		}
	}
	for pos, ref := range EdgeRefs {
		found := false
		for piece := range EdgeRefs {
			if usedEdges&(1<<piece) != 0 {
				continue
			}
			for flip := 0; flip < 2; flip++ {
				if matchesEdge(in, pos, piece, flip) {
					writeEdge(completed, ref, piece, flip, in)
					usedEdges |= 1 << piece
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return noCompletion()
		}
	}
	report := Validate(completed)
	if !report.Legal {
		return noCompletion()
	}
	report.Completion = completed
	unique := false
	report.Ambiguous = &unique
	return report
}

func matchesCorner(in Faces, pos, piece, twist int) bool {
	for slot, ref := range CornerRefs[pos].Stickers {
		want := in[faceLabels[ref.Face]][ref.Index]
		colorFace := CornerRefs[piece].Stickers[(slot-twist+3)%3].Face
		if want != "?" && want != in[faceLabels[colorFace]][4] {
			return false
		}
	}
	return true
}

func matchesEdge(in Faces, pos, piece, flip int) bool {
	for slot, ref := range EdgeRefs[pos].Stickers {
		want := in[faceLabels[ref.Face]][ref.Index]
		colorFace := EdgeRefs[piece].Stickers[(slot-flip+2)%2].Face
		if want != "?" && want != in[faceLabels[colorFace]][4] {
			return false
		}
	}
	return true
}

func writeCorner(out Faces, ref CornerRef, piece, twist int, in Faces) {
	for slot, sticker := range ref.Stickers {
		colorFace := CornerRefs[piece].Stickers[(slot-twist+3)%3].Face
		out[faceLabels[sticker.Face]][sticker.Index] = in[faceLabels[colorFace]][4]
	}
}

func writeEdge(out Faces, ref EdgeRef, piece, flip int, in Faces) {
	for slot, sticker := range ref.Stickers {
		colorFace := EdgeRefs[piece].Stickers[(slot-flip+2)%2].Face
		out[faceLabels[sticker.Face]][sticker.Index] = in[faceLabels[colorFace]][4]
	}
}

func noCompletion() *Report {
	return violation("no_completion", "no reachable cube matches the observed stickers", nil, nil, nil)
}
