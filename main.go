package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"sort"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type LayerDataProperty struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type LayerData = struct {
	Data       []float64 `json:"data"`
	Name       string    `json:"name"`
	Visible    bool      `json:"visible"`
	Properties []LayerDataProperty
}

func LayerGetProp(ld LayerData, name string) int {
	for _, p := range ld.Properties {
		if p.Name == name {
			return p.Value
		}

	}
	return -1
}

type TileMapData = struct {
	TileWidth  int         `json:"tilewidth"`
	TileHeight int         `json:"tileheight"`
	Width      int         `json:"width"`
	Height     int         `json:"height"`
	Layers     []LayerData `json:"layers"`
}

func parseMap(filepath string) (TileMapData, error) {
	jsonFile, err := os.Open(filepath)
	if err != nil {
		return TileMapData{}, err
	}
	defer jsonFile.Close()

	var res TileMapData
	buffer, err := io.ReadAll(jsonFile)
	if err != nil {
		return TileMapData{}, err
	}
	err = json.Unmarshal(buffer, &res)
	if err != nil {
		return TileMapData{}, err
	}
	return res, nil
}

type Tile struct {
	Variant int
	Type    string
	Pos     rl.Vector2
}

type Tilemap struct {
	TileLayers   []map[rl.Vector2]Tile
	Objects      []Tile
	Obstacles    map[rl.Vector2]bool
	tilesetAsset rl.Texture2D
	Tilesize     int
	tilesetCols  int
}

func (tm Tilemap) Unload() {
	rl.UnloadTexture(tm.tilesetAsset)
}

func (tm *Tilemap) Draw(offset rl.Vector2, screenSize rl.Vector2) {
	cstartX, cendX := computeCellRange(float64(offset.X), float64(offset.X+screenSize.X), float64(tm.Tilesize))
	cstartY, cendY := computeCellRange(float64(offset.Y), float64(offset.Y+screenSize.Y), float64(tm.Tilesize))
	for _, layers := range tm.TileLayers {
		for y := cstartY; y <= cendY; y++ {
			for x := cstartX; x <= cendX; x++ {
				pos := rl.NewVector2(float32(x), float32(y))
				tile, ok := layers[pos]
				if !ok {
					continue
				}
				viewpos := rl.Vector2Subtract(rl.NewVector2(pos.X*float32(tm.Tilesize), pos.Y*float32(tm.Tilesize)), offset)
				variant := tile.Variant
				cols := tm.tilesetCols

				tx := float32((variant % cols) * tm.Tilesize)
				ty := float32((variant / cols) * tm.Tilesize)
				srcRect := rl.NewRectangle(tx, ty, float32(tm.Tilesize), float32(tm.Tilesize))
				rl.DrawTextureRec(tm.tilesetAsset, srcRect, viewpos, rl.White)
			}
		}

	}
}

func (tm *Tilemap) ExtractObjectOne(obj string) *Tile {
	var idx = -1
	for i, o := range tm.Objects {
		if o.Type == obj {
			idx = i
			break
		}
	}
	if idx >= 0 {
		res := tm.Objects[idx]
		tm.Objects = append(tm.Objects[:idx], tm.Objects[idx+1:]...)
		return &res
	}
	return nil
}

func computeCellRange(start float64, end float64, tilesize float64) (int, int) {
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

func loadTilemap(tmd *TileMapData, tilesize int) Tilemap {
	var img = rl.LoadImage("./resources/map/tilesets.png")
	defer rl.UnloadImage(img)
	scale := tilesize / tmd.TileWidth
	rl.ImageResize(img, img.Width*int32(scale), img.Height*int32(scale))

	var tm Tilemap
	tm.tilesetCols = 64
	tm.TileLayers = []map[rl.Vector2]Tile{}
	tm.Tilesize = tilesize
	tm.tilesetAsset = rl.LoadTextureFromImage(img)
	tm.Obstacles = map[rl.Vector2]bool{}
	tm.Objects = []Tile{}

	var width = tmd.Width
	sort.SliceStable(tmd.Layers, func(i, j int) bool {
		return LayerGetProp(tmd.Layers[i], "z") < LayerGetProp(tmd.Layers[j], "z")
	})

	for _, layer := range tmd.Layers {
		z := LayerGetProp(layer, "z")
		tiles := map[rl.Vector2]Tile{}
		for i, id := range layer.Data {
			x := i % width
			y := i / width
			cellpos := rl.NewVector2(float32(x), float32(y))
			if layer.Name == "obstacles" {
				tm.Obstacles[cellpos] = true
				continue
			}
			if id == 0 {
				continue
			}
			if z == -1 {
				tm.Objects = append(tm.Objects, Tile{Pos: cellpos, Variant: int(id - 1), Type: layer.Name})
			} else {
				tiles[cellpos] = Tile{Type: layer.Name, Variant: int(id - 1), Pos: cellpos}
			}
		}
		if len(tiles) > 0 {
			if z != -1 {
				tm.TileLayers = append(tm.TileLayers, tiles)
			}
		}
	}
	return tm
}

type Player struct {
	Pos           rl.Vector2
	HitAreaOffset rl.Rectangle
	Asset         rl.Texture2D
	AssetSize     rl.Vector2
	TileSize      int
	Size          rl.Vector2
}

func NewPlayer(pos rl.Vector2, tilesize int, scale int) Player {
	playerImg := rl.LoadTexture("./resources/characters/Human/IDLE/base_idle_strip9.png")
	assetSize := rl.NewVector2(float32(playerImg.Width)/9, float32(playerImg.Height))
	size := rl.NewVector2(float32(assetSize.X)*float32(scale), float32(assetSize.Y)*float32(scale))

	hitboxSize := float32(tilesize) * 0.5 * float32(scale)
	hitRect := rl.NewRectangle(size.X/2-hitboxSize/2, size.Y/2-hitboxSize/2, hitboxSize, hitboxSize)
	return Player{
		Pos:           pos,
		HitAreaOffset: hitRect,
		Asset:         playerImg,
		AssetSize:     assetSize,
		TileSize:      tilesize,
		Size:          size,
	}
}

func (p *Player) Unload() {
	rl.UnloadTexture(p.Asset)
}

func (p *Player) Update(dt float32, movement rl.Vector2) {
	p.Pos.X += movement.X * dt * 100
	p.Pos.Y += movement.Y * dt * 100
}

func (p Player) Draw(offset rl.Vector2) {
	rl.DrawRectangleRec(p.Hitbox(offset), rl.Red)
	srcRect := rl.NewRectangle(0, 0, p.AssetSize.X, p.AssetSize.Y)
	destRect := rl.NewRectangle(p.Pos.X-offset.X, p.Pos.Y-offset.Y, p.Size.X, p.Size.Y)
	rl.DrawTexturePro(p.Asset, srcRect, destRect, rl.NewVector2(0, 0), 0, rl.White)
}

func (p Player) Hitbox(offset rl.Vector2) rl.Rectangle {
	return rl.NewRectangle(p.Pos.X+p.HitAreaOffset.X-offset.X, p.Pos.Y+p.HitAreaOffset.Y-offset.Y, p.HitAreaOffset.Width, p.HitAreaOffset.Height)
}

func main() {
	const WIDTH = 1280
	const HEIGHT = 720
	rl.InitWindow(WIDTH, HEIGHT, "Farm sim")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)
	originalTilesize := 16

	tmd, _ := parseMap("./resources/map/0.tmj")
	tm := loadTilemap(&tmd, 32)
	defer tm.Unload()

	playerTile := tm.ExtractObjectOne("player")
	if playerTile == nil {
		return
	}
	startingPlayerPos := rl.NewVector2(playerTile.Pos.X*float32(tm.Tilesize), playerTile.Pos.Y*float32(tm.Tilesize))
	fmt.Printf("%v, %v\n", playerTile.Pos, startingPlayerPos)
	player := NewPlayer(
		startingPlayerPos,
		tm.Tilesize,
		tm.Tilesize/originalTilesize,
	)

	var camScroll = rl.NewVector2(0, 0)
	for !rl.WindowShouldClose() {
		playerMoveX := []float32{0, 0}
		playerMoveY := []float32{0, 0}
		dt := rl.GetFrameTime()
		if rl.IsKeyDown(rl.KeyUp) {
			playerMoveY[0] = 1
		}
		if rl.IsKeyDown(rl.KeyDown) {
			playerMoveY[1] = 1
		}
		if rl.IsKeyDown(rl.KeyLeft) {
			playerMoveX[0] = 1
		}
		if rl.IsKeyDown(rl.KeyRight) {
			playerMoveX[1] = 1
		}

		if rl.IsKeyUp(rl.KeyUp) {
			playerMoveY[0] = 0
		}
		if rl.IsKeyUp(rl.KeyDown) {
			playerMoveY[1] = 0
		}
		if rl.IsKeyUp(rl.KeyLeft) {
			playerMoveX[0] = 0
		}
		if rl.IsKeyUp(rl.KeyRight) {
			playerMoveX[1] = 0
		}

		camScrollDest := rl.NewVector2(player.Pos.X-WIDTH/2, player.Pos.Y-HEIGHT/2)
		dCamScroll := rl.NewVector2((camScrollDest.X-camScroll.X)*2, (camScrollDest.Y-camScroll.Y)*2)

		camScroll.X += dCamScroll.X * dt
		camScroll.Y += dCamScroll.Y * dt
		player.Update(dt, rl.NewVector2(playerMoveX[1]-playerMoveX[0], playerMoveY[1]-playerMoveY[0]))

		rl.BeginDrawing()
		rl.ClearBackground(rl.White)
		tm.Draw(camScroll, rl.NewVector2(WIDTH, HEIGHT))
		player.Draw(camScroll)
		rl.EndDrawing()
	}
}
