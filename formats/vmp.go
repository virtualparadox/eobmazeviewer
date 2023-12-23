package formats

import (
	"encoding/binary"
	"fmt"
	"os"
)

type VMP struct {
	NbrCodes int
	Codes    []int
}

func NewVMPFromByteArray(data *[]byte) (*VMP, error) {
	if len(*data) < 2 {
		return nil, fmt.Errorf("file too short")
	}

	nrCodes := int(binary.LittleEndian.Uint16(*data))
	if len(*data) < 2+nrCodes*2 {
		return nil, fmt.Errorf("file too short for number of codes")
	}

	codes := make([]int, nrCodes)
	for i := 0; i < nrCodes; i++ {
		codes[i] = int((*data)[3+i*2])<<8 | int((*data)[2+i*2])
	}

	return &VMP{
		NbrCodes: nrCodes,
		Codes:    codes}, nil

}

func NewVMPFromFile(filename string) (*VMP, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return NewVMPFromByteArray(&data)
}
