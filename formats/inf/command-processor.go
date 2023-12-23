package inf

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

func rb(buffer *bytes.Reader) byte {
	b, _ := buffer.ReadByte()
	return b
}

func rw(buffer *bytes.Reader) uint16 {
	var word uint16
	binary.Read(buffer, binary.LittleEndian, &word)
	return word
}

func ProcessCommand(command byte, buffer *bytes.Reader, lastOffset int64) bool {
	fmt.Printf("command %04x", command)
	switch command {
	case 0xee:
		error := !parseConditional(buffer, lastOffset)
		return error
	case 0xe6:
		parseEncounters(buffer)
	case 0xf8:
		parseMessage(buffer)
	case 0xf7:
		parseSetFlag(buffer)
	}
	return false
}

func parseSetFlag(buffer *bytes.Reader) {

}

func parseMessage(buffer *bytes.Reader) {
	var result bytes.Buffer
	buf := make([]byte, 1)

	for {
		_, err := buffer.Read(buf)
		if err != nil {
			if err == io.EOF {
				panic("Reached end of file while reading a null terminated string")
			}
		}

		if buf[0] == 0 {
			break
		}

		result.WriteByte(buf[0])
	}
	color := rb(buffer)
	fmt.Printf("Message: %s Color: %d\n", result.String(), color)
}

func parseEncounters(buffer *bytes.Reader) {
	fmt.Printf("Encounters\n;{\n")

	unknown := rb(buffer)
	fmt.Printf(";   Encounter#: $%02x\n", unknown)

	fmt.Printf(";}\n")
}

func parseConditional(buffer *bytes.Reader, lastOffset int64) bool {
	dynamicStack := false
	simulatedStack := NewSimulatedStack()
	stackDepth := 0
	fmt.Println("Conditional\n{")

	for {
		command := rb(buffer)
		if command == 0xee {
			break
		}

		fmt.Print(";   ")
		for i := 0; i < stackDepth; i++ {
			fmt.Print(">")
		}

		if command >= 0x80 && command <= 0xf7 {
			dynamicStack = true
		}

		switch command {
		case 0xf3:
			subcode := rb(buffer)
			if subcode == 0xff {
				pos := rw(buffer)
				fmt.Printf("push(countMonstersAt([%d,%d]))\n", pos&31, pos/32)
				stackDepth++
			} else {
				for {
					fmt.Printf("push(countMonstersOfType(%d)); ", subcode)
					comparator := rb(buffer)
					fmt.Printf("push(0x%02x)", comparator)
					stackDepth += 2

					subcode = rb(buffer)
					if subcode == 0 {
						break
					} else {
						fmt.Print("; ")
					}
				}
				fmt.Println()
			}
		case 0xda:
			fmt.Printf("push(isPartyVisible())\n")
			stackDepth++
		case 0xdb:
			rolls := rb(buffer)
			sides := rb(buffer)
			base := rb(buffer)
			fmt.Printf("push(rollDice(%dT%d+%d))\n", rolls, sides, base)
			stackDepth++
		case 0xdd:
			fmt.Printf("push(party.containsRace(%d))\n", rb(buffer))
			stackDepth++
		case 0xce:
			fmt.Printf("push(party.containsAlignment(%d))\n", rb(buffer))
			stackDepth++
		case 0xdc:
			fmt.Printf("push(party.containsClass(%d))\n", rb(buffer))
			stackDepth++
		case 0xe0:
			fmt.Printf("push(trigger.Flags)\n")
			stackDepth++
		case 0xed:
			fmt.Printf("push(party.getDirection())\n")
			stackDepth++
		case 0xf0:
			fmt.Printf("push(getFlag(Global, %d))\n", rb(buffer))
			stackDepth++
		case 0xe7:
			subcode := rb(buffer)
			if subcode == 0xe1 {
				fmt.Printf("push(party.pointeritem.type)\n")
			} else if subcode == 0xf5 {
				fmt.Printf("push(party.pointeritem)\n")
			} else if subcode == 0xf6 {
				fmt.Printf("push(party.pointeritem.value)\n")
			} else if subcode == 0xd0 {
				value := rb(buffer)
				fmt.Printf("push(party.pointeritem.unidname==%d)\n", value)
			} else if subcode == 0xcf {
				value := rb(buffer)
				fmt.Printf("push(party.pointeritem.idname==%d)\n", value)
			} else {
				fmt.Printf("push(party.pointeritem.???\n")
				os.Exit(0)
			}
			stackDepth++
		case 0xe9:
			side := rb(buffer)
			pos := rw(buffer)
			fmt.Printf("push(maze.getWallSide(%d, [%d, %d])\n", side, pos&31, pos/32)
			stackDepth++
		case 0xf1: // Party
			subcode := rb(buffer)
			if subcode == 0xf5 { // Count items
				t := rw(buffer)
				flags := rb(buffer)
				fmt.Printf("push(party.inventory.count(0x%04x(type), 0x%02x(Flags?))\n", t, flags)
			} else { // Check party position
				pos := (rb(buffer) << 8) | subcode
				fmt.Printf("push(party.getPos()==[%d,%d])\n", pos&31, pos/32)
			}
			stackDepth++
		case 0xf5:
			itemtype := rb(buffer)
			pos := rw(buffer)

			if itemtype == 0xff {
				fmt.Printf("push(maze.countItems([%d,%d], item.type=ANY))\n", pos&31, pos/32)
			} else {
				fmt.Printf("push(maze.countItems([%d,%d], item.type=$%02x))\n", pos&31, pos/32, itemtype)
			}

			stackDepth++
		case 0xf7:
			pos := rw(buffer)
			fmt.Printf("push(maze.getWallNumber([%d, %d]))\n", pos&31, pos/32)
			stackDepth++
		case 0xef:
			flag := rb(buffer)
			fmt.Printf("push(maze.getFlag(%d))\n", flag)
			stackDepth++
		case 0xff:
			if !dynamicStack {
				a := simulatedStack.Back()
				simulatedStack.PopBack()
				b := simulatedStack.Back()
				simulatedStack.PopBack()
				//simulatedStack.push_back(a==b);
				if a == b {
					simulatedStack.PushBack(1)
				} else {
					simulatedStack.PushBack(0)
				}
			}
			fmt.Printf("push(pop()==pop())\n")
			stackDepth--
		case 0xfe:
			if !dynamicStack {
				a := simulatedStack.Back()
				simulatedStack.PopBack()
				b := simulatedStack.Back()
				simulatedStack.PopBack()
				//simulatedStack.push_back(a!=b);
				if a != b {
					simulatedStack.PushBack(1)
				} else {
					simulatedStack.PushBack(0)
				}
			}
			fmt.Printf("push(pop()!=pop())\n")
			stackDepth--
		case 0xfd:
			if !dynamicStack {
				a := simulatedStack.Back()
				simulatedStack.PopBack()
				b := simulatedStack.Back()
				simulatedStack.PopBack()
				//simulatedStack.PushBack(a<b);
				if a < b {
					simulatedStack.PushBack(1)
				} else {
					simulatedStack.PushBack(0)
				}
			}
			fmt.Printf("push(pop()<pop())\n")
			stackDepth--
		case 0xfc:
			if !dynamicStack {
				a := simulatedStack.Back()
				simulatedStack.PopBack()
				b := simulatedStack.Back()
				simulatedStack.PopBack()
				//simulatedStack.PushBack(a<=b);
				if a <= b {
					simulatedStack.PushBack(1)
				} else {
					simulatedStack.PushBack(0)
				}
			}
			fmt.Printf("push(pop()<=pop())\n")
			stackDepth--
		case 0xfb:
			if !dynamicStack {
				a := simulatedStack.Back()
				simulatedStack.PopBack()
				b := simulatedStack.Back()
				simulatedStack.PopBack()
				//simulatedStack.PushBack(a>b);
				if a > b {
					simulatedStack.PushBack(1)
				} else {
					simulatedStack.PushBack(0)
				}
			}
			fmt.Printf("push(pop()>pop())\n")
			stackDepth--
		case 0xfa:
			if !dynamicStack {
				a := simulatedStack.Back()
				simulatedStack.PopBack()
				b := simulatedStack.Back()
				simulatedStack.PopBack()
				//simulatedStack.PushBack(a>=b);
				if a >= b {
					simulatedStack.PushBack(1)
				} else {
					simulatedStack.PushBack(0)
				}
			}
			fmt.Printf("push(pop()>=pop())\n")
			stackDepth--
		case 0xf9:
			if !dynamicStack {
				a := simulatedStack.Back()
				simulatedStack.PopBack()
				b := simulatedStack.Back()
				simulatedStack.PopBack()
				//simulatedStack.PushBack(a&&b);
				aBool := a != 0
				bBool := b != 0
				if aBool && bBool {
					simulatedStack.PushBack(1)
				} else {
					simulatedStack.PushBack(0)
				}
			}
			fmt.Printf("push(pop()&&pop())\n")
			stackDepth--
		case 0xf8:
			if !dynamicStack {
				a := simulatedStack.Back()
				simulatedStack.PopBack()
				b := simulatedStack.Back()
				simulatedStack.PopBack()
				//simulatedStack.PushBack(a||b);
				aBool := a != 0
				bBool := b != 0
				if aBool || bBool {
					simulatedStack.PushBack(1)
				} else {
					simulatedStack.PushBack(0)
				}
			}
			fmt.Printf("push(pop()||pop())\n")
			stackDepth--
		case 0x00:
			if !dynamicStack {
				simulatedStack.PushBack(0)
			}

			fmt.Printf("push(false/0)\n")
			stackDepth++
		case 0x01:
			if !dynamicStack {
				simulatedStack.PushBack(1)
			}

			fmt.Printf("push(true/1)\n")
			stackDepth++
		default:
			if !dynamicStack {
				simulatedStack.PushBack(int(command))
			}

			fmt.Printf("push(0x%02x)\n", command)
			stackDepth++

		}
	}

	// Handle other cases as per the original C++ code.

	falseAddress := rw(buffer)
	fmt.Printf(";   if (!pop()) then jump 0x%04x\n", falseAddress)
	fmt.Println(";}")
	stackDepth--

	if stackDepth != 0 {
		fmt.Printf("CONDITIONAL PARSE ERROR: StackDepth was %d at exit!\n", stackDepth)
		return false
	}

	// Handle the dynamicStack and simulatedStack logic.
	if (!dynamicStack) && (simulatedStack.Size() == 1) {
		returnValue := simulatedStack.Back()
		simulatedStack.PopBack()
		if returnValue == 0 {
			fmt.Printf("; Always false\n")
			fmt.Printf(".assert _0x%04x - * <= 255, error, \"Illegal branch\"\n", falseAddress)
			fmt.Printf(".byte $f2, <(_0x%04x - *)\n", falseAddress)
		} else {
			fmt.Printf("; Always true\n")
		}
	} else {
		offset := getOffset(buffer)
		endOffset := offset - 3
		if endOffset-lastOffset < 128 {
			fmt.Printf(".byte $%02x", offset-3-lastOffset)
			endOffset--
		} else {
			fmt.Printf(".byte $ee")
		}

		for i := lastOffset + 1; i <= endOffset; i++ {
			fmt.Printf(",$%02x", readByteFrom(buffer, i))
		}
		fmt.Printf("\n", readByteFrom(buffer, endOffset))

		// Previously we changed the end offset to x-3 and read from there
		// this is why we need to step 3 bytes in the buffer. Sorry, awful solution,
		// will be fixed later.
		rb(buffer)
		rb(buffer)
		rb(buffer)

		//printf("<_0x%04x,>_0x%04x", falseAddress, falseAddress);
		if lastOffset < int64(falseAddress) {
			fmt.Printf(".assert _0x%04x - * <= 255, error, \"Illegal branch\"\n", falseAddress)
			fmt.Printf(".byte <(_0x%04x - *)\n", falseAddress)
		} else {
			fmt.Printf("Conditional branch makes a negative jump at 0x%04x\n", lastOffset)
			return false
		}
		fmt.Printf("\n")
	}

	return true
}

func getOffset(buffer *bytes.Reader) int64 {
	offset, _ := buffer.Seek(0, io.SeekCurrent)
	return offset
}

func readByteFrom(buffer *bytes.Reader, i int64) any {
	buffer.Seek(i, io.SeekStart)
	return rb(buffer)
}
