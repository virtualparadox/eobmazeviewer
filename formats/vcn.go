package formats

import (
	"encoding/binary"
)

type VCN struct {
	rawData          []byte
	numberOfTiles    int
	tiles            [][]byte
	backgroundColors []byte
	wallColors       []byte
}

func (v *VCN) GetBackgroundColors() []byte {
	return v.backgroundColors
}

func (v *VCN) GetTile(index int) []byte {
	if index > len(v.tiles) {
		return v.tiles[0]
	}
	return v.tiles[index]
}

func (v *VCN) GetWallColors() []byte {
	return v.wallColors
}

func NewVCNFromByteArray(data *[]byte) (*VCN, error) {
	cps, _ := NewCPSFromByteArray(data)
	return buildVCN(cps.GetRawData())
}

func NewVCNFromFile(filename string) (*VCN, error) {
	cps, _ := NewCPSFromFile(filename)
	data := cps.GetRawData()
	return buildVCN(data)
}

func buildVCN(data *[]byte) (*VCN, error) {
	vcn := &VCN{}
	vcn.rawData = *data

	vcn.numberOfTiles = int(binary.LittleEndian.Uint16(*data)) //int(data[0]) | int(data[1])<<8
	vcn.tiles = make([][]byte, vcn.numberOfTiles)
	for i := range vcn.tiles {
		vcn.tiles[i] = make([]byte, 64)
	}

	vcn.backgroundColors = make([]byte, 16)
	vcn.wallColors = make([]byte, 16)
	for i := 0; i < 16; i++ {
		d := *data
		vcn.backgroundColors[i] = d[2+i]
		vcn.wallColors[i] = d[2+16+i]
	}

	for i := 0; i < vcn.numberOfTiles; i++ {
		offset := 2 + 32 + i*32
		for j := 0; j < 32; j++ {
			d := *data
			twoPixels := d[offset+j]
			vcn.tiles[i][j*2+0] = (twoPixels >> 4) & 0x0f
			vcn.tiles[i][j*2+1] = twoPixels & 0x0f
		}
	}

	return vcn, nil
}
