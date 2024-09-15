package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path"
	"path/filepath"
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

func (t Tile) Center(tilesize float32) rl.Vector2 {
	return rl.NewVector2(t.Pos.X+tilesize, t.Pos.Y)
}

type Tilemap struct {
	TileLayers   []map[rl.Vector2]Tile
	Objects      []Tile
	Obstacles    map[rl.Vector2]bool
	tilesetAsset rl.Texture2D
	Tilesize     int
	tilesetCols  int
	Roofs        []Tile
	FarmTiles    map[rl.Vector2]FarmTile
	TileScale    int
	CropAssets   map[string][]rl.Texture2D
}
type FarmTile struct {
	Pos rl.Vector2
	// empty, digged, planted
	State    string
	PlantAge int
}

func (tm Tilemap) Unload() {
	rl.UnloadTexture(tm.tilesetAsset)
}

func (tm *Tilemap) DrawTerrain(offset rl.Vector2, screenSize rl.Vector2) {
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
				tm.DrawTile(tile, offset)
			}
		}
	}
	for _, ft := range tm.FarmTiles {
		if ft.State == "digged" {

			cellpos := ft.Pos
			tex := tm.CropAssets["soil"][0]
			tilesize := float32(tm.Tilesize)
			viewpos := rl.Vector2Subtract(
				rl.NewVector2(
					cellpos.X*tilesize+tilesize*0.5-float32(tex.Width)*float32(tm.TileScale)*0.5,
					cellpos.Y*tilesize+tilesize*0.5-float32(tex.Height)*float32(tm.TileScale)*0.5,
				),
				offset,
			)
			rl.DrawTextureEx(tex, viewpos, 0, float32(tm.TileScale), rl.White)
			// rl.DrawRectangleV(viewpos, rl.NewVector2(float32(tm.Tilesize), float32(tm.Tilesize)), rl.NewColor(150, 81, 9, 50))
		}
	}

	// Draw obstacles
	// for obstaclePos := range tm.Obstacles {
	// 	rl.DrawRectangle(int32(obstaclePos.X*float32(tm.Tilesize)-offset.X), int32(obstaclePos.Y*float32(tm.Tilesize)-offset.Y), int32(tm.Tilesize), int32(tm.Tilesize), rl.White)
	// }
}

func (tm *Tilemap) DrawRoof(offset rl.Vector2) {
	for _, obj := range tm.Roofs {
		if obj.Type == "house_roof_float_front" || obj.Type == "house_roof_float" {
			continue
		}
		tm.DrawTile(obj, offset)
	}
}

func (tm *Tilemap) DrawTile(tile Tile, offset rl.Vector2) {
	cols := tm.tilesetCols
	tx := float32((tile.Variant % cols) * tm.Tilesize)
	ty := float32((tile.Variant / cols) * tm.Tilesize)
	srcRect := rl.NewRectangle(tx, ty, float32(tm.Tilesize), float32(tm.Tilesize))

	cellpos := tile.Pos
	viewpos := rl.Vector2Subtract(rl.NewVector2(cellpos.X*float32(tm.Tilesize), cellpos.Y*float32(tm.Tilesize)), offset)
	rl.DrawTextureRec(tm.tilesetAsset, srcRect, viewpos, rl.White)
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

func (tm *Tilemap) ExtractObjects(objs []string) []Tile {
	tiles := []Tile{}
	for i := 0; i < len(tm.Objects); i++ {
		o := tm.Objects[i]
		if slices.Contains(objs, o.Type) {
			tm.Objects = append(tm.Objects[:i], tm.Objects[i+1:]...)
			i-- // Since we just deleted a[i], we must redo that index
			tiles = append(tiles, o)
		}
	}
	return tiles
}

func (tm *Tilemap) GetObstaclesAround(pos rl.Vector2) []rl.Rectangle {
	return GetTileRectsAround(tm.Obstacles, pos, float32(tm.Tilesize))
}
func (tm *Tilemap) GetFarmRectsAround(pos rl.Vector2) []rl.Rectangle {
	return GetTileRectsAround(tm.FarmTiles, pos, float32(tm.Tilesize))
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

func (tm *Tilemap) AddFarmHole(pos rl.Vector2) {
	cellpos := GetCellPos(pos, float64(tm.Tilesize))
	if t, ok := tm.FarmTiles[cellpos]; ok {
		t.State = "digged"
		tm.FarmTiles[cellpos] = t
	}
}

func GetCellPos(pos rl.Vector2, tilesize float64) rl.Vector2 {
	return rl.NewVector2(float32(math.Floor(float64(pos.X)/tilesize)), float32(math.Floor(float64(pos.Y)/tilesize)))
}

func (tm *Tilemap) GetFloatingRoofs() []Tile {
	return tm.GetTiles(tm.Roofs, []string{"house_roof_float", "house_roof_float_front"})
}

func (tm *Tilemap) GetTiles(tiles []Tile, types []string) []Tile {
	res := []Tile{}
	for _, s := range tiles {
		if slices.Contains(types, s.Type) {
			res = append(res, s)
		}
	}
	return res
}

func loadTilemap(tmd *TileMapData, cropAssets map[string][]rl.Texture2D, tilesize int) Tilemap {
	var img = rl.LoadImage("./resources/map/tilesets.png")
	defer rl.UnloadImage(img)
	scale := tilesize / tmd.TileWidth
	rl.ImageResizeNN(img, img.Width*int32(scale), img.Height*int32(scale))

	var tm Tilemap
	tm.tilesetCols = 64
	tm.TileLayers = []map[rl.Vector2]Tile{}
	tm.Tilesize = tilesize
	tm.tilesetAsset = rl.LoadTextureFromImage(img)
	tm.Obstacles = map[rl.Vector2]bool{}
	tm.Objects = []Tile{}
	tm.FarmTiles = map[rl.Vector2]FarmTile{}
	tm.TileScale = scale
	tm.CropAssets = cropAssets
	fmt.Println(tm.CropAssets)

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
			if layer.Name == "obstacles" && id > 0 {
				tm.Obstacles[cellpos] = true
				continue
			}
			if layer.Name == "farm_tile" && id > 0 {
				tm.FarmTiles[cellpos] = FarmTile{
					Pos:      cellpos,
					State:    "empty",
					PlantAge: 0,
				}
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
	sort.SliceStable(tm.Objects, func(i, j int) bool {
		return tm.Objects[i].Center(float32(tm.Tilesize)).Y < tm.Objects[j].Center(float32(tm.Tilesize)).Y
	})

	houseRoofs := []string{"house_roof_float", "house_roof_float_front", "house_roof", "house_roof_front"}
	tm.Roofs = tm.ExtractObjects(houseRoofs)
	sort.SliceStable(tm.Roofs, func(i, j int) bool {
		if tm.Roofs[i].Type == tm.Roofs[j].Type {
			return tm.Roofs[i].Center(float32(tm.Tilesize)).Y < tm.Roofs[j].Center(float32(tm.Tilesize)).Y
		}
		return slices.Index(houseRoofs, tm.Roofs[i].Type) < slices.Index(houseRoofs, tm.Roofs[i].Type)
	})

	return tm
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

	supportedStyles := []string{"IDLE", "WALKING", "WATERING", "DIG", "AXE"}

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

func FlipTextureHorizontal(tex rl.Texture2D) rl.Texture2D {
	baseImg := rl.LoadImageFromTexture(tex)
	defer rl.UnloadImage(baseImg)
	rl.ImageFlipHorizontal(baseImg)
	return rl.LoadTextureFromImage(baseImg)
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
	AssetSize    rl.Vector2
	Image        rl.Texture2D
	ImageFlipped rl.Texture2D
	X            float32
	Speed        float32
	StripCount   float32
}

func (a *Animation) Update(dt float32) {
	a.X += dt * a.Speed
	if a.X >= a.StripCount {
		a.X = 0
	}
}
func (a *Animation) Reset() {
	a.X = 0
}

func (a Animation) SrcRect() rl.Rectangle {
	x := float32(math.Floor(float64(a.X)))
	return rl.NewRectangle(x*a.AssetSize.X, 0, a.AssetSize.X, a.AssetSize.Y)
}

type Player struct {
	Pos               rl.Vector2
	HitAreaOffset     rl.Rectangle
	AssetSize         rl.Vector2
	TileSize          int
	Size              rl.Vector2
	AnimStyles        AnimStyles
	AnimStylesFlipped AnimStyles
	AnimState         string
	BaseAnimations    map[string]Animation
	ToolAnimations    map[string]Animation
	StyleAnimations   map[string]Animation
	Flipped           bool
	Tool              string
	Tools             []string
	ToolCounter       float32
}

func NewPlayer(pos rl.Vector2, tilesize int, scale int, animStyles AnimStyles, animStylesFlipped AnimStyles, tools []string, style string) Player {
	playerImg := animStyles["IDLE"].Base
	stripCount := animStyles["IDLE"].StripCount

	assetSize := rl.NewVector2(float32(playerImg.Width)/float32(stripCount), float32(playerImg.Height))
	size := rl.NewVector2(float32(assetSize.X)*float32(scale), float32(assetSize.Y)*float32(scale))

	hitboxSize := assetSize.X * 0.4

	hitRect := rl.NewRectangle(size.X/2-hitboxSize/2, size.Y/2-hitboxSize/2, hitboxSize, hitboxSize)

	baseAnimations := map[string]Animation{}
	toolAnimations := map[string]Animation{}
	styleAnimations := map[string]Animation{}
	animSpeed := 12
	for anim, animStyle := range animStyles {
		baseAnimations[anim] = Animation{
			Image:        animStyle.Base,
			AssetSize:    assetSize,
			X:            0,
			Speed:        float32(animSpeed),
			StripCount:   float32(animStyle.StripCount),
			ImageFlipped: animStylesFlipped[anim].Base,
		}

		baseAnimations[anim] = Animation{
			Image:        animStyle.Base,
			AssetSize:    assetSize,
			X:            0,
			Speed:        float32(animSpeed),
			StripCount:   float32(animStyle.StripCount),
			ImageFlipped: animStylesFlipped[anim].Base,
		}
		if _, ok := animStyle.Variants[style]; ok {
			styleAnimations[anim] = Animation{
				Image:        animStyle.Variants[style],
				AssetSize:    assetSize,
				X:            0,
				Speed:        float32(animSpeed),
				StripCount:   float32(animStyle.StripCount),
				ImageFlipped: animStylesFlipped[anim].Variants[style],
			}
		}
		if _, ok := animStyle.Variants["tools"]; ok {
			toolAnimations[anim] = Animation{
				Image:        animStyle.Variants["tools"],
				AssetSize:    assetSize,
				X:            0,
				Speed:        float32(animSpeed),
				StripCount:   float32(animStyle.StripCount),
				ImageFlipped: animStylesFlipped[anim].Variants["tools"],
			}
		}
	}

	return Player{
		Pos:               pos,
		HitAreaOffset:     hitRect,
		AssetSize:         assetSize,
		Size:              size,
		TileSize:          tilesize,
		AnimStyles:        animStyles,
		AnimStylesFlipped: animStylesFlipped,
		AnimState:         "IDLE",
		BaseAnimations:    baseAnimations,
		ToolAnimations:    toolAnimations,
		StyleAnimations:   styleAnimations,
		Flipped:           false,
		Tool:              "water",
		Tools:             tools,
		ToolCounter:       0,
	}
}

func (p *Player) SwitchTool() {
	if p.ToolCounter > 0 {
		return
	}
	idx := slices.Index(p.Tools, p.Tool)
	idx += 1
	if idx >= len(p.Tools) {
		idx = 0
	}
	p.Tool = p.Tools[idx]
}

func (p *Player) UseTool() {
	if p.ToolCounter > 0 {
		return
	}
	p.ToolCounter = 200
}

func (p *Player) Center() rl.Vector2 {
	return rl.NewVector2(p.Pos.X+p.Size.X*0.5, p.Pos.Y+p.Size.Y*0.5)
}

func (p *Player) Update(dt float32, movement rl.Vector2, getObstacles func(pos rl.Vector2) []rl.Rectangle, addFarmHole func(pos rl.Vector2)) {
	frameMovement := rl.Vector2Normalize(movement)
	if p.ToolCounter == 0 {
		p.Pos.X += frameMovement.X * dt * 150
		for _, obstacle := range getObstacles(p.Center()) {
			hitbox := p.Hitbox(rl.NewVector2(0, 0))
			if rl.CheckCollisionRecs(hitbox, obstacle) {
				if frameMovement.X > 0 {
					p.Pos.X = obstacle.X - hitbox.Width - p.HitAreaOffset.X
				} else if frameMovement.X < 0 {
					p.Pos.X = obstacle.X + obstacle.Width - p.HitAreaOffset.X
				}
			}
		}

		p.Pos.Y += frameMovement.Y * dt * 150
		for _, obstacle := range getObstacles(p.Center()) {
			hitbox := p.Hitbox(rl.NewVector2(0, 0))
			if rl.CheckCollisionRecs(hitbox, obstacle) {
				if frameMovement.Y > 0 {
					p.Pos.Y = obstacle.Y - hitbox.Height - p.HitAreaOffset.Y
				} else if frameMovement.Y < 0 {
					p.Pos.Y = obstacle.Y + obstacle.Height - p.HitAreaOffset.Y
				}
			}
		}
	}

	p.AnimState = "IDLE"
	if p.ToolCounter > 0 {
		p.ToolCounter -= 100 * dt
		switch p.Tool {
		case "water":
			p.AnimState = "WATERING"
		case "shovel":
			p.AnimState = "DIG"
		case "axe":
			p.AnimState = "AXE"
		}
	} else {
		p.ToolCounter = 0
		if movement.Y > 0 {
			p.AnimState = "WALKING"
		}
		if movement.Y < 0 {
			p.AnimState = "WALKING"
		}
		if movement.X > 0 {
			p.AnimState = "WALKING"
			p.Flipped = false
		}
		if movement.X < 0 {
			p.AnimState = "WALKING"
			p.Flipped = true
		}
	}

	isToolAnimState := slices.Contains([]string{"WATERING", "DIG", "AXE"}, p.AnimState)

	baseAnim := p.BaseAnimations[p.AnimState]
	baseAnim.Update(dt)
	if p.ToolCounter <= 0 && isToolAnimState {
		baseAnim.Reset()
	}
	p.BaseAnimations[p.AnimState] = baseAnim

	styleAnim := p.StyleAnimations[p.AnimState]
	styleAnim.Update(dt)
	if p.ToolCounter <= 0 && isToolAnimState {
		styleAnim.Reset()
	}
	p.StyleAnimations[p.AnimState] = styleAnim

	if toolAnim, ok := p.ToolAnimations[p.AnimState]; ok {
		toolAnim.Update(dt)
		if p.ToolCounter <= 0 && isToolAnimState {
			if p.AnimState == "DIG" {
				addFarmHole(p.ToolHitPoint())
			}
			toolAnim.Reset()
		}
		p.ToolAnimations[p.AnimState] = toolAnim
	}
}

func (p Player) Draw(offset rl.Vector2) {
	destRect := rl.NewRectangle(p.Pos.X-offset.X, p.Pos.Y-offset.Y, p.Size.X, p.Size.Y)
	// base
	baseAnim := p.BaseAnimations[p.AnimState]
	baseImg := baseAnim.Image
	if p.Flipped {
		baseImg = baseAnim.ImageFlipped
	}
	rl.DrawTexturePro(baseImg, baseAnim.SrcRect(), destRect, rl.NewVector2(0, 0), 0, rl.White)
	styleAnim := p.StyleAnimations[p.AnimState]
	styleImg := styleAnim.Image
	if p.Flipped {
		styleImg = styleAnim.ImageFlipped
	}
	rl.DrawTexturePro(styleImg, styleAnim.SrcRect(), destRect, rl.NewVector2(0, 0), 0, rl.White)
	if p.ToolCounter == 0 {
		p.DrawTool(offset)
	}
}

func (p *Player) ToolHitPoint() rl.Vector2 {
	var pos rl.Vector2
	if p.Flipped {
		pos = rl.NewVector2(p.Center().X-float32(p.TileSize), p.Center().Y+float32(p.TileSize)/3)
	} else {
		pos = rl.NewVector2(p.Center().X+float32(p.TileSize), p.Center().Y+float32(p.TileSize)/3)
	}
	return pos
}

func (p Player) DrawTool(offset rl.Vector2) {
	destRect := rl.NewRectangle(p.Pos.X-offset.X, p.Pos.Y-offset.Y, p.Size.X, p.Size.Y)
	if toolAnim, ok := p.ToolAnimations[p.AnimState]; ok {
		img := toolAnim.Image
		if p.Flipped {
			img = toolAnim.ImageFlipped
		}
		rl.DrawTexturePro(img, toolAnim.SrcRect(), destRect, rl.NewVector2(0, 0), 0, rl.White)
	}
}

func (p *Player) Hitbox(offset rl.Vector2) rl.Rectangle {
	return rl.NewRectangle(p.Pos.X+p.HitAreaOffset.X-offset.X, p.Pos.Y+p.HitAreaOffset.Y-offset.Y, p.HitAreaOffset.Width, p.HitAreaOffset.Height)
}

type Sprite struct {
	Draw   func(offset rl.Vector2, drawRoof bool)
	Center func() rl.Vector2
}

func DrawDepth(offset rl.Vector2, sprites []Sprite, drawRoof bool) {
	slices.SortStableFunc(sprites, func(a Sprite, b Sprite) int {
		return int(b.Center().Y - a.Center().Y)
	})
	for _, sprite := range sprites {
		sprite.Draw(offset, drawRoof)
	}
}

func LoadToolUIAssets() map[string]rl.Texture2D {
	res := map[string]rl.Texture2D{}
	res["axe"] = rl.LoadTexture("./resources/UI/axe.png")
	res["shovel"] = rl.LoadTexture("./resources/UI/shovel.png")
	res["water"] = rl.LoadTexture("./resources/UI/water.png")
	return res
}
func UnloadTextureMap[K comparable](assets map[K]rl.Texture2D) {
	for _, tex := range assets {
		rl.UnloadTexture(tex)
	}
}

func LoadCropAssets(rootpath string) (map[string][]rl.Texture2D, error) {
	res := map[string][]rl.Texture2D{}
	err := filepath.Walk(rootpath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if strings.Contains(info.Name(), "_") {
			n := strings.Split(info.Name(), "_")[0]
			fmt.Println(n)
			if _, ok := res[n]; !ok {
				res[n] = []rl.Texture2D{}
			}
			arr := res[n]
			texpath := filepath.Join(rootpath, info.Name())
			fmt.Println(texpath)
			arr = append(arr, rl.LoadTexture(texpath))
			res[n] = arr
		}
		return nil
	})
	if err != nil {
		return map[string][]rl.Texture2D{}, err
	}

	return res, nil
}

func UnloadCropAssets(assetMap map[string][]rl.Texture2D) {
	for _, ls := range assetMap {
		for _, t := range ls {
			rl.UnloadTexture(t)
		}

	}
}

func main() {
	const WIDTH = 1280
	const HEIGHT = 720
	rl.InitWindow(WIDTH, HEIGHT, "Farm sim")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)
	originalTilesize := 16

	cropAssets, err := LoadCropAssets("./resources/elements/Crops")
	defer UnloadCropAssets(cropAssets)
	if err != nil {
		log.Fatal(err)
	}
	tmd, _ := parseMap("./resources/map/0.tmj")
	tm := loadTilemap(&tmd, cropAssets, 48)
	defer tm.Unload()

	humanAnimStyles := NewAnimStyles("./resources/characters/Human")
	defer UnloadAnimStyles(humanAnimStyles)
	humanAnimStylesFlipped := NewAnimStyles("./resources/characters_flipped/Human")
	defer UnloadAnimStyles(humanAnimStylesFlipped)
	// TODO render initial soil when dig

	toolUiAssets := LoadToolUIAssets()
	defer UnloadTextureMap(toolUiAssets)
	tools := []string{}
	for t := range toolUiAssets {
		tools = append(tools, t)
	}

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
		humanAnimStylesFlipped,
		tools,
		"shorthair",
	)

	depthSprites := []Sprite{}

	for _, t := range tm.Objects {
		if t.Type != "house_walls" {
			depthSprites = append(depthSprites, Sprite{
				Draw: func(offset rl.Vector2, drawRoof bool) {
					tm.DrawTile(t, offset)
				},
				Center: func() rl.Vector2 {
					return t.Center(float32(tm.Tilesize))
				},
			})
		}
	}

	depthSprites = append(depthSprites, Sprite{
		Draw: func(offset rl.Vector2, drawRoof bool) {
			player.Draw(offset)
		},
		Center: player.Center,
	})

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

		if rl.IsKeyPressed(rl.KeyS) {
			player.SwitchTool()
		} else if rl.IsKeyPressed(rl.KeyC) {
			if player.Tool == "shovel" && player.ToolCounter == 0 {
				hp := player.ToolHitPoint()
				rects := tm.GetFarmRectsAround(hp)
				idx := slices.IndexFunc(rects, func(r rl.Rectangle) bool {
					return rl.CheckCollisionCircleRec(hp, 5, r)
				})
				if idx != -1 {
					// TODO add digged tile, dont dig again if digged
					r := rects[idx]
					p := GetCellPos(rl.NewVector2(r.X, r.Y), float64(tm.Tilesize))
					if ft, ok := tm.FarmTiles[p]; ok && ft.State == "empty" {
						player.UseTool()
					}
				}
			}
		}

		camScrollDest := rl.NewVector2(player.Pos.X-WIDTH/2, player.Pos.Y-HEIGHT/2)
		dCamScroll := rl.NewVector2((camScrollDest.X-camScroll.X)*2, (camScrollDest.Y-camScroll.Y)*2)

		camScroll.X += dCamScroll.X * dt
		camScroll.Y += dCamScroll.Y * dt
		player.Update(dt, rl.NewVector2(playerMoveX[1]-playerMoveX[0], playerMoveY[1]-playerMoveY[0]), tm.GetObstaclesAround, tm.AddFarmHole)

		rl.BeginDrawing()
		rl.ClearBackground(rl.White)
		tm.DrawTerrain(camScroll, rl.NewVector2(WIDTH, HEIGHT))
		for _, t := range tm.GetTiles(tm.Objects, []string{"house_walls"}) {
			tm.DrawTile(t, camScroll)
		}

		for _, t := range tm.GetFloatingRoofs() {
			tm.DrawTile(t, camScroll)
		}
		DrawDepth(camScroll, depthSprites, true)
		if player.ToolCounter > 0 {
			player.DrawTool(camScroll)
		}

		tm.DrawRoof(camScroll)

		// draw ui
		rl.DrawTextureEx(toolUiAssets[player.Tool], rl.NewVector2(float32(tm.Tilesize), float32(HEIGHT-60)), 0, float32(tm.TileScale), rl.White)
		rl.EndDrawing()
	}
}
