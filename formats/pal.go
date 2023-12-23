package formats

import (
	"image/color"
	"os"
)

type PAL struct {
	rawData []byte
	palette [][]byte
}

func NewPALFromByteArray(data *[]byte) (*PAL, error) {
	paletteLength := len(*data) / 3
	pal := &PAL{
		rawData: *data,
		palette: make([][]byte, paletteLength),
	}

	for i := 0; i < paletteLength; i++ {
		pal.palette[i] = make([]byte, 3)
		for j := 0; j < 3; j++ {
			pal.palette[i][j] = byte(float32((*data)[i*3+j]&0x3F) * 255 / 63)
		}
	}

	return pal, nil
}

func NewPALFromFile(filename string) (*PAL, error) {
	rawData, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return NewPALFromByteArray(&rawData)
}

func (p *PAL) GetPalette() color.Palette {
	if p.palette == nil {
		return nil
	}

	palette := make(color.Palette, len(p.palette))
	for i, col := range p.palette {
		palette[i] = color.RGBA{
			R: col[0],
			G: col[1],
			B: col[2],
			A: 255,
		}
	}

	return palette
}
