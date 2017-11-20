package dunmini

import (
	"image"
	"image/draw"
	"log"

	"github.com/mewrnd/blizzconv/configs/min"
	"github.com/mewrnd/blizzconv/images/cel"
)

var arches []image.Image

// Image returns an image constructed from the pillars associated with each
// coordinate of the dungeon map.
//
// ref: GetPillarRect (illustration of map coordinate system)
func (dungeon *Dungeon) Image(colCount, rowCount int, pillars []min.Pillar, levelFrames []image.Image) (img image.Image) {
	if arches == nil {
		conf, err := cel.GetConf("l1s.cel", "levels/l1data/l1.pal")
		if err != nil {
			log.Fatalln(err)
		}
		arches, err = cel.DecodeAll("l1s.cel", conf)
		if err != nil {
			log.Fatalln(err)
		}
	}
	pillarHeight := pillars[0].Height()
	maxCount := colCount
	if rowCount > maxCount {
		maxCount = rowCount
	}
	// TODO: Fix this logic, the placement of squares is sometimes a bit off.
	mapWidth := maxCount*min.BlockWidth + maxCount*min.BlockWidth
	mapHeight := maxCount*(min.BlockHeight/2) + maxCount*(min.BlockHeight/2) + (pillarHeight - min.BlockHeight)
	//mapWidth := colCount*min.BlockWidth + rowCount*min.BlockWidth
	//mapHeight := colCount*(min.BlockHeight/2) + rowCount*(min.BlockHeight/2) + (pillarHeight - min.BlockHeight)
	dst := image.NewRGBA(image.Rect(0, 0, mapWidth, mapHeight))
	for row := 0; row < rowCount; row++ {
		for col := 0; col < colCount; col++ {
			pillarNum, ok := dungeon[col][row]["pillarNum"]
			if ok {
				rect := GetPillarRect(col, row, mapWidth, pillarHeight)
				src := pillars[pillarNum].Image(levelFrames)
				draw.Draw(dst, rect, src, image.ZP, draw.Over)
				archID, ok := getArchID(pillarNum)
				if ok {
					draw.Draw(dst, rect, arches[archID], image.ZP, draw.Over)
				}
			}
		}
	}
	return dst
}

// GetPillarRect returns an image.Rectangle based on the col and row
// coordinates. The calculations are based on the map coordinate system
// illustrated below:
//
// Map coordinate system:
//                 (0, 0)
//
//                   /\
//                r /\/\ c
//               o /\/\/\ o
//              w /\/\/\/\ l
//               /\/\/\/\/\
//    (0, 111)   \/\/\/\/\/   (111, 0)
//                \/\/\/\/
//                 \/\/\/
//                  \/\/
//                   \/
//
//               (111, 111)
func GetPillarRect(col, row, mapWidth, pillarHeight int) (rect image.Rectangle) {
	minX := mapWidth/2 - min.BlockWidth - row*min.BlockWidth + col*min.BlockWidth
	minY := row*(min.BlockHeight/2) + col*(min.BlockHeight/2)
	maxX := minX + min.PillarWidth
	maxY := minY + pillarHeight
	return image.Rect(minX, minY, maxX, maxY)
}

// getArchID returns the arch ID of the provided pillarID and true, or 0 and
// false if there is no arch associated with the provided pillar ID.
func getArchID(pillarID int) (archID int, ok bool) {
	switch pillarID {
	case PillarIDFloorShadowArchSw_1, PillarIDFloorShadowArchSw_2, PillarIDFloorShadowArchSw_3, PillarIDFloorShadowArchSw_4, PillarIDFloorShadowArchSw_5, PillarIDFloorShadowArchSw_6:
		return ArchSw, true
	case PillarIDFloorShadowArchSe_1, PillarIDFloorShadowArchSe_2, PillarIDFloorShadowArchSe_3, PillarIDFloorShadowArchSe_4, PillarIDFloorShadowArchSe_5, PillarIDFloorShadowArchSe_6:
		return ArchSe, true
	case PillarIDFloorShadowArchSwBroken2_1:
		return ArchSwBroken2, true
	case PillarIDFloorShadowArchSw2_1:
		return ArchSw2, true
	}
	return 0, false
}

// Pillar ids for layout 1.
const (
	// Floor shadows for arches.
	PillarIDFloorShadowArchSe_1        = 10
	PillarIDFloorShadowArchSe_2        = 248
	PillarIDFloorShadowArchSe_3        = 324
	PillarIDFloorShadowArchSe_4        = 330
	PillarIDFloorShadowArchSe_5        = 343
	PillarIDFloorShadowArchSe_6        = 420
	PillarIDFloorShadowArchSw2_1       = 258
	PillarIDFloorShadowArchSwBroken2_1 = 254
	PillarIDFloorShadowArchSw_1        = 11
	PillarIDFloorShadowArchSw_2        = 70
	PillarIDFloorShadowArchSw_3        = 210
	PillarIDFloorShadowArchSw_4        = 320
	PillarIDFloorShadowArchSw_5        = 340
	PillarIDFloorShadowArchSw_6        = 417
)

// Arch ids for layout 1.
const (
	ArchSe        = 1
	ArchSeBroken  = 2
	ArchSeDoor    = 7
	ArchSw        = 0
	ArchSw2       = 4
	ArchSwBroken  = 5
	ArchSwBroken2 = 3
	ArchSwDoor    = 6
)
