# Cube facelet and cubie indexing

The API receives the six faces `U`, `R`, `F`, `D`, `L`, and `B` as arrays of nine strings in row-major order, as seen when looking directly at that face.

## Face orientation

When unfolding the cube into the common net, the given edge of each face points toward:

- `F/R/B/L`: its upper edge points toward `U`.
- `U`: its upper edge points toward `B`.
- `D`: its upper edge points toward `F`.

Each face's nine entries are therefore numbered as follows:

```text
0 1 2
3 4 5
6 7 8
```

The side faces are aligned vertically with each other. `U` is viewed from above with `B` at the top, and `D` is viewed from below with `F` at the top.

## Coordinate model

Coordinates use one sticker axis value of `-1`, `0`, or `+1`; center cubies have two zero values. The positive directions are:

| Axis | + direction | - direction |
| --- | --- | --- |
| x | R | L |
| y | U | D |
| z | F | B |

A clockwise face move is a +90° rotation about that face's outward normal. Thus, for negative-direction faces (`L`, `D`, `B`), this is the opposite sign of rotation about the positive coordinate axis.

## Corner position indices

The position order is fixed, regardless of the color labels supplied by the caller. The three stickers of each position are listed in the orientation convention used by the authenticator.

| Index | Position | Stickers (`face:facelet-index`) |
| ---: | --- | --- |
| 0 | URF | `U:8`, `R:0`, `F:2` |
| 1 | UFL | `U:6`, `F:0`, `L:2` |
| 2 | ULB | `U:0`, `L:0`, `B:2` |
| 3 | UBR | `U:2`, `B:0`, `R:2` |
| 4 | DFR | `D:2`, `F:8`, `R:6` |
| 5 | DLF | `D:0`, `L:8`, `F:6` |
| 6 | DBL | `D:6`, `B:8`, `L:6` |
| 7 | DRB | `D:8`, `R:8`, `B:6` |

Corner orientation is the standard modulo-three convention. A corner's three stickers are read in the position-specific order listed in the corner table above, and the orientation equals the index in that list occupied by the cubie's physical U/D color:

- `0`: the cubie's U/D color is on the position's U/D sticker.
- `1`: the U/D color is on the first side sticker listed for the position.
- `2`: it is on the second side sticker.

The sum of corner orientations must be `0 mod 3`.

## Edge position indices

| Index | Position | First sticker | Second sticker |
| ---: | --- | --- | --- |
| 0 | UR | `U:5` | `R:1` |
| 1 | UF | `U:7` | `F:1` |
| 2 | UL | `U:3` | `L:1` |
| 3 | UB | `U:1` | `B:1` |
| 4 | DR | `D:5` | `R:7` |
| 5 | DF | `D:1` | `F:7` |
| 6 | DL | `D:3` | `L:7` |
| 7 | DB | `D:7` | `B:7` |
| 8 | FR | `F:5` | `R:3` |
| 9 | FL | `F:3` | `L:5` |
| 10 | BL | `B:5` | `L:3` |
| 11 | BR | `B:3` | `R:5` |

Each edge piece has a canonical first color: the face named first in the row of its home position. For U/D pieces that is the U/D color; for the middle pieces `FR`, `FL`, `BL`, and `BR`, it is the F/B color. Edge orientation is the standard modulo-two convention:

- `0`: the cubie's canonical first color occupies the first sticker slot listed for its current position.
- `1`: the two stickers are exchanged relative to that reference (the edge is flipped).

The sum of edge orientations must be `0 mod 2`.

## Piece names and permutation validity

A valid corner contains exactly one color from each opposite axis pair: `U/D`, `R/L`, and `F/B`. A valid edge contains colors from two different opposite axis pairs.

Piece names (`URF`, `UR`, etc.) describe the cubie's physical color triple or pair. A legal cube must contain each of the eight corner pieces and twelve edge pieces exactly once. If that permutation is bijective, its corner permutation parity and edge permutation parity must also be equal.
