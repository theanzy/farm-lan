package main

import (
	"encoding/json"
	"io"
	"log"
	"math"
	"os"
	"path"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

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

type AnimStyle = struct {
	Variants   map[string]rl.Texture2D
	Base       rl.Texture2D
	StripCount int
}
type AnimStyles = map[string]AnimStyle

func NewAnimStyles(dir string) AnimStyles {
	styles := AnimStyles{}

	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	r := regexp.MustCompile(`[a-z](\d+)\.png`)
	supportedStyles := []string{"IDLE", "WALKING"}

	for _, e := range entries {
		if !slices.Contains(supportedStyles, e.Name()) {
			continue
		}
		var style = AnimStyle{
			Variants:   map[string]rl.Texture2D{},
			StripCount: 0,
		}
		fullpath := path.Join(dir, e.Name())
		files, err := os.ReadDir(fullpath)
		if err != nil {
			log.Fatal(err)
		}

		for _, f := range files {
			variantName := strings.Split(f.Name(), "_")[0]
			s := r.FindStringSubmatch(f.Name())[1]
			strip, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				log.Fatal(err)
			}
			style.StripCount = int(strip)

			imgPath := path.Join(fullpath, f.Name())
			if variantName == "base" {
				style.Base = rl.LoadTexture(imgPath)
			} else {
				style.Variants[variantName] = rl.LoadTexture(imgPath)
			}
		}
		styles[e.Name()] = style
	}
	return styles
}
func UnloadAnimStyles(s AnimStyles) {
	for _, style := range s {
		rl.UnloadTexture(style.Base)
		for _, variant := range style.Variants {
			rl.UnloadTexture(variant)
		}
	}
}

type Animation struct {
	AssetSize  rl.Vector2
	Image      rl.Texture2D
	X          float32
	Speed      float32
	StripCount float32
}

func (a *Animation) Update(dt float32) {
	a.X += dt * a.Speed
	if a.X >= a.StripCount {
		a.X = 0
	}
}

func (a Animation) SrcRect() rl.Rectangle {
	x := float32(math.Floor(float64(a.X)))
	return rl.NewRectangle(x*a.AssetSize.X, 0, a.AssetSize.X, a.AssetSize.Y)
}

type Player struct {
	Pos            rl.Vector2
	HitAreaOffset  rl.Rectangle
	AssetSize      rl.Vector2
	TileSize       int
	Size           rl.Vector2
	AnimStyles     AnimStyles
	AnimState      string
	BaseAnimations map[string]Animation
}

func NewPlayer(pos rl.Vector2, tilesize int, scale int, animStyles AnimStyles) Player {
	playerImg := animStyles["IDLE"].Base
	stripCount := animStyles["IDLE"].StripCount

	assetSize := rl.NewVector2(float32(playerImg.Width)/float32(stripCount), float32(playerImg.Height))
	size := rl.NewVector2(float32(assetSize.X)*float32(scale), float32(assetSize.Y)*float32(scale))

	hitboxSize := assetSize.X * 0.4

	hitRect := rl.NewRectangle(size.X/2-hitboxSize/2, size.Y/2-hitboxSize/2, hitboxSize, hitboxSize)

	baseAnimations := map[string]Animation{}
	for anim, animStyle := range animStyles {
		baseAnimations[anim] = Animation{Image: animStyle.Base, AssetSize: assetSize, X: 0, Speed: 20, StripCount: float32(animStyle.StripCount)}
	}

	return Player{
		Pos:            pos,
		HitAreaOffset:  hitRect,
		AssetSize:      assetSize,
		Size:           size,
		TileSize:       tilesize,
		AnimStyles:     animStyles,
		AnimState:      "IDLE",
		BaseAnimations: baseAnimations,
	}
}

func (p *Player) Update(dt float32, movement rl.Vector2) {
	frameMovement := rl.Vector2Normalize(movement)
	p.Pos.X += frameMovement.X * dt * 150
	p.Pos.Y += frameMovement.Y * dt * 150

	if movement.Y > 0 {
		p.AnimState = "WALKING"
	} else if movement.Y < 0 {
		p.AnimState = "WALKING"
	} else if movement.X > 0 {
		// TODO flip
		p.AnimState = "WALKING"
	} else if movement.X < 0 {
		// TODO flip
		p.AnimState = "WALKING"
	} else {
		p.AnimState = "IDLE"
	}

	baseAnim := p.BaseAnimations[p.AnimState]
	baseAnim.Update(dt)
	p.BaseAnimations[p.AnimState] = baseAnim
}

func (p Player) Draw(offset rl.Vector2) {
	rl.DrawRectangleRec(p.Hitbox(offset), rl.Red)
	destRect := rl.NewRectangle(p.Pos.X-offset.X, p.Pos.Y-offset.Y, p.Size.X, p.Size.Y)
	baseAnim := p.BaseAnimations[p.AnimState]
	rl.DrawTexturePro(baseAnim.Image, baseAnim.SrcRect(), destRect, rl.NewVector2(0, 0), 0, rl.White)
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
	tm := loadTilemap(&tmd, 48)
	defer tm.Unload()

	humanAnimStyles := NewAnimStyles("./resources/characters/Human")
	defer UnloadAnimStyles(humanAnimStyles)

	playerTile := tm.ExtractObjectOne("player")
	if playerTile == nil {
		return
	}
	startingPlayerPos := rl.NewVector2(playerTile.Pos.X*float32(tm.Tilesize), playerTile.Pos.Y*float32(tm.Tilesize))
	player := NewPlayer(
		startingPlayerPos,
		tm.Tilesize,
		tm.Tilesize/originalTilesize,
		humanAnimStyles,
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
