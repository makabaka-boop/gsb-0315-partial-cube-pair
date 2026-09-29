# Cube authenticator

A small Go JSON command-line authenticator for a 3x3x3 Rubik's cube. It rejects states that merely have nine stickers of every color but cannot be produced by legal face turns.

The `cube` service in `docker-compose.yml` runs the authenticator. It reads one JSON object from standard input and writes one JSON report to standard output.

## Input

Provide all six faces, each as nine sticker colors in row-major order while looking directly at that face:

- `F/R/B/L` top edges point toward `U`.
- `U`'s top edge points toward `B`.
- `D`'s top edge points toward `F`.

Facelet, corner, and edge indices are documented in [`docs/cube.md`](docs/cube.md).

The six centers must be pairwise different. Their colors are not required to be a particular palette or to match a standard color scheme. Each center color identifies the physical face to which that center belongs and must occur exactly nine times.

The scanner may also send `"?"` for an unreadable non-center sticker. In that case the report is legal if at least one fully colored, reachable cube agrees with every observed sticker and the six observed centers. A legal partial report includes `completion`, a six-face sticker grid with no `"?"`, alongside the usual cubie representation. If no such cube exists, the report has `legal: false` and violation code `no_completion`; unknown colors or ambiguous centers retain the `colors` violation. Fully observed inputs keep their ordinary validation and report format.

Legal partial reports also include `ambiguous`: `true` when another different reachable full sticker grid agrees with the same observations, `false` when the reported completion is the only one. This lets the scanner distinguish a definitive reading from one that still needs another sticker photographed.

Example:

```bash
docker compose run --rm -T cube <<'JSON'
{
  "U": ["W","W","W","W","W","W","W","W","W"],
  "R": ["G","G","G","G","G","G","G","G","G"],
  "F": ["O","O","O","O","O","O","O","O","O"],
  "D": ["Y","Y","Y","Y","Y","Y","Y","Y","Y"],
  "L": ["B","B","B","B","B","B","B","B","B"],
  "B": ["X","X","X","X","X","X","X","X","X"]
}
JSON
```

## Validation order

The authenticator reports the first failed invariant:

1. Center colors are distinct, all colors are known center colors, and each color occurs exactly nine times.
2. Every sticker triple/pair forms a real corner/edge, and each of the eight corner and twelve edge pieces occurs exactly once.
3. Corner orientation is `0,1,2` with sum `0 mod 3`.
4. Edge orientation is `0,1` with sum `0 mod 2`.
5. Corner and edge permutation parities are equal.

For failures, `violation.code` is one of `colors`, `piece_count`, `corner_orientation`, `edge_orientation`, or `permutation_parity`, and the report includes the involved cubie(s) and stickers.

A legal report includes a standardized cubie representation:

- `corner_positions[position]` identifies the corner piece at that position.
- `corner_orientations[position]` is its modulo-three orientation.
- `edge_positions[position]` identifies the edge piece.
- `edge_orientations[position]` is its modulo-two orientation.

Indices and names use the tables in [`docs/cube.md`](docs/cube.md).

## Exit codes

- `0`: input was parsed and the cube state is legal.
- `1`: input was parsed, but the state violates a cube invariant.
- `2`: malformed request or input/output failure.

## Local development

```bash
go test ./...
go run ./cmd/authenticator < state.json
```

The Go tests do not rely on randomized sticker layouts. They build states from a solved cube using an independent three-dimensional face-turn model, and also construct the important isolated failure cases: one twisted corner, one flipped edge, a single transposition of two edges, and duplicated/missing pieces.
