package inf

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"sort"
)

type Position struct {
	X, Y int
}

type Trigger struct {
	Pos           Position
	Flags         int
	CollapseFlags int
	Address       int
}

var collapseTriggerFlags = map[int]int{
	0x00: 0, 0x08: 1, 0x18: 2, 0x20: 3,
	0x28: 4, 0x40: 5, 0x48: 6, 0x60: 7,
	0x78: 8, 0x80: 9, 0x88: 10, 0xa8: 11,
}

func LoadTriggers(buffer *bytes.Reader, offset uint16) (*[]Trigger, error) {
	// save current position
	currentPosition, _ := buffer.Seek(0, io.SeekCurrent)

	buffer.Seek(int64(offset), io.SeekStart)
	var length uint16
	if err := binary.Read(buffer, binary.LittleEndian, &length); err != nil {
		return nil, err
	}

	triggers := make([]Trigger, length)

	// Now, triggers is a slice of *Trigger, with size determined by the value read
	// You can initialize each Trigger in the slice as needed
	for i := range triggers {
		t, _ := NewTriggerFromBytesReader(buffer)
		triggers[i] = *t
	}

	// jump back
	buffer.Seek(currentPosition, io.SeekStart)

	// sort triggers
	sort.Slice(triggers, func(i, j int) bool {
		if triggers[i].Pos.Y == triggers[j].Pos.Y {
			return triggers[i].Pos.X < triggers[j].Pos.X
		}
		return triggers[i].Pos.Y < triggers[j].Pos.Y
	})

	return &triggers, nil
}

func NewTriggerFromBytesReader(r *bytes.Reader) (*Trigger, error) {
	pos := rw(r)
	flags := rb(r)
	address := rw(r)
	x := pos & 31
	y := pos / 32

	position := Position{X: int(x), Y: int(y)}
	return NewTrigger(position, int(flags), int(address)), nil
}

func NewTrigger(pos Position, flags, address int) *Trigger {
	collapseFlags, ok := collapseTriggerFlags[flags]
	if !ok {
		collapseFlags = -1 // Or handle this case as needed
	}
	return &Trigger{
		Pos:           pos,
		Flags:         flags,
		Address:       address,
		CollapseFlags: collapseFlags,
	}
}

func (t Trigger) String() string {
	return fmt.Sprintf("Position: (%d, %d), Flags: %d, CollapseFlags: %d, Address: %d",
		t.Pos.X, t.Pos.Y, t.Flags, t.CollapseFlags, t.Address)
}
