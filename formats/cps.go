package formats

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)

type CPS struct {
	rawData []byte
}

func (cps CPS) GetRawData() *[]byte {
	return &cps.rawData
}

const (
	MaxSrcLen  = 64 * 1024   // max length of crunched data
	MaxDestLen = 1024 * 1024 // max length of unpacked data
)

var (
	ErrEmptyFile   = errors.New("empty file")
	ErrOverflow    = errors.New("output overflow")
	ErrUnknownComp = errors.New("unknown compression type")
	ErrOpenFile    = errors.New("can't open file")
	ErrZeroLength  = errors.New("zero-length file")
)

func NewCPSFromByteArray(data *[]byte) (*CPS, error) {
	slen := int(binary.LittleEndian.Uint16(*data))
	if slen != len(*data) && slen+2 != len(*data) {
		return nil, fmt.Errorf("invalid data stream length, cannot load CPS file")
	}

	compressionType := int(binary.LittleEndian.Uint16((*data)[2:]))
	fmt.Printf("\tcompressionType: %d\n", compressionType)

	uncompressedSize := int(binary.LittleEndian.Uint32((*data)[4:]))
	fmt.Printf("\tuncompressedSize: %d\n", uncompressedSize)

	paletteSize := int(binary.LittleEndian.Uint16((*data)[8:]))
	fmt.Printf("\tpalette size: %d\n", paletteSize)

	var decompressedData []byte
	var err error

	switch (*data)[2] {
	case 0:
		decompressedData, err = cpsCopy((*data)[4:])
	case 3:
		decompressedData, err = cpsRLE((*data)[4:])
	case 4:
		decompressedData, err = cpsLZ77((*data)[10:])
	default:
		return nil, fmt.Errorf("%w %d", ErrUnknownComp, (*data)[2])
	}

	if err != nil {
		return nil, err
	}

	cps := &CPS{
		rawData: decompressedData,
	}

	return cps, nil
}

func NewCPSFromFile(cpsFilename string) (*CPS, error) {
	fmt.Printf("Processing CPS %s", cpsFilename)
	data, err := os.ReadFile(cpsFilename)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cpsFilename, ErrOpenFile)
	}

	cps, err := NewCPSFromByteArray(&data)
	if err != nil {
		switch {
		case errors.Is(err, ErrEmptyFile):
			return nil, fmt.Errorf("%s: %w", cpsFilename, ErrZeroLength)
		case errors.Is(err, ErrOverflow):
			return nil, fmt.Errorf("%s: %w", cpsFilename, ErrOverflow)
		default:
			return nil, fmt.Errorf("%s: %w", cpsFilename, err)
		}
	}

	return cps, nil

}

func cpsCopy(src []byte) ([]byte, error) {
	if len(src) < 6 {
		return nil, ErrUnknownComp
	}

	// Get the length of the 'uncompressed' data
	length := int(binary.LittleEndian.Uint32(src[:4]))
	if length == 0 {
		return nil, ErrEmptyFile
	}

	offset := int(binary.LittleEndian.Uint16(src[4:6]))

	// Check for source overflow. We can't overflow the destination
	if length < 0 || length > (MaxSrcLen-offset-6) {
		return nil, ErrOverflow
	}

	// Skip to the start of data and do a direct copy from source to dest
	src = src[offset+6:]
	if length > len(src) {
		return nil, ErrOverflow
	}

	dest := make([]byte, length)
	copy(dest, src[:length])

	return dest, nil
}

func cpsRLE(src []byte) ([]byte, error) {
	if len(src) < 6 {
		return nil, ErrUnknownComp
	}

	length := int(binary.LittleEndian.Uint32(src[:4]))
	if length == 0 {
		return nil, ErrEmptyFile
	}

	if length < 0 || length > MaxDestLen {
		return nil, ErrOverflow
	}

	offset := int(binary.LittleEndian.Uint16(src[4:6]))
	src = src[offset+6:]

	dest := make([]byte, 0, length)

	for length > 0 {
		if length < 1 {
			return nil, ErrOverflow
		}
		rlen := int(src[0])
		src = src[1:]

		if rlen < 128 {
			if length < rlen || rlen > len(src) {
				return nil, ErrOverflow
			}
			dest = append(dest, src[:rlen]...)
			src = src[rlen:]
			length -= rlen
		} else {
			if rlen == 0 {
				if length < 2 {
					return nil, ErrOverflow
				}
				rlen = int(binary.LittleEndian.Uint16(src[:2]))
				src = src[2:]
			} else {
				rlen = 256 - rlen
			}

			if length < rlen || length < 1 {
				return nil, ErrOverflow
			}

			rep := src[0]
			src = src[1:]
			for i := 0; i < rlen; i++ {
				dest = append(dest, rep)
			}
			length -= rlen
		}
	}

	return dest, nil
}

func cpsLZ77(src []byte) ([]byte, error) {
	dest := make([]byte, 0, MaxDestLen)
	var rep int

	for i := 0; i < len(src) && len(dest) < MaxDestLen; {
		code := src[i]
		i++
		var length int

		switch {
		case code == 0xFE:
			length = int(binary.LittleEndian.Uint16(src[i : i+2]))
			i += 2
			code = src[i]
			i++
			if len(dest)+length > MaxDestLen {
				return nil, ErrOverflow
			}
			for length > 0 {
				dest = append(dest, code)
				length--
			}
		case code >= 0xC0:
			if code == 0xFF {
				length = int(binary.LittleEndian.Uint16(src[i : i+2]))
				i += 2
			} else {
				length = int(code&0x3F) + 3
			}
			rep = int(binary.LittleEndian.Uint16(src[i : i+2]))
			i += 2
			if len(dest)+length > MaxDestLen {
				return nil, ErrOverflow
			}
			for length > 0 && rep < len(dest) {
				dest = append(dest, dest[rep])
				rep++
				length--
			}
		case code >= 0x80:
			if code == 0x80 {
				break
			}
			length = int(code & 0x3F)
			rep = i
			i += length
			if i > len(src) {
				return nil, ErrOverflow
			}
			dest = append(dest, src[rep:i]...)
		default:
			length = int(code>>4) + 3
			rep = len(dest) - (int(code&0x0F)<<8 | int(src[i]))
			i++
			if len(dest)+length > MaxDestLen || rep < 0 {
				return nil, ErrOverflow
			}
			for length > 0 && rep < len(dest) {
				dest = append(dest, dest[rep])
				rep++
				length--
			}
		}
	}
	return dest, nil
}
