package renderer

import (
	"EOB1MazeViewer/formats"
)

// Direction represents the direction with x and y steps.
type Direction struct {
	xs, ys int
}

// MazePosition represents a position in the maze with deltas and a direction.
type MazePosition struct {
	XDelta, YDelta, Direction int
}

// MazeDirection is an array representing maze directions.
var MazeDirection = [4]Direction{
	{xs: 1, ys: 1},   // north
	{xs: -1, ys: 1},  // east
	{xs: -1, ys: -1}, // south
	{xs: 1, ys: -1},  // west
}

// MazePositions is an array representing maze positions.
var MazePositions = [25]MazePosition{
	{XDelta: -3, YDelta: -3, Direction: 1},
	{XDelta: -2, YDelta: -3, Direction: 1},
	{XDelta: -1, YDelta: -3, Direction: 1},
	{XDelta: 1, YDelta: -3, Direction: 3},
	{XDelta: 2, YDelta: -3, Direction: 3},
	{XDelta: 3, YDelta: -3, Direction: 3},

	{XDelta: -2, YDelta: -3, Direction: 2},
	{XDelta: -1, YDelta: -3, Direction: 2},
	{XDelta: 0, YDelta: -3, Direction: 2},
	{XDelta: 1, YDelta: -3, Direction: 2},
	{XDelta: 2, YDelta: -3, Direction: 2},

	{XDelta: -2, YDelta: -2, Direction: 1},
	{XDelta: -1, YDelta: -2, Direction: 1},
	{XDelta: 1, YDelta: -2, Direction: 3},
	{XDelta: 2, YDelta: -2, Direction: 3},

	{XDelta: -1, YDelta: -2, Direction: 2},
	{XDelta: 0, YDelta: -2, Direction: 2}, //
	{XDelta: 1, YDelta: -2, Direction: 2},

	{XDelta: -1, YDelta: -1, Direction: 1},
	{XDelta: 1, YDelta: -1, Direction: 3},

	{XDelta: -1, YDelta: -1, Direction: 2},
	{XDelta: 0, YDelta: -1, Direction: 2}, //
	{XDelta: 1, YDelta: -1, Direction: 2},

	{XDelta: -1, YDelta: 0, Direction: 1},
	{XDelta: 1, YDelta: 0, Direction: 3},
}

type ViewportData *[]byte

type ViewportDataProvider struct {
	inf *formats.InfHeader
	maz *formats.Maz
}

func NewViewportDataProvider(inf *formats.InfHeader, maz *formats.Maz) *ViewportDataProvider {
	return &ViewportDataProvider{inf: inf, maz: maz}
}

func (vdp *ViewportDataProvider) GetViewportData(x, y, direction int) ViewportData {
	viewportData := make([]byte, 25)
	var deltaX, deltaY int
	for i := A_EAST; i <= Q_WEST; i++ {
		if direction%2 != 0 {
			deltaX = MazeDirection[direction].xs * MazePositions[i].YDelta
			deltaY = MazeDirection[direction].ys * MazePositions[i].XDelta
		} else {
			deltaX = MazeDirection[direction].xs * MazePositions[i].XDelta
			deltaY = MazeDirection[direction].ys * MazePositions[i].YDelta
		}

		finalX := x + deltaX
		finalY := y + deltaY

		wallDirection := (direction + MazePositions[i].Direction) & 0x03

		wallMappingIndex := vdp.maz.GetMazeBlockByCoordinateOrFake(finalX, finalY).Wall[wallDirection]
		viewportData[i] = wallMappingIndex
	}

	return &viewportData

}
