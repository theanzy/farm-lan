package world

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func GetCellPos(pos rl.Vector2, tilesize float64) rl.Vector2 {
	return rl.NewVector2(float32(math.Floor(float64(pos.X)/tilesize)), float32(math.Floor(float64(pos.Y)/tilesize)))
}

var NEIGHBOR_OFFSET = []rl.Vector2{
	rl.NewVector2(-1, -1),
	rl.NewVector2(0, -1),
	rl.NewVector2(1, -1),
	rl.NewVector2(-1, 0),
	rl.NewVector2(1, 0),
	rl.NewVector2(-1, 1),
	rl.NewVector2(0, 1),
	rl.NewVector2(1, 1),
	rl.NewVector2(0, 0),
}

func GetTileRectsAround[V any](tiles map[rl.Vector2]V, pos rl.Vector2, tilesize float32) []rl.Rectangle {
	var res = []rl.Rectangle{}
	cellPos := GetCellPos(pos, float64(tilesize))
	for _, offset := range NEIGHBOR_OFFSET {
		neighborPos := rl.NewVector2(cellPos.X+offset.X, cellPos.Y+offset.Y)
		if _, ok := tiles[neighborPos]; ok {
			res = append(
				res,
				rl.NewRectangle(
					neighborPos.X*tilesize,
					neighborPos.Y*tilesize,
					tilesize,
					tilesize,
				),
			)
		}
	}
	return res
}

func ComputeCellRange(start float64, end float64, tilesize float64) (int, int) {
	var sc = math.Floor(start / tilesize)
	var startC = int(sc)
	if sc < 0 {
		startC -= 1
	}
	var ec = math.Floor(end / tilesize)
	var endC = int(ec) + 1
	if ec < 0 {
		endC -= 1
	}
	return startC, endC
}
