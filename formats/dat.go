package formats

import (
	"bytes"
	"encoding/binary"
	"os"
)

type DecorationData struct {
	NbrDecorations          uint16
	Decorations             []Decoration
	NbrDecorationRectangles uint16
	Rectangles              []DecorationRectangle
}

type Decoration struct {
	RectangleIndices     [10]byte
	LinkToNextDecoration byte
	Flags                byte
	XCoords              [10]uint16
	YCoords              [10]uint16
}

type DecorationRectangle struct {
	X uint16
	Y uint16
	W uint16
	H uint16
}

func NewDATFromByteArray(rawData *[]byte) (*DecorationData, error) {
	reader := bytes.NewReader(*rawData)

	data := &DecorationData{}

	// Read number of decorations
	err := binary.Read(reader, binary.LittleEndian, &data.NbrDecorations)
	if err != nil {
		return nil, err
	}

	// Read decorations
	data.Decorations = make([]Decoration, data.NbrDecorations)
	for i := 0; i < int(data.NbrDecorations); i++ {
		var decoration Decoration
		err = binary.Read(reader, binary.LittleEndian, &decoration)
		if err != nil {
			return nil, err
		}

		// Handle 0xFF case for RectangleIndices
		for j, index := range decoration.RectangleIndices {
			if index == 0xFF {
				// Handle the case where there is no decoration
				// You can add specific logic here if needed
				decoration.RectangleIndices[j] = 0xFF
			}
		}

		data.Decorations[i] = decoration
	}

	// Read number of decoration rectangles
	err = binary.Read(reader, binary.LittleEndian, &data.NbrDecorationRectangles)
	if err != nil {
		return nil, err
	}

	// Read decoration rectangles
	data.Rectangles = make([]DecorationRectangle, data.NbrDecorationRectangles)
	for i := range data.Rectangles {
		err = binary.Read(reader, binary.LittleEndian, &data.Rectangles[i])
		if err != nil {
			return nil, err
		}
	}

	return data, nil

}

func NewDATDataFromFile(filename string) (*DecorationData, error) {
	rawData, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return NewDATFromByteArray(&rawData)
}
