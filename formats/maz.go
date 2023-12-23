package formats

import (
	"bytes"
	"encoding/binary"
	"os"
)

type MazeBlock struct {
	Wall [4]byte
}

type Maz struct {
	Width              uint16
	Height             uint16
	Nof                uint16
	WallMappingIndices []MazeBlock
}

func (maz *Maz) GetMazeBlockByCoordinateOrFake(x int, y int) *MazeBlock {
	if x < 0 || y < 0 || x >= int(maz.Width) || y >= int(maz.Height) {
		return FakeMazeBlock()
	}

	index := y*int(maz.Width) + x
	return &maz.WallMappingIndices[index]
}

func FakeMazeBlock() *MazeBlock {
	return &MazeBlock{
		Wall: [4]byte{0, 0, 0, 0},
	}
}

func NewMazFromByteArray(data *[]byte) (*Maz, error) {
	reader := bytes.NewReader(*data)
	// Read the width, height, and nof
	var m Maz
	err := binary.Read(reader, binary.LittleEndian, &m.Width)
	if err != nil {
		return nil, err
	}
	err = binary.Read(reader, binary.LittleEndian, &m.Height)
	if err != nil {
		return nil, err
	}
	err = binary.Read(reader, binary.LittleEndian, &m.Nof)
	if err != nil {
		return nil, err
	}

	// Calculate the total number of MazeBlocks
	totalBlocks := int(m.Width) * int(m.Height)
	m.WallMappingIndices = make([]MazeBlock, totalBlocks)

	// Read the MazeBlocks
	for i := 0; i < totalBlocks; i++ {
		err = binary.Read(reader, binary.LittleEndian, &m.WallMappingIndices[i])
		if err != nil {
			return nil, err
		}
	}

	return &m, nil
}

func NewMazFromFile(filename string) (*Maz, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return NewMazFromByteArray(&data)
}
