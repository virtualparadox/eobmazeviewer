package formats

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"golang.org/x/exp/maps"
	"strings"
)

type Monster struct {
	Index      uint8
	LevelType  uint8
	Pos        uint16
	Subpos     uint8
	Direction  uint8
	Type       uint8
	Picture    uint8
	Phase      uint8
	Pause      uint8
	Weapon     uint16
	PocketItem uint16
}

type rawInfHeader struct {
	TriggerOffset                      uint16
	MazeName                           [12]byte
	VmpVcnName                         [12]byte
	PaletteName                        [12]byte
	Unknown1                           [4]byte
	TypeOfCallingCommand1              uint8
	Command1GenerationFrequencyInTicks uint16
	Command1GenerationFrequencyInSteps uint16
	Monster1CompressionMethod          uint8
	Monster1Name                       [12]byte
	Monster2CompressionMethod          uint8
	Monster2Name                       [12]byte
	Unknown2                           [5]byte
	Monsters                           [30]Monster
	NbrDecCommands                     uint16
}

type InfHeader struct {
	TriggersOffset                     uint16
	MazeName                           string
	VmpVcnName                         string
	PaletteName                        string
	TypeOfCallingCommand1              uint8
	Command1GenerationFrequencyInTicks uint16
	Command1GenerationFrequencyInSteps uint16
	Monster1CompressionMethod          uint8
	Monster1Name                       string
	Monster2CompressionMethod          uint8
	Monster2Name                       string
	Monsters                           [30]Monster
	WallMapping                        map[int]WallMapping
}

type WallMapping struct {
	WallMappingIndex int
	WallSetId        int
	DecorationId     int
	DatName          string
	CpsName          string
	EventMask        int
	Flags            int
}

func (inf *InfHeader) GetMazeName() string {
	return inf.MazeName
}

func toString(name [12]byte) string {
	pos := clean(name[:])
	var builder strings.Builder
	for i := 0; i < pos; i++ {
		builder.WriteByte(name[i])
	}
	return builder.String()
}

func clean(n []byte) int {
	for i := 0; i < len(n); i++ {
		if n[i] == 0 {
			return i
		}
	}
	return len(n)
}

func NewInfFromByteArray(data *[]byte) (*InfHeader, error) {
	cps, err := NewCPSFromByteArray(data)
	if err != nil {
		return nil, fmt.Errorf("unable to load CPS")
	}
	return buildInfHeader(cps, err)
}

func NewInfFromFile(filename string) (*InfHeader, error) {
	cps, err := NewCPSFromFile(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to uncps %s", filename)
	}

	return buildInfHeader(cps, err)
}

func buildInfHeader(cps *CPS, err error) (*InfHeader, error) {
	var internalInfHeader rawInfHeader

	// Read fixed-size fields
	data := cps.GetRawData()
	buffer := bytes.NewReader(*data)
	err = binary.Read(buffer, binary.LittleEndian, &internalInfHeader)
	if err != nil {
		return nil, err
	}

	wallMap := prefillWallMappings()

	wallMap, err = loadDecoration(buffer, internalInfHeader.NbrDecCommands, wallMap)
	if err != nil {
		return nil, err
	}

	result := InfHeader{
		TriggersOffset:                     internalInfHeader.TriggerOffset,
		MazeName:                           toString(internalInfHeader.MazeName),
		VmpVcnName:                         toString(internalInfHeader.VmpVcnName),
		PaletteName:                        toString(internalInfHeader.PaletteName),
		TypeOfCallingCommand1:              internalInfHeader.TypeOfCallingCommand1,
		Command1GenerationFrequencyInSteps: internalInfHeader.Command1GenerationFrequencyInSteps,
		Command1GenerationFrequencyInTicks: internalInfHeader.Command1GenerationFrequencyInTicks,
		Monster1CompressionMethod:          internalInfHeader.Monster1CompressionMethod,
		Monster1Name:                       toString(internalInfHeader.Monster1Name),
		Monster2CompressionMethod:          internalInfHeader.Monster2CompressionMethod,
		Monster2Name:                       toString(internalInfHeader.Monster2Name),
		Monsters:                           internalInfHeader.Monsters,
		WallMapping:                        *wallMap,
	}

	return &result, nil
}

func loadDecoration(buffer *bytes.Reader, commands uint16, wallMap *map[int]WallMapping) (*map[int]WallMapping, error) {
	currentCpsName := ""
	currentDatName := ""

	for i := 0; i < int(commands); i++ {
		command, _ := buffer.ReadByte()
		if command == 0xEC {
			// read DatName name
			rawCpsName := make([]byte, 12)
			_, err := buffer.Read(rawCpsName)
			if err != nil {
				return nil, err
			}

			rawDatName := make([]byte, 12)
			_, err = buffer.Read(rawDatName)
			if err != nil {
				return nil, err
			}

			currentCpsName = toString([12]byte(rawCpsName))
			currentDatName = toString([12]byte(rawDatName))

			fmt.Printf("0xEC: %s - %s\n", currentCpsName, currentDatName)
		} else if command == 0xFB {
			wallMappingIndex, _ := buffer.ReadByte()
			wallType, _ := buffer.ReadByte()
			decorationId, _ := buffer.ReadByte()
			evantMask, _ := buffer.ReadByte()
			flags, _ := buffer.ReadByte()

			fmt.Printf("0xFB: %d %d %d %d %d\n", wallMappingIndex, wallType, decorationId, evantMask, flags)
			wm := WallMapping{WallMappingIndex: int(wallMappingIndex), WallSetId: int(wallType), DecorationId: int(decorationId), EventMask: int(evantMask), Flags: int(flags), DatName: currentDatName, CpsName: currentCpsName}
			(*wallMap)[int(wallMappingIndex)] = wm
		}

	}
	return wallMap, nil
}

func prefillWallMappings() *map[int]WallMapping {
	// Prefill wallmappings with standard walls
	var wallTypeInit = []int{0, 1, 2, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3}
	var wallEventMaskInit = []int{0, 0, 0, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0}
	var wallFlagsInit = []int{1, 4, 4, 0x2c, 0x2c, 0x2c, 0x2c, 0x19, 0x2c, 0x2c, 0x2c, 0x2c, 0x19, 0x2e, 0x2e, 0x2e, 0x2e, 0x19, 0x2e, 0x2e, 0x2e, 0x2e, 0x19}

	wallMap := make(map[int]WallMapping)
	for i := 0; i < len(wallTypeInit); i++ {
		wm := WallMapping{WallMappingIndex: i, WallSetId: wallTypeInit[i], DecorationId: 0xff, EventMask: wallEventMaskInit[i], Flags: wallFlagsInit[i]}
		wallMap[i] = wm
	}

	return &wallMap
}

func (inf *InfHeader) GetDecorationCPSNames() []string {
	var set = make(map[string]bool)
	for _, v := range inf.WallMapping {
		set[v.CpsName] = true
	}

	result := maps.Keys(set)
	return result
}

func (inf *InfHeader) FindWallMappingByIndex(index byte) *WallMapping {
	for _, value := range inf.WallMapping {
		if value.WallMappingIndex == int(index) {
			return &value
		}
	}
	return nil
}
