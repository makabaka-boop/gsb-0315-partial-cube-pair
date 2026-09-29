package cube_test

import "cube-auth/internal/cube"

type vec3 struct {
	x, y, z int
}

type slot3 struct {
	pos    vec3
	normal vec3
	color  string
}

// turn performs one external-clockwise face quarter-turn by rotating the
// three-dimensional stickers in that layer. Tests use this independent model
// instead of only shuffling random stickers.
func turn(in cube.Faces, face byte) cube.Faces {
	n := faceNormal(face)
	axis := nonzeroAxis(n)
	slots := allSlots(in)
	out := make(map[[6]int]slot3, len(slots))
	for _, s := range slots {
		if component(s.pos, axis) != component(n, axis) {
			out[stickerKey(s.pos, s.normal)] = s
			continue
		}
		rotated := slot3{
			pos:    rotateCW(s.pos, n),
			normal: rotateCW(s.normal, n),
			color:  s.color,
		}
		out[stickerKey(rotated.pos, rotated.normal)] = rotated
	}

	faces := cube.Faces{}
	for _, name := range []string{"U", "R", "F", "D", "L", "B"} {
		faces[name] = make([]string, 9)
	}
	for face := 0; face < 6; face++ {
		for row := 0; row < 3; row++ {
			for col := 0; col < 3; col++ {
				f := faceLayout(face)
				p, n := f.slot(row, col)
				faceName := faceLetter(n)
				faces[faceName][row*3+col] = out[stickerKey(p, n)].color
			}
		}
	}
	return faces
}

func allSlots(in cube.Faces) []slot3 {
	var slots []slot3
	for _, face := range []int{0, 1, 2, 3, 4, 5} {
		f := faceLayout(face)
		for row := 0; row < 3; row++ {
			for col := 0; col < 3; col++ {
				p, n := f.slot(row, col)
				slots = append(slots, slot3{pos: p, normal: n, color: in[faceLetter(n)][row*3+col]})
			}
		}
	}
	return slots
}

type faceGeometry struct {
	normal vec3
	up     vec3
	right  vec3
}

func faceLayout(face int) faceGeometry {
	switch face {
	case 0: // U: top points to B.
		return faceGeometry{normal: vec3{0, 1, 0}, up: vec3{0, 0, -1}, right: vec3{1, 0, 0}}
	case 1: // R
		return faceGeometry{normal: vec3{1, 0, 0}, up: vec3{0, 1, 0}, right: vec3{0, 0, -1}}
	case 2: // F
		return faceGeometry{normal: vec3{0, 0, 1}, up: vec3{0, 1, 0}, right: vec3{1, 0, 0}}
	case 3: // D: top points to F.
		return faceGeometry{normal: vec3{0, -1, 0}, up: vec3{0, 0, 1}, right: vec3{1, 0, 0}}
	case 4: // L
		return faceGeometry{normal: vec3{-1, 0, 0}, up: vec3{0, 1, 0}, right: vec3{0, 0, 1}}
	default: // B
		return faceGeometry{normal: vec3{0, 0, -1}, up: vec3{0, 1, 0}, right: vec3{-1, 0, 0}}
	}
}

func (g faceGeometry) slot(row, col int) (vec3, vec3) {
	vertical := []int{1, 0, -1}[row]
	horizontal := []int{-1, 0, 1}[col]
	p := g.normal
	p = add(p, scale(g.up, vertical))
	p = add(p, scale(g.right, horizontal))
	return p, g.normal
}

func faceNormal(face byte) vec3 {
	switch face {
	case 'U':
		return vec3{0, 1, 0}
	case 'R':
		return vec3{1, 0, 0}
	case 'F':
		return vec3{0, 0, 1}
	case 'D':
		return vec3{0, -1, 0}
	case 'L':
		return vec3{-1, 0, 0}
	case 'B':
		return vec3{0, 0, -1}
	default:
		panic("unsupported move")
	}
}

func faceLetter(n vec3) string {
	switch n {
	case vec3{0, 1, 0}:
		return "U"
	case vec3{1, 0, 0}:
		return "R"
	case vec3{0, 0, 1}:
		return "F"
	case vec3{0, -1, 0}:
		return "D"
	case vec3{-1, 0, 0}:
		return "L"
	case vec3{0, 0, -1}:
		return "B"
	default:
		panic("bad face normal")
	}
}

// rotateCW rotates -90 degrees around the outward normal. Looking from the
// outside toward the center, that is clockwise.
func rotateCW(v, n vec3) vec3 {
	switch {
	case n.x != 0:
		// (x,y,z) -> (x, n.x*z, -n.x*y)
		return vec3{x: v.x, y: n.x * v.z, z: -n.x * v.y}
	case n.y != 0:
		// (x,y,z) -> (-n.y*z, y, n.y*x)
		return vec3{x: -n.y * v.z, y: v.y, z: n.y * v.x}
	default:
		// (x,y,z) -> (n.z*y, -n.z*x, z)
		return vec3{x: n.z * v.y, y: -n.z * v.x, z: v.z}
	}
}

func nonzeroAxis(v vec3) string {
	switch {
	case v.x != 0:
		return "x"
	case v.y != 0:
		return "y"
	default:
		return "z"
	}
}

func component(v vec3, axis string) int {
	switch axis {
	case "x":
		return v.x
	case "y":
		return v.y
	default:
		return v.z
	}
}

func add(a, b vec3) vec3 {
	return vec3{x: a.x + b.x, y: a.y + b.y, z: a.z + b.z}
}

func scale(v vec3, s int) vec3 {
	return vec3{x: v.x * s, y: v.y * s, z: v.z * s}
}

func stickerKey(pos, normal vec3) [6]int {
	return [6]int{pos.x, pos.y, pos.z, normal.x, normal.y, normal.z}
}
