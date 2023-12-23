package renderer

import (
	"EOB1MazeViewer/formats"
)

type DecorationPosition struct {
	XFlip  int
	Wall   int
	XDelta int
}

/*
	 wall positions:

		9 7 3 7 9
		8 6 2 6 8
		8 5 1 5 8
		  4 0 4
		    ^=party pos.
*/
var DecorationPositions = map[int]DecorationPosition{
	A_EAST: {XFlip: 0, Wall: -1, XDelta: 0},
	B_EAST: {XFlip: 0, Wall: 9, XDelta: 0},
	C_EAST: {XFlip: 0, Wall: 7, XDelta: 0},

	E_WEST: {XFlip: 1, Wall: 7, XDelta: 0},
	F_WEST: {XFlip: 1, Wall: 9, XDelta: 0},
	G_WEST: {XFlip: 0, Wall: -1, XDelta: 0},

	B_SOUTH: {XFlip: 0, Wall: 3, XDelta: -12},
	C_SOUTH: {XFlip: 0, Wall: 3, XDelta: -6},
	D_SOUTH: {XFlip: 0, Wall: 3, XDelta: 0}, // middle front wall
	E_SOUTH: {XFlip: 0, Wall: 3, XDelta: 6},
	F_SOUTH: {XFlip: 0, Wall: 3, XDelta: 12},

	H_EAST: {XFlip: 0, Wall: 8, XDelta: 0},
	I_EAST: {XFlip: 0, Wall: 6, XDelta: 0},

	K_WEST: {XFlip: 1, Wall: 6, XDelta: 0},
	L_WEST: {XFlip: 1, Wall: 8, XDelta: 0},

	I_SOUTH: {XFlip: 0, Wall: 2, XDelta: -10},
	J_SOUTH: {XFlip: 0, Wall: 2, XDelta: 0}, // middle front wall
	K_SOUTH: {XFlip: 0, Wall: 2, XDelta: 10},

	M_EAST: {XFlip: 0, Wall: 5, XDelta: 0},

	O_WEST: {XFlip: 1, Wall: 5, XDelta: 0},

	M_SOUTH: {XFlip: 0, Wall: -1, XDelta: -16},
	N_SOUTH: {XFlip: 0, Wall: 1, XDelta: 0}, // middle front wall
	O_SOUTH: {XFlip: 0, Wall: -1, XDelta: 16},

	P_EAST: {XFlip: 0, Wall: 4, XDelta: 0},

	Q_WEST: {XFlip: 1, Wall: 4, XDelta: 0},
}

type DecorationRenderer struct {
	decorationContainer *DecorationContainer
}

func NewDecorationRenderer(decorationContainer *DecorationContainer) *DecorationRenderer {
	return &DecorationRenderer{decorationContainer: decorationContainer}
}

func (dr *DecorationRenderer) DrawCompleteDecoration(background *[]byte, wallMapping formats.WallMapping, renderPosition int, isAtWall bool) *[]byte {
	if wallMapping.DecorationId == 0xFF {
		return background
	}

	bitmap := dr.decorationContainer.GetDecorationBitmapByName(wallMapping.CpsName)
	decoration := dr.decorationContainer.GetDecoration(wallMapping.DecorationId)

	dr.drawDecoration(background, decoration, renderPosition, isAtWall, bitmap)
	for decoration.LinkToNextDecoration != 0 {
		decoration = dr.decorationContainer.GetDecoration(int(decoration.LinkToNextDecoration))
		dr.drawDecoration(background, decoration, renderPosition, isAtWall, bitmap)
	}
	return background
}

func (dr *DecorationRenderer) drawDecoration(background *[]byte, decoration formats.Decoration, renderPosition int, isAtWall bool, decorationBitmap *[]byte) {
	var i, j, s, t, dx, pos int
	var mirrored bool
	var q byte

	dx = 0
	if isAtWall {
		dx = 8 * DecorationPositions[renderPosition].XDelta
	} else {
		switch renderPosition {
		case 6:
			dx = -88
		case 7:
			dx = -40
		case 8:
			dx = 88
		case 9:
			dx = 40
		case 15:
			dx = -59
		case 16:
			dx = 59
		case 20:
			dx = -98
		case 21:
			dx = 98
		}
	}

	pos = DecorationPositions[renderPosition].Wall
	if pos >= 0 {
		q = decoration.RectangleIndices[pos]

		if q != 0xFF {
			mirrored = false
			switch renderPosition {
			case 6, 7, 8, 9, 10, 15, 16, 17, 20, 21, 22, 25:
				mirrored = decoration.Flags&0x01 != 0
			}

			t = int(decoration.YCoords[pos])

			rectangle := dr.decorationContainer.GetCPSRectangle(q)
			for j = int(rectangle.Y); j < int(rectangle.Y+rectangle.H); j++ {
				if mirrored {
					s = int(22*8 - decoration.XCoords[pos] - 1)
				} else {
					s = int(decoration.XCoords[pos])
				}

				for i = int(rectangle.X * 8); i < int(rectangle.X*8+rectangle.W*8); i++ {
					if mirrored {
						putPixel(background, s+dx, t, (*decorationBitmap)[320*j+i])
						s--
					} else {
						if DecorationPositions[renderPosition].XFlip == 0 {
							putPixel(background, s+dx, t, (*decorationBitmap)[320*j+i])
						} else {
							putPixel(background, 22*8-(s+dx), t, (*decorationBitmap)[320*j+i])
						}
						s++
					}
				}
				t++
			}
		}
	}
}

func (dr *DecorationRenderer) GetDecorationBitmapByName(name string) interface{} {
	return dr.decorationContainer.GetDecorationBitmapByName(name)
}

func putPixel(background *[]byte, x int, y int, b byte) {
	if b == 0x00 {
		return
	}
	coord := 176*y + x
	if coord >= 0 && coord < len(*background) {
		(*background)[coord] = b
	}
}
