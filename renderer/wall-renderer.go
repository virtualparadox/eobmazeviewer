package renderer

import (
	"EOB1MazeViewer/formats"
	"fmt"
)

var offsetTable = [][]int{
	{330, 3, 15},
	{375, 3, 12},
	{411, 2, 8},
	{427, 1, 5},
	{432, 3, 5},
	{447, 2, 6},
	{459, 6, 5},
	{489, 10, 8},
	{569, 16, 12},
}

type WallRenderer struct {
	vcn *formats.VCN
	vmp *formats.VMP
}

func NewWallRenderer(vcn *formats.VCN, vmp *formats.VMP) *WallRenderer {
	return &WallRenderer{vcn: vcn, vmp: vmp}
}

func (wr *WallRenderer) Render(baseOffset int, width int, height int, colors []byte) *[]byte {
	rawPixels := make([]byte, width*8*height*8)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			offset := baseOffset + x + y*width
			var vmpCode = 0
			if offset > len(wr.vmp.Codes) {
				fmt.Printf("Invalid vmp code offset: %d\n", offset)
				continue
			}
			vmpCode = wr.vmp.Codes[offset]
			tileIndex := vmpCode & 0x3fff
			tileFlipped := (vmpCode & 0x4000) == 0x4000
			tile := wr.vcn.GetTile(tileIndex)

			wr.drawBlock(tileFlipped, x, y, width, rawPixels, colors, tile)
		}
	}
	return &rawPixels
}

func (wr *WallRenderer) drawBlock(tileFlipped bool, x int, y int, width int, rawPixels []byte, colors []byte, tile []byte) {
	if !tileFlipped {
		for py := 0; py < 8; py++ {
			for px := 0; px < 8; px++ {
				posDst := (x*8 + px) + (y*8+py)*width*8
				posSrc := px + py*8
				rawPixels[posDst] = colors[tile[posSrc]]
			}
		}
	} else {
		for py := 0; py < 8; py++ {
			for px := 0; px < 8; px++ {
				posDst := (x*8 + px) + (y*8+py)*width*8
				posSrc := 7 - px + py*8
				rawPixels[posDst] = colors[tile[posSrc]]
			}
		}
	}
}

func (wr *WallRenderer) RenderBackground() *[]byte {
	return wr.Render(0, 22, 15, wr.vcn.GetBackgroundColors())
}

func (wr *WallRenderer) RenderFakeBackground(v formats.VCN) *[]byte {
	result := make([]byte, 176*120)
	for i := 0; i < 176*120; i++ {
		result[i] = 0xC
	}
	return &result
}

func (wr *WallRenderer) GetBackgroundSize() (int, int) {
	return 22, 15
}

func (wr *WallRenderer) RenderWall(wallSet int, wall int) *[]byte {
	base := offsetTable[wall][0] + wallSet*431
	wallW := offsetTable[wall][1]
	wallH := offsetTable[wall][2]
	return wr.Render(base, wallW, wallH, wr.vcn.GetWallColors())
}

func (wr *WallRenderer) GetWallSize(wall int) (int, int) {
	wallW := offsetTable[wall][1]
	wallH := offsetTable[wall][2]
	return wallW, wallH
}
