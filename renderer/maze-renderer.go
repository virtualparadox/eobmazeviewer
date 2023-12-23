package renderer

import (
	inf2 "EOB1MazeViewer/formats"
	"github.com/elliotchance/orderedmap/v2"
)

// WallRenderData
// baseOffset: Base offset into wallTiles[wallType-1]. This can be negative, but don't be deceived, this is just a base. The other values will add to this base making sure the wallTiles[wallType-1] array is accessed within bounds.\\
// offsetInViewPort: Block index in the viewport where to start render.
//
//	xpos = offsetInViewPort%22;
//	ypos = offsetInViewPort/22;
//
// visibleWidthInBlocks: How many visible blocks wide the wall is in the viewport.
// visibleHeightInBlocks: How many visible blocks hight the wall is in the viewport.
// skipValue: Number of tiles to the next row in the wallTiles[wallType-1] array.
// flipFlag: if the wall is to be x-flipped in the viewport. Generally all right-walls are x-flipped.
type WallRenderData struct {
	offsetInViewPort      int
	visibleHeightInBlocks int
	visibleWidthInBlocks  int
	flipFlag              int
	wallIndex             int
}

const (
	A_EAST = 0
	B_EAST = 1
	C_EAST = 2
	E_WEST = 3
	F_WEST = 4
	G_WEST = 5

	B_SOUTH = 6
	C_SOUTH = 7
	D_SOUTH = 8
	E_SOUTH = 9
	F_SOUTH = 10

	H_EAST = 11
	I_EAST = 12
	K_WEST = 13
	L_WEST = 14

	I_SOUTH = 15
	J_SOUTH = 16
	K_SOUTH = 17

	M_EAST = 18
	O_WEST = 19

	M_SOUTH = 20
	N_SOUTH = 21
	O_SOUTH = 22

	P_EAST = 23
	Q_WEST = 24
)

/*
	 	  A|B|C|D|E|F|G
			¯ ¯ ¯ ¯ ¯
			H|I|J|K|L
			  ¯ ¯ ¯
			  M|N|O
			  ¯ ¯ ¯
			  P|^|Q
*/
var wallRenderData = map[int]WallRenderData{ /* 25 different wall positions exists */
	/* Side-Walls left back */
	A_EAST: {66, 5, 1, 0, 4}, /* A-east 0/3*/
	B_EAST: {68, 5, 3, 0, 4}, /* B-east 2/3*/
	C_EAST: {74, 5, 1, 0, 3}, /* C-east 8/3*/

	/* Side-Walls right back */
	E_WEST: {79, 5, 1, 1, 3}, /* E-west 13/3*/
	F_WEST: {83, 5, 3, 1, 4}, /* F-west 17/3*/
	G_WEST: {87, 5, 1, 1, 4}, /* G-west 21/3*/

	/* Frontwalls back */
	B_SOUTH: {66, 5, 2, 0, 6}, /* B-south */
	C_SOUTH: {68, 5, 6, 0, 6}, /* C-south */
	D_SOUTH: {74, 5, 6, 0, 6}, /* D-south */
	E_SOUTH: {80, 5, 6, 0, 6}, /* E-south */
	F_SOUTH: {86, 5, 2, 0, 6}, /* F-south */

	/* Side walls middle back left */
	H_EAST: {66, 6, 2, 0, 5}, /* H-east */
	I_EAST: {50, 8, 2, 0, 2}, /* I-east */

	/* Side walls middle back right */
	K_WEST: {58, 8, 2, 1, 2}, /* K-west */
	L_WEST: {86, 6, 2, 1, 5}, /* L-west */

	/* Frontwalls middle back */
	I_SOUTH: {44, 8, 6, 0, 7},  /* I-south */
	J_SOUTH: {50, 8, 10, 0, 7}, /* J-south */
	K_SOUTH: {60, 8, 6, 0, 7},  /* K-south */

	/* Side walls middle front left */
	M_EAST: {25, 12, 3, 0, 1}, /* M-east */

	/* Side walls middle front right */
	O_WEST: {38, 12, 3, 1, 1}, /* O-west */

	/* Frontwalls middle front */
	M_SOUTH: {22, 12, 3, 0, 8},  /* M-south */
	N_SOUTH: {25, 12, 16, 0, 8}, /* N-south */
	O_SOUTH: {41, 12, 3, 0, 8},  /* O-south */

	/* Side wall front left */
	P_EAST: {0, 15, 3, 0, 0}, /* P-east */

	/* Side wall front right */
	Q_WEST: {19, 15, 3, 1, 0}, /* Q-west */
}

type MazeRenderer struct {
	viewportDataProvider *ViewportDataProvider
	wallRenderer         *WallRenderer
	decorationRenderer   *DecorationRenderer
	Palette              *inf2.PAL
}

func NewMazeRenderer(inf *inf2.InfHeader, maz *inf2.Maz, vcn *inf2.VCN, vmp *inf2.VMP, pal *inf2.PAL, decorationContainer *DecorationContainer) *MazeRenderer {
	viewportDataProvider := NewViewportDataProvider(inf, maz)
	wallRenderer := NewWallRenderer(vcn, vmp)
	decorationRenderer := NewDecorationRenderer(decorationContainer)

	return &MazeRenderer{
		viewportDataProvider: viewportDataProvider,
		wallRenderer:         wallRenderer,
		decorationRenderer:   decorationRenderer,
		Palette:              pal,
	}
}

func (mr *MazeRenderer) RenderMaze(x int, y int, direction int) (*[]byte, error) {
	viewportData := mr.viewportDataProvider.GetViewportData(x, y, direction)
	background := mr.wallRenderer.RenderBackground()
	if (x+y+direction)%2 == 0 {
		background = flipBackgroundX(background, 176, 120)
	}
	return mr.renderAndOverlay(viewportData, background)
}

func (mr *MazeRenderer) renderAndOverlay(viewportData ViewportData, background *[]byte) (*[]byte, error) {
	mazeWallDataMap := orderedmap.NewOrderedMap[int, inf2.WallMapping]()
	for i := A_EAST; i <= Q_WEST; i++ {
		index := (*viewportData)[i]
		mazeWallDataMap.Set(i, *mr.viewportDataProvider.inf.FindWallMappingByIndex(index))
	}

	for renderPosition := range mazeWallDataMap.Keys() {
		mazeWallData, _ := mazeWallDataMap.Get(renderPosition)
		renderData := wallRenderData[renderPosition]

		var wallData = make([]byte, 0)
		wallDataPtr := &wallData

		needWallRender := true
		switch mazeWallData.WallMappingIndex {
		case 0:
			needWallRender = false
		case 1:
			wallDataPtr = mr.wallRenderer.RenderWall(1, renderData.wallIndex)
		case 2:
			wallDataPtr = mr.wallRenderer.RenderWall(1, renderData.wallIndex)
		case 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22:
			// door stuff
			wallDataPtr = mr.wallRenderer.RenderWall(2, renderData.wallIndex)
		case 23:
			wallDataPtr = mr.wallRenderer.RenderWall(3, renderData.wallIndex)
		case 24:
			wallDataPtr = mr.wallRenderer.RenderWall(4, renderData.wallIndex)
		default:
			if mazeWallData.WallSetId == 0 {
				needWallRender = false
			} else {
				wallDataPtr = mr.wallRenderer.RenderWall(mazeWallData.WallSetId, renderData.wallIndex)
			}
		}

		if needWallRender {
			w, h := mr.wallRenderer.GetWallSize(renderData.wallIndex)
			w *= 8
			h *= 8

			positionX := (renderData.offsetInViewPort % 22) * 8
			positionY := (renderData.offsetInViewPort / 22) * 8

			cropWidth := renderData.visibleWidthInBlocks * 8
			cropHeight := renderData.visibleHeightInBlocks * 8
			cropWallData := cropImage(wallDataPtr, w, cropWidth, cropHeight)
			background, _ = mr.overlayImage(background, cropWallData, cropWidth, cropHeight, positionX, positionY, renderData.flipFlag)
		}

		background = mr.decorationRenderer.DrawCompleteDecoration(background, mazeWallData, renderPosition, renderData.wallIndex != 0)
	}

	return background, nil
}

func cropImage(original *[]byte, originalWidth, newWidth, newHeight int) *[]byte {
	// Create a new slice for the cropped image
	cropped := make([]byte, newWidth*newHeight)

	// Iterate over each row and column of the new image
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			// Calculate the position in the original and cropped slices
			originalPos := y*originalWidth + x
			croppedPos := y*newWidth + x

			// Copy the pixel
			cropped[croppedPos] = (*original)[originalPos]
		}
	}

	return &cropped
}

func (mr *MazeRenderer) overlayImage(backgroundImageBytes *[]byte, anotherImageBytes *[]byte, anotherWidth int, anotherHeight int, dX int, dY int, flipFlag int) (*[]byte, error) {
	for row := 0; row < anotherHeight; row++ {
		for col := 0; col < anotherWidth; col++ {
			// Calculate the position in the first image
			destX := dX + col
			destY := dY + row

			// Check if the position is within the bounds of the first image
			destPos := destY*176 + destX

			var srcPos int
			if flipFlag == 1 {
				// Flip the image horizontally
				srcPos = row*anotherWidth + (anotherWidth - 1 - col)
			} else {
				// Normal positioning
				srcPos = row*anotherWidth + col
			}

			if destX >= 0 && destX < 176 && destY >= 0 && destY < 120 {
				if (*anotherImageBytes)[srcPos] != 0x00 {
					(*backgroundImageBytes)[destPos] = (*anotherImageBytes)[srcPos]
				}
			}
		}
	}
	return backgroundImageBytes, nil
}

// FlipImageOnX flips the image horizontally.
// `imageBytes` is the byte array representing the image.
// `width` and `height` are the dimensions of the image.
func flipBackgroundX(background *[]byte, width, height int) *[]byte {
	flippedImage := make([]byte, len(*background))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			flippedImage[y*width+x] = (*background)[y*width+width-x-1]
		}
	}

	return &flippedImage
}
