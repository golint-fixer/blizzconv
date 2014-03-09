// Package dun implements functionality for parsing DUN files.
//
// DUN files contain information about how to arrange the squares, which are
// constructed based on the TIL format, in order to form a dungeon. Below is a
// description of the DUN format:
//
// DUN format:
//    dunQWidth       uint16
//    dunQHeight      uint16
//    squareNumsPlus1 [dunQWidth][dunQHeight]uint16
//    // dunWidth  = 2*dunQWidth
//    // dunHeight = 2*dunQHeight
//    unknown         [dunWidth][dunHeight]uint16
//    dunMonsterIDs   [dunWidth][dunHeight]uint16
//    dunObjectIDs    [dunWidth][dunHeight]uint16
//    transparencies  [dunWidth][dunHeight]uint16
package dun

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/mewrnd/blizzconv/configs/dunconf"
	"github.com/mewrnd/blizzconv/configs/til"
	"github.com/mewrnd/blizzconv/mpq"
)

// The maximum number of cols and rows in a dungeon map.
const (
	ColMax = 112
	RowMax = 112
)

// A Dungeon maps from a col and a row to the dungeon information about a cell,
// such as its pillarNum.
//
// The valid keys are:
//    "pillarNum"
//    "unknown" // TODO: update this key once known.
//    "dunMonstersIDs"
//    "dunObjectIDs"
//    "transparencies"
type Dungeon [ColMax][RowMax]map[string]int

// objects maps from object idx to object names.
var objects = []string{
	0:   "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	1:   "Lever (position a)",               // lever (frame 0)
	2:   "Crucified Skeleton (south)",       // cruxsk1 (frame 0)
	3:   "Crucified Skeleton (south east)",  // cruxsk2 (frame 0)
	4:   "Crucified Skeleton (south west)",  // cruxsk3 (frame 0)
	5:   "Angel",                            // angel (frame 0)
	6:   "Banner (south east, theme 3)",     // banner (frame 1)
	7:   "Banner (theme 3)",                 // banner (frame 0)
	8:   "Banner (south west, theme 3)",     // banner (frame 2)
	9:   "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	10:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	11:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	12:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	13:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	14:  "Ancient Tome or Book of Vileness", // book2 (frame 0)
	15:  "Mythical Book",                    // book2 (frame 3)
	16:  "Burning Cross",                    // burncros (animated, ticksPerFrame 0)
	17:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	18:  "Invalid 1",                        // l1braz (invalid frame)
	19:  "Candle (theme 1)",                 // candle2 (animated, ticksPerFrame 2)
	20:  "Invalid 2",                        // l1braz (invalid frame)
	21:  "Cauldron",                         // cauldren (frame 0)
	22:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	23:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	24:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	25:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	26:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	27:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	28:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	29:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	30:  "Flame",                            // flame1 (frame 0)
	31:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	32:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	33:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	34:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	35:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	36:  "Magic Circle Pentagram",           // mcirl (frame 0)
	37:  "Magic Circle",                     // mcirl (frame 0) [frame 2 in game]
	38:  "Skull Fire (theme 3)",             // skulfire (animated, ticksPerFrame 2)
	39:  "Skulpile",                         // skulpile (invalid frame)
	40:  "Invalid 3",                        // l1braz (invalid frame)
	41:  "Invalid 4",                        // l1braz (invalid frame)
	42:  "Invalid 5",                        // l1braz (invalid frame)
	43:  "Invalid 6",                        // l1braz (invalid frame)
	44:  "Invalid 7",                        // l1braz (invalid frame)
	45:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	46:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	47:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	48:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	49:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	50:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	51:  "Skull Lever",                      // switch4 (frame 0)
	52:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	53:  "Traphole (south west)",            // traphole (frame 0)
	54:  "Traphole (south east)",            // traphole (frame 1)
	55:  "Tortured Soul 0",                  // tsoul (frame 0)
	56:  "Tortured Soul 1",                  // tsoul (frame 1)
	57:  "Tortured Soul 2",                  // tsoul (frame 2)
	58:  "Tortured Soul 3",                  // tsoul (frame 3)
	59:  "Tortured Soul 4",                  // tsoul (frame 4)
	60:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	61:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	62:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	63:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	64:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	65:  "Nude",                             // nude2 (animated, ticksPerFrame 3)
	66:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	67:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	68:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	69:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	70:  "Tortured Nude Man 0",              // tnudem (frame 0)
	71:  "Tortured Nude Man 1 (theme 6)",    // tnudem (frame 1)
	72:  "Tortured Nude Man 2 (theme 6)",    // tnudem (frame 2)
	73:  "Tortured Nude Man 3 (theme 6)",    // tnudem (frame 3)
	74:  "Tortured Nude Woman 0 (theme 6)",  // tnudew (frame 0)
	75:  "Tortured Nude Woman 1 (theme 6)",  // tnudew (frame 1)
	76:  "Tortured Nude Woman 2 (theme 6)",  // tnudew (frame 2)
	77:  "Small Chest",                      // chest1 (frame 0)
	78:  "Small Chest",                      // chest1 (frame 0)
	79:  "Small Chest",                      // chest1 (frame 0)
	80:  "Chest",                            // chest2 (frame 0)
	81:  "Chest",                            // chest2 (frame 0)
	82:  "Chest",                            // chest2 (frame 0)
	83:  "Large Chest",                      // chest3 (frame 0)
	84:  "Large Chest",                      // chest3 (frame 0)
	85:  "Large Chest",                      // chest3 (frame 0)
	86:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	87:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	88:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	89:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	90:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	91:  "Pedestal of Blood",                // pedistl (frame 0)
	92:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	93:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	94:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	95:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	96:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	97:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	98:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	99:  "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	100: "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	101: "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	102: "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	103: "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	104: "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	105: "Altar Boy",                        // altboy (frame 0)
	106: "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	107: "Brazier",                          // l1braz (animated, ticksPerFrame 1)
	108: "Armor Stand (Warlord of Blood)",   // armstand (frame 0)
	109: "Weapon Rack (Warlord of Blood)",   // weapstnd (frame 0)
	110: "Wall Torch (south east)",          // wtorch2 (animated, ticksPerFrame 1)
	111: "Wall Torch (south west)",          // wtorch1 (animated, ticksPerFrame 1)
	112: "Mushroom Patch",                   // mushptch (frame 0)
	113: "Brazier",                          // l1braz (animated, ticksPerFrame 1)
}

// New returns a new Dungeon.
func New() (dungeon *Dungeon) {
	dungeon = new(Dungeon)
	for row := 0; row < RowMax; row++ {
		for col := 0; col < ColMax; col++ {
			dungeon[col][row] = make(map[string]int)
		}
	}
	return dungeon
}

// Parse parses a given DUN file and stores each pillarNum at a coordinate in
// the dungeon, based on the DUN format described above.
//
// Below is a description of how the squares are positioned on the dungeon map:
//    1) Start at the coordinates colStart, rowStart.
//    2) Place a square.
//       - Each square is two cols in width and two rows in height.
//    3) Increment col with two.
//    4) goto 2) dunQWidth number of times.
//    5) Increment row with two.
//    6) goto 2) dunQHeight number of times.
//
// ref: GetPillarRect (illustration of map coordinate system)
//
// Any additional cell data is stored afterwards using row major.
func (dungeon *Dungeon) Parse(dunName string) (err error) {
	dunPath, err := mpq.GetPath(dunName)
	if err != nil {
		return err
	}
	fr, err := os.Open(dunPath)
	if err != nil {
		return err
	}
	defer fr.Close()
	var tmp [2]uint16
	err = binary.Read(fr, binary.LittleEndian, &tmp)
	if err != nil {
		return err
	}
	dunQWidth := int(tmp[0])
	dunQHeight := int(tmp[1])
	colStart, err := dunconf.GetColStart(dunName)
	if err != nil {
		return err
	}
	rowStart, err := dunconf.GetRowStart(dunName)
	if err != nil {
		return err
	}
	nameWithoutExt, err := GetLevelName(dunName)
	if err != nil {
		return err
	}

	// squareNumsPlus1.
	squares, err := til.Parse(nameWithoutExt + ".til")
	if err != nil {
		return err
	}
	row := rowStart
	for i := 0; i < dunQHeight; i++ {
		col := colStart
		for j := 0; j < dunQWidth; j++ {
			var x uint16
			err = binary.Read(fr, binary.LittleEndian, &x)
			if err != nil {
				return err
			}
			squareNumPlus1 := int(x)
			if squareNumPlus1 != 0 {
				square := squares[squareNumPlus1-1]
				dungeon[col][row]["pillarNum"] = square.PillarNumTop
				dungeon[col+1][row]["pillarNum"] = square.PillarNumRight
				dungeon[col][row+1]["pillarNum"] = square.PillarNumLeft
				dungeon[col+1][row+1]["pillarNum"] = square.PillarNumBottom
			}
			col += 2
		}
		row += 2
	}

	dunWidth := 2 * dunQWidth
	dunHeight := 2 * dunQHeight

	// TODO: Figure out what these values are used for. Items?
	row = rowStart
	for i := 0; i < dunHeight; i++ {
		col := colStart
		for j := 0; j < dunWidth; j++ {
			var x uint16
			err = binary.Read(fr, binary.LittleEndian, &x)
			if err != nil {
				// Some DUN files only contain the pillar IDs.
				if err == io.EOF && i == 0 && j == 0 {
					return nil
				}
				return err
			}
			dungeon[col][row]["unknown"] = int(x)
			col++
		}
		row++
	}

	// dunMonsterIDs.
	row = rowStart
	for i := 0; i < dunHeight; i++ {
		col := colStart
		for j := 0; j < dunWidth; j++ {
			var x uint16
			err = binary.Read(fr, binary.LittleEndian, &x)
			if err != nil {
				if err == io.EOF && i == 0 && j == 0 {
					return nil
				}
				return err
			}
			// TODO: Lookup monster idx from dunMonsterID.
			// ref: 4B6C98
			dungeon[col][row]["dunMonsterID"] = int(x)
			col++
		}
		row++
	}

	// dunObjectIDs.
	row = rowStart
	for i := 0; i < dunHeight; i++ {
		col := colStart
		for j := 0; j < dunWidth; j++ {
			var x uint16
			err = binary.Read(fr, binary.LittleEndian, &x)
			if err != nil {
				if err == io.EOF && i == 0 && j == 0 {
					return nil
				}
				return err
			}
			// TODO: Lookup object idx from dunObjectID.
			// ref: 4AAD28
			dungeon[col][row]["dunObjectID"] = int(x)
			col++
		}
		row++
	}

	// transparencies.
	row = rowStart
	for i := 0; i < dunHeight; i++ {
		col := colStart
		for j := 0; j < dunWidth; j++ {
			var x uint16
			err = binary.Read(fr, binary.LittleEndian, &x)
			if err != nil {
				if err == io.EOF && i == 0 && j == 0 {
					return nil
				}
				return err
			}
			dungeon[col][row]["transparency"] = int(x)
			col++
		}
		row++
	}

	return nil
}

// GetLevelName returns the level name (without extension) of a given DUN file.
func GetLevelName(dunName string) (nameWithoutExt string, err error) {
	relDunPath, err := mpq.GetRelPath(dunName)
	if err != nil {
		return "", err
	}
	dunDir, _ := path.Split(relDunPath)
	switch dunDir {
	case "levels/l1data/":
		nameWithoutExt = "l1"
	case "levels/l2data/":
		nameWithoutExt = "l2"
	case "levels/l3data/":
		nameWithoutExt = "l3"
	case "levels/l4data/":
		nameWithoutExt = "l4"
	case "levels/towndata/":
		nameWithoutExt = "town"
	default:
		return "", fmt.Errorf("invalid dunDir (%s).", dunDir)
	}
	return nameWithoutExt, nil
}
