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

func ParseScripts(buffer *bytes.Reader, triggers *[]Trigger, triggerOffset uint16) error {
	//error := false
	offset := getOffset(buffer)

	var error = false
	for offset < int64(triggerOffset) && !error {
		trigger := checkTriggerReference(triggers, offset)
		fmt.Printf("_0x%04x: ;", offset)

		lastOffset := offset
		command := rb(buffer)
		switch command {
		case 0xff:
			parseSetWall(buffer)
		case 0xfe:
			parseChangeWall(buffer)
		case 0xfd:
			parseOpenDoor(buffer)
		case 0xfc:
			parseCloseDoor(buffer)
		case 0xfb:
			parseCreateMonster(buffer)
		case 0xfa:
			parseTeleport(buffer)
		case 0xf9:
			parseStealSmallItems(buffer)
		case 0xf8:
			parseMessage(buffer)
		case 0xf7:
			parseSetFlag(buffer)
		case 0xf6:
			parseSound(buffer)
		case 0xf5:
			parseClearFlag(buffer)
		case 0xf4:
			// parseHeal (never used in EobI)
			continue
		case 0xf3:
			parseDamage(buffer)
		case 0xf2:
			error = !parseJump(buffer, lastOffset)
		case 0xf1:
			parseEndCode(buffer)
		case 0xf0:
			parseReturn(buffer)
		case 0xef:
			parseCall(buffer)
		case 0xee:
			error = !parseConditional(buffer, lastOffset)
		case 0xed:
			parseItemConsume(buffer)
		case 0xec:
			currentLevel := 0
			parseChangeLevel(buffer, trigger, currentLevel)
		case 0xeb:
			parseGiveXP(buffer)
		case 0xea:
			parseNewItem(buffer)
		case 0xe9:
			parseLauncher(buffer)
		case 0xe8:
			parseTurn(buffer)
		case 0xe7:
			parseIdentAllItems(buffer)
		case 0xe6:
			parseEncounters(buffer)
		case 0xe5:
			parseWait(buffer)
		case 0xe4:
			// parseUpdateScreen (never used in EobI)
			continue
		case 0xe3:
			// parseTextMenu (never used in EobI)
			continue
		case 0xe2:
			// parseSpecialWindowPictures (never used in EobI)
			continue
		default:
			fmt.Printf("Unknown code 0x%02x offset: %d\n", command, offset-1)
			offset = lastOffset + 16
			error = true
		}
		offset = getOffset(buffer)
	}
	return nil
}

func parseWait(buffer *bytes.Reader) {
	fmt.Printf("Wait\n;{\n")

	ticks := rw(buffer)
	fmt.Printf(";   Ticks = %d\n", ticks)

	fmt.Printf("}\n")
}

func parseIdentAllItems(buffer *bytes.Reader) {
	fmt.Printf("IdentAllItems\n;{\n")

	pos := rw(buffer)

	fmt.Printf(";   Position = [%d,%d]\n", pos&31, pos/32)

	fmt.Printf(";}\n")
}

func parseTurn(buffer *bytes.Reader) {
	fmt.Printf("Turn\n;{\n")

	t := rb(buffer)
	dir := rb(buffer)

	if t == 0xf1 {
		fmt.Printf(";   Type = Party (0x%02x)\n", t)
	} else if t == 0xf5 {
		fmt.Printf(";   Type = Item (0x%02x)\n", t)
	} else {
		fmt.Printf(";   Type = Unknown (0x%02x)\n", t)
		fmt.Printf(";   Direction = %d\n", dir)
	}

	fmt.Printf(";}\n")
}

func parseLauncher(buffer *bytes.Reader) {
	fmt.Printf("Launcher\n;{\n")

	kind := rb(buffer)
	itemno := rw(buffer)
	pos := rw(buffer)
	dir := rb(buffer)
	subpos := rb(buffer)

	var t = kind == 0xdf
	var n = ""
	if t {
		n = "Spell"
	} else {
		n = "Item"
	}

	fmt.Printf(";   Kind = %s\n", n)
	fmt.Printf(";   Item#/Spell# = %d\n", itemno)
	fmt.Printf(";   Pos = [%d,%d:%d]\n", pos&31, pos/32, subpos)
	fmt.Printf(";   Direction = %d\n", dir)

	fmt.Printf(";}\n")
}

func parseNewItem(buffer *bytes.Reader) {
	fmt.Printf("NewItem\n;{\n")

	itemno := rw(buffer)
	pos := rw(buffer)
	subpos := rb(buffer)
	fmt.Printf(";   Item# = $%04x\n", itemno)
	if pos != 0xffff {
		fmt.Printf(";   Position = [%d,%d:%d]\n", pos&31, pos/32, subpos)
	} else {
		fmt.Printf(";   Position = n/a\n")
	}

	fmt.Printf(";}\n")
}

func parseGiveXP(buffer *bytes.Reader) {
	fmt.Printf("GiveExperience\n;{\n")

	t := rb(buffer)
	if t == 0xe2 {
		amount := rw(buffer)
		fmt.Printf(";   Type = Party\n")
		fmt.Printf(";   Amount = %d\n", amount)
	} else {
		fmt.Printf(";   Type = Unknown (0x%02x)\n", t)
	}

	fmt.Printf(";}\n")
}

func parseChangeLevel(buffer *bytes.Reader, trigger *Trigger, currentLevel int) {
	fmt.Printf("ChangeLevel\n;{\n")

	t := rb(buffer)
	if t == 0xe5 {
		fmt.Printf(";   Type = Real level change\n")
		level := rb(buffer)
		position := rw(buffer)
		direction := rb(buffer)
		x := position & 31
		y := position / 32

		fmt.Printf(";   Target =   X:%d\n", x)
		fmt.Printf(";              Y:%d\n", y)
		fmt.Printf(";            Dir:%d\n", direction)
		fmt.Printf(";            Lvl:%d\n", level)

		fmt.Printf(";}\n")

		// A simple hole!
		if (int(level) == currentLevel+1) && (direction == 255) && (x == uint16(trigger.Pos.X)) && (y == uint16(trigger.Pos.Y)) {
			fmt.Printf(".byte $e4; New C64 byte code for falling down.\n")
			return
		} else {
			//emitC64Raw();
		}
	} else {
		fmt.Printf(";   Type = Inter level change\n")
		direction := rb(buffer)
		position := rw(buffer)
		fmt.Printf(";   Target =   X:%d\n", position&31)
		fmt.Printf(";              Y:%d\n", position/32)
		fmt.Printf(";            Dir:%d\n", direction)

		fmt.Printf(";}\n")
		//emitC64Raw();
	}

}

func parseItemConsume(buffer *bytes.Reader) {
	fmt.Printf("ItemConsume\n;{\n")

	loci := rb(buffer)

	if loci == 0xff {
		fmt.Printf(";   Location = Mouse pointer\n")
	} else if loci == 0xfe {
		pos := rw(buffer)
		fmt.Printf(";   Position = [%d,%d:*]\n", pos&31, pos/32)
	} else {
		pos := rw(buffer)
		fmt.Printf(";   Position = [%d,%d]. Item.type=$%02x\n", pos&31, pos/32, loci)
	}

	fmt.Printf(";}\n")
}

func parseCall(buffer *bytes.Reader) {
	address := rw(buffer)
	fmt.Printf("Call 0x%04x\n", address)

	fmt.Printf(".byte $ef,<_0x%04x,>_0x%04x\n", address, address)

}

func parseReturn(buffer *bytes.Reader) {
	fmt.Printf("Return\n")
}

func parseEndCode(buffer *bytes.Reader) {
	fmt.Printf("Abort event\n")
}

func parseJump(buffer *bytes.Reader, lastOffset int64) bool {
	address := rw(buffer)
	fmt.Printf("jump 0x%04x\n", address)

	if lastOffset < int64(address) {
		fmt.Printf(".assert _0x%04x - * <= 255, error, \"Illegal branch\"\n", address)
		fmt.Printf(".byte $f2, <(_0x%04x - *)\n", address)
	} else {
		fmt.Printf("ERROR: Jump is negative at 0x%04x\n", lastOffset)
		return false
	}
	return true

}

func parseDamage(buffer *bytes.Reader) {
	fmt.Printf("Damage\n;{\n")

	whom := rb(buffer)
	flag1 := rb(buffer)
	flag2 := rb(buffer)
	flag3 := rb(buffer)

	fmt.Printf(";   Whom = ")
	if whom == 0xff {
		fmt.Printf("All\n")
	} else {
		fmt.Printf("Memeber %d\n", whom)
	}

	fmt.Printf(";   Flag1 = 0x%02x\n", flag1)
	fmt.Printf(";   Flag2 = 0x%02x\n", flag2)
	fmt.Printf(";   Flag3 = 0x%02x\n", flag3)

	fmt.Printf(";}\n")
}

func parseClearFlag(buffer *bytes.Reader) {
	fmt.Printf("ClearFlag\n;{\n")

	target := rb(buffer)
	fmt.Printf(";   Target = ")
	if target == 0xef {
		fmt.Printf("Maze\n")
		flag := rb(buffer)
		fmt.Printf(";   Flag = %d\n", flag)
	} else if target == 0xf0 {
		fmt.Printf("Global\n")
		flag := rb(buffer)
		fmt.Printf(";   Flag = %d\n", flag)
	} else if target == 0xf3 {
		fmt.Printf("Monster\n")
		id := rb(buffer)
		flag := rb(buffer)
		fmt.Printf(";   Monster = %d\n", id)
		fmt.Printf(";   Flag = %d\n", flag)
	} else if target == 0xe4 {
		fmt.Printf("Event\n")
	} else if target == 0xd1 {
		fmt.Printf("Party_Function(FUNC_SETVAL, PARTY_SAVEREST, 0);\n")
	}
	fmt.Printf(";}\n")
}

func parseSound(buffer *bytes.Reader) {
	fmt.Printf("Sound\n;{\n")

	id := rb(buffer)
	pos := rw(buffer)

	if pos > 0 {
		fmt.Printf(";   ID: $%02x\n", id)
		fmt.Printf(";   Position: [%d,%d]\n", pos&31, pos/32)
	} else {
		fmt.Printf(";   ID: $%02x\n", id)
	}

	fmt.Printf(";}\n")
}

func parseStealSmallItems(buffer *bytes.Reader) {
	fmt.Printf("StealSmallItems\n;{\n")

	whom := rb(buffer)

	fmt.Printf(";   Whom = ")
	if whom == 0xff {
		fmt.Printf("Random\n")
	} else {
		fmt.Printf("Member %d\n", whom)
	}

	pos := rw(buffer)
	subpos := rb(buffer)
	fmt.Printf(";   Drop position = [%d,%d:%d]\n", pos&31, pos/32, subpos)

	fmt.Printf(";}\n")
}

func parseTeleport(buffer *bytes.Reader) {
	fmt.Printf("Teleport\n;{\n")

	t := rb(buffer)
	var source uint16 = 0
	var dest uint16 = 0

	switch t {
	case 0xe8: // Teleport party
		rw(buffer)
		dest = rw(buffer)
		fmt.Printf(";   Type = Party\n")
		fmt.Printf(";   Dest = [%d,%d]\n", dest&31, dest/32)
	case 0xf3: // Monster
		source = rw(buffer)
		dest = rw(buffer)
		fmt.Printf(";   Type = Monster\n")
		fmt.Printf(";   Source = [%d,%d]\n", source&31, source/32)
		fmt.Printf(";   Dest = [%d,%d]\n", dest&31, dest/32)
	case 0xf5: // Item
		source = rw(buffer)
		dest = rw(buffer)
		fmt.Printf(";   Type = Item\n")
		fmt.Printf(";   Source = [%d,%d]\n", source&31, source/32)
		fmt.Printf(";   Dest = [%d,%d]\n", dest&31, dest/32)
	default:
		source = rw(buffer)
		dest = rw(buffer)
		fmt.Printf(";   Type = Unknown ($%02x)\n", t)
		fmt.Printf(";   Source = [%d,%d]\n", source&31, source/32)
		fmt.Printf(";   Dest = [%d,%d]\n", dest&31, dest/32)
	}

	fmt.Printf(";}\n")

}

func parseCreateMonster(buffer *bytes.Reader) {
	fmt.Printf("CreateMonster\n;{\n")
	rb(buffer)

	movetime := rb(buffer)
	pos := rw(buffer)
	subpos := rb(buffer)
	dir := rb(buffer)
	t := rb(buffer)
	pic := rb(buffer)
	phase := rb(buffer)
	pause := rb(buffer)
	pocket := rw(buffer)
	weapon := rw(buffer)

	fmt.Printf(";   Move time = %d\n", movetime)
	fmt.Printf(";   Position = [%d,%d:%d]\n", pos&31, pos/32, subpos)
	fmt.Printf(";   Direction = %d\n", dir)
	fmt.Printf(";   Type = %d\n", t)
	fmt.Printf(";   Pic = %d\n", pic)
	fmt.Printf(";   Phase = %d\n", phase)
	fmt.Printf(";   Pause = %d\n", pause)
	fmt.Printf(";   Pocket = %d\n", pocket)
	fmt.Printf(";   Weapon = %d\n", weapon)

	fmt.Printf(";}\n")
}

func parseCloseDoor(buffer *bytes.Reader) {
	fmt.Printf("OpenDoor\n;{\n")

	pos := rw(buffer)
	fmt.Printf(";   Position = [%d,%d]\n", pos&31, pos/32)

	fmt.Printf(";}\n")
}

func parseOpenDoor(buffer *bytes.Reader) {
	fmt.Printf("OpenDoor\n;{\n")

	pos := rw(buffer)
	fmt.Printf(";   Position = [%d,%d]\n", pos&31, pos/32)

	fmt.Printf(";}\n")
}

func parseChangeWall(buffer *bytes.Reader) {
	fmt.Printf("ChangeWall\n;{\n")
	forcedSpecialMazeList := make(map[uint16]bool)

	t := rb(buffer)
	if t == 0xf7 {
		fmt.Printf(";   Type = Change all sides\n")
		pos := rw(buffer)
		to := rb(buffer)
		from := rb(buffer)
		forcedSpecialMazeList[pos] = true
		fmt.Printf(";   Position = [%d,%d]\n", pos&31, pos/32)
		fmt.Printf(";   Change from = %d\n", from)
		fmt.Printf(";   Change to = %d\n", to)
	} else if t == 0xe9 {
		fmt.Printf(";   Type = Change one side\n")
		pos := rw(buffer)
		side := rb(buffer)
		to := rb(buffer)
		from := rb(buffer)
		forcedSpecialMazeList[pos] = true
		fmt.Printf(";   Position = [%d,%d]\n", pos&31, pos/32)
		fmt.Printf(";   Side= %d\n", side)
		fmt.Printf(";   Change from = %d\n", from)
		fmt.Printf(";   Change to = %d\n", to)
	} else if t == 0xea {
		fmt.Printf(";   Type = Open door\n")
		pos := rw(buffer)
		forcedSpecialMazeList[pos] = true
		fmt.Printf(";   Position = [%d,%d]\n", pos&31, pos/32)
	}

	fmt.Printf(";}\n")
}

func parseSetWall(buffer *bytes.Reader) {
	fmt.Printf("SetWall\n;{\n")
	forcedSpecialMazeList := make(map[uint16]bool)

	t := rb(buffer)
	if t == 0xf7 {
		fmt.Printf(";   Type = Change all sides\n")
		pos := rw(buffer)
		to := rb(buffer)
		forcedSpecialMazeList[pos] = true
		fmt.Printf(";   Position = [%d,%d]\n", pos&31, pos/32)
		fmt.Printf(";   Change to = %d\n", to)
	} else if t == 0xe9 {
		fmt.Printf(";   Type = Change one side\n")
		pos := rw(buffer)
		side := rb(buffer)
		to := rb(buffer)
		forcedSpecialMazeList[pos] = true
		fmt.Printf(";   Position = [%d,%d]\n", pos&31, pos/32)
		fmt.Printf(";   Side = %d\n", side)
		fmt.Printf(";   Change to = %d\n", to)
	} else if t == 0xed {
		fmt.Printf(";   Type = Change party direction\n")
		direction := rb(buffer)
		fmt.Printf(";   Position = %d\n", direction)
	}

	fmt.Printf(";}\n")
}

func checkTriggerReference(triggers *[]Trigger, address int64) *Trigger {
	for i, trigger := range *triggers {
		if int64(trigger.Address) == address {
			fmt.Printf("\n\n\n; --------------------------------------------------------------------\n")
			fmt.Printf("; Referenced by trigger $%02x. Pos:[%d,%d] Flags: ", i, trigger.Pos.X, trigger.Pos.Y)
			for b := 7; b >= 0; b-- {
				bitSet := (1<<b)&trigger.Flags != 0
				fmt.Printf("%+v ", bitSet)
			}
			fmt.Printf("\n; --------------------------------------------------------------------\n")
			return &trigger
		}
	}
	return nil
}

func parseSetFlag(buffer *bytes.Reader) {
	fmt.Printf("SetFlag\n;{\n")

	target := rb(buffer)
	fmt.Printf(";   Target = ")
	if target == 0xef {
		fmt.Printf("Maze\n")
		flag := rb(buffer)
		fmt.Printf(";   Flag = %d\n", flag)
	} else if target == 0xf0 {
		fmt.Printf("Global\n")
		flag := rb(buffer)
		fmt.Printf(";   Flag = %d\n", flag)
	} else if target == 0xf3 {
		fmt.Printf("Monster\n")
		id := rb(buffer)
		flag := rb(buffer)
		fmt.Printf(";   Monster = %d\n", id)
		fmt.Printf(";   Flag = %d\n", flag)
	} else if target == 0xe4 {
		fmt.Printf("Event\n")
	} else if target == 0xd1 {
		fmt.Printf("Party_Function(FUNC_SETVAL, PARTY_SAVEREST, 0);\n")
	}
	fmt.Printf(";}\n")
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
	rb(buffer)
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
