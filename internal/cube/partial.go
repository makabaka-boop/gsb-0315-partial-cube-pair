package cube

// Partial facelets use "?" for a sticker whose color was not observed.
// Centers always identify the six physical colors.

var faceLabels = [6]string{"U", "R", "F", "D", "L", "B"}

// candidate places one concrete piece into one position with one orientation,
// consistent with every sticker observed at that position.
type candidate struct {
	piece  int
	orient int
}

// assignment is one bijection of pieces over all positions, plus the
// orientation of the piece at each position.
type assignment struct {
	pieces  []int
	orients []int
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
	for _, face := range faceLabels {
		for _, color := range in[face] {
			if color == "?" {
				continue
			}
			if !colors[color] {
				return violation("colors", "an observed sticker has no matching center", nil, nil, nil)
			}
		}
	}
	counts := make(map[string]int)
	for _, face := range faceLabels {
		for _, color := range in[face] {
			if color != "?" {
				counts[color]++
				if counts[color] > 9 {
					return violation("colors", "an observed color occurs more than nine times", nil, nil, nil)
				}
			}
		}
	}

	cornerCandidates := make([][]candidate, 8)
	for pos := range CornerRefs {
		for piece := range CornerRefs {
			for twist := 0; twist < 3; twist++ {
				if matchesCorner(in, pos, piece, twist) {
					cornerCandidates[pos] = append(cornerCandidates[pos], candidate{piece, twist})
				}
			}
		}
	}
	edgeCandidates := make([][]candidate, 12)
	for pos := range EdgeRefs {
		for piece := range EdgeRefs {
			for flip := 0; flip < 2; flip++ {
				if matchesEdge(in, pos, piece, flip) {
					edgeCandidates[pos] = append(edgeCandidates[pos], candidate{piece, flip})
				}
			}
		}
	}

	// Enumerate complete, observation-consistent bijections, bucketed by the
	// permutation parity that must agree between corners and edges. Two
	// solutions per bucket are enough: together they decide whether the whole
	// partial state has zero, one, or several completions.
	cornerSols := enumerateAssignments(8, 3, cornerCandidates)
	edgeSols := enumerateAssignments(12, 2, edgeCandidates)

	var completed Faces
	solutionCount := 0
	for parity := 0; parity < 2; parity++ {
		bucketCount := len(cornerSols[parity]) * len(edgeSols[parity])
		if bucketCount == 0 {
			continue
		}
		solutionCount += bucketCount
		if completed == nil {
			cp := cornerSols[parity][0]
			ep := edgeSols[parity][0]
			completed = make(Faces)
			for _, face := range faceLabels {
				completed[face] = append([]string(nil), in[face]...)
			}
			for pos, ref := range CornerRefs {
				writeCorner(completed, ref, cp.pieces[pos], cp.orients[pos], in)
			}
			for pos, ref := range EdgeRefs {
				writeEdge(completed, ref, ep.pieces[pos], ep.orients[pos], in)
			}
		}
	}
	if completed == nil {
		return noCompletion()
	}

	report := Validate(completed)
	if !report.Legal {
		return noCompletion()
	}
	report.Completion = completed
	ambiguous := solutionCount >= 2
	report.Ambiguous = &ambiguous
	return report
}

// enumerateAssignments returns up to two complete assignments of every
// position for each permutation parity (0 for even, 1 for odd). Orientations
// are required to sum to zero modulo orientMod.
func enumerateAssignments(n, orientMod int, candidates [][]candidate) map[int][]assignment {
	solutions := map[int][]assignment{}
	pieces := make([]int, n)
	orients := make([]int, n)
	var usedPieces, filledPositions uint32

	var dfs func(orientSum int)
	dfs = func(orientSum int) {
		if len(solutions[0]) == 2 && len(solutions[1]) == 2 {
			return
		}
		if filledPositions == (uint32(1)<<n)-1 {
			if orientSum%orientMod != 0 {
				return
			}
			parity := permutationParityOf(pieces)
			if len(solutions[parity]) < 2 {
				solutions[parity] = append(solutions[parity], assignment{
					pieces:  append([]int(nil), pieces...),
					orients: append([]int(nil), orients...),
				})
			}
			return
		}

		// Choose the unfilled position with the fewest still-available pieces.
		pos := -1
		best := 0
		for p := 0; p < n; p++ {
			if filledPositions&(1<<p) != 0 {
				continue
			}
			count := 0
			for _, cand := range candidates[p] {
				if usedPieces&(1<<cand.piece) == 0 {
					count++
				}
			}
			if count == 0 {
				return
			}
			if pos == -1 || count < best {
				pos, best = p, count
			}
		}

		// Drop this branch if the remaining slots cannot close the orientation
		// sum. This ignores piece collisions, so it is an over-approximation.
		if !orientationSumFeasible(candidates, filledPositions, n, orientSum, orientMod) {
			return
		}

		for _, cand := range candidates[pos] {
			if usedPieces&(1<<cand.piece) != 0 {
				continue
			}
			pieces[pos], orients[pos] = cand.piece, cand.orient
			usedPieces |= 1 << cand.piece
			filledPositions |= 1 << pos
			dfs(orientSum + cand.orient)
			usedPieces &^= 1 << cand.piece
			filledPositions &^= 1 << pos
			if len(solutions[0]) == 2 && len(solutions[1]) == 2 {
				return
			}
		}
	}

	dfs(0)
	return solutions
}

// orientationSumFeasible reports whether every position still to be filled
// (including the position the caller is about to iterate) can contribute the
// orientation residue still needed. Piece collisions are ignored, so the
// result is an over-approximation and only prunes impossible branches.
func orientationSumFeasible(candidates [][]candidate, filled uint32, n, sum, mod int) bool {
	reachable := []bool{true}
	for p := 0; p < n; p++ {
		if filled&(1<<p) != 0 {
			continue
		}
		next := make([]bool, mod)
		for _, cand := range candidates[p] {
			for residue, ok := range reachable {
				if ok {
					next[(residue+cand.orient)%mod] = true
				}
			}
		}
		reachable = next
	}
	need := (-sum%mod + mod) % mod
	return reachable[need]
}

func permutationParityOf(p []int) int {
	seen := make([]bool, len(p))
	parity := 0
	for start := 0; start < len(p); start++ {
		if seen[start] {
			continue
		}
		length := 0
		at := start
		for !seen[at] {
			seen[at] = true
			length++
			at = p[at]
		}
		parity ^= (length - 1) % 2
	}
	return parity
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
