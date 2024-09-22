package main

import (
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/theanzy/farmsim/internal/anim"
	"github.com/theanzy/farmsim/internal/crop"
	"github.com/theanzy/farmsim/internal/inventory"
	"github.com/theanzy/farmsim/internal/tileset"
	"github.com/theanzy/farmsim/internal/world"
)

type Tile struct {
	Variant int
	Type    string
	Pos     rl.Vector2
}

func (t Tile) Center(tilesize float32) rl.Vector2 {
	return rl.NewVector2(t.Pos.X*tilesize+tilesize*0.5, t.Pos.Y*tilesize+0.5)
}

type FarmTile struct {
	Pos rl.Vector2
	// empty, digged, name of plant
	IsWet    bool
	State    string
	PlantAge int
}

type Tilemap struct {
	TileLayers   []map[rl.Vector2]Tile
	Objects      []Tile
	Obstacles    map[rl.Vector2]bool
	Beds         map[rl.Vector2]bool
	tilesetAsset rl.Texture2D
	Tilesize     int
	tilesetCols  int
	tilesetRows  int
	Cols         int
	Rows         int
	Roofs        []Tile
	FarmTiles    map[rl.Vector2]FarmTile
	CropAssets   map[string]crop.StripImg
	TileScale    int
}

func (tm Tilemap) Unload() {
	rl.UnloadTexture(tm.tilesetAsset)
}

func (tm *Tilemap) DrawTerrain(offset rl.Vector2, screenSize rl.Vector2) {
	cstartX, cendX := world.ComputeCellRange(float64(offset.X), float64(offset.X+screenSize.X), float64(tm.Tilesize))
	cstartY, cendY := world.ComputeCellRange(float64(offset.Y), float64(offset.Y+screenSize.Y), float64(tm.Tilesize))
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

	// Draw obstacles
	// for obstaclePos := range tm.Obstacles {
	// 	rl.DrawRectangle(int32(obstaclePos.X*float32(tm.Tilesize)-offset.X), int32(obstaclePos.Y*float32(tm.Tilesize)-offset.Y), int32(tm.Tilesize), int32(tm.Tilesize), rl.White)
	// }
}

func (tm *Tilemap) DrawFarmTiles(offset rl.Vector2) {
	for _, ft := range tm.FarmTiles {
		cellpos := ft.Pos
		tilesize := float32(tm.Tilesize)
		viewpos := rl.Vector2Subtract(
			rl.NewVector2(
				cellpos.X*tilesize,
				cellpos.Y*tilesize,
			),
			offset,
		)
		if ft.State == "empty" {
			if ft.IsWet {
				rl.DrawRectangleV(viewpos, rl.NewVector2(tilesize, tilesize), rl.NewColor(139, 69, 19, 60))
			}
		}
		if ft.State == "digged" {
			cellpos := ft.Pos
			tilesize := float32(tm.Tilesize)
			viewpos := rl.Vector2Subtract(
				rl.NewVector2(
					cellpos.X*tilesize,
					cellpos.Y*tilesize,
				),
				offset,
			)
			DrawTilesetId(tm.tilesetAsset, 818, viewpos, tm.tilesetCols, float32(tm.Tilesize))
			if ft.IsWet {
				rl.DrawRectangleV(viewpos, rl.NewVector2(tilesize, tilesize), rl.NewColor(139, 69, 19, 60))
			}
		} else if ca, ok := tm.CropAssets[ft.State]; ok {
			soil := tm.CropAssets["soil"]
			age := ft.PlantAge
			rl.DrawTexturePro(
				soil.Img,
				soil.SrcRects[age],
				rl.NewRectangle(viewpos.X, viewpos.Y, tilesize, tilesize),
				rl.NewVector2(0, 0),
				0,
				rl.White,
			)
			if ft.IsWet {
				rl.DrawRectangleV(viewpos, rl.NewVector2(tilesize, tilesize), rl.NewColor(139, 69, 19, 60))
			}
			rl.DrawTexturePro(
				ca.Img,
				ca.SrcRects[age],
				rl.NewRectangle(viewpos.X, viewpos.Y, tilesize, tilesize),
				rl.NewVector2(0, 0),
				0,
				rl.White,
			)

		}
	}
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

	cellpos := tile.Pos
	viewpos := rl.Vector2Subtract(rl.NewVector2(cellpos.X*float32(tm.Tilesize), cellpos.Y*float32(tm.Tilesize)), offset)
	DrawTilesetId(tm.tilesetAsset, tile.Variant, viewpos, tm.tilesetCols, float32(tm.Tilesize))
}

func DrawTilesetId(tileset rl.Texture2D, id int, pos rl.Vector2, tilesetCols int, tilesize float32) {
	cols := tilesetCols
	tx := float32((id % cols)) * tilesize
	ty := float32((id / cols)) * tilesize
	srcRect := rl.NewRectangle(tx, ty, tilesize, tilesize)
	rl.DrawTextureRec(tileset, srcRect, pos, rl.White)
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
	return world.GetTileRectsAround(tm.Obstacles, pos, float32(tm.Tilesize))
}
func (tm *Tilemap) GetFarmRectsAround(pos rl.Vector2) []rl.Rectangle {
	return world.GetTileRectsAround(tm.FarmTiles, pos, float32(tm.Tilesize))
}

func (tm *Tilemap) AddFarmHole(pos rl.Vector2) {
	cellpos := world.GetCellPos(pos, float64(tm.Tilesize))
	if t, ok := tm.FarmTiles[cellpos]; ok {
		t.State = "digged"
		tm.FarmTiles[cellpos] = t
	}
}

func (tm *Tilemap) AddWetTile(pos rl.Vector2) {
	cellpos := world.GetCellPos(pos, float64(tm.Tilesize))
	if t, ok := tm.FarmTiles[cellpos]; ok {
		t.IsWet = true
		tm.FarmTiles[cellpos] = t
	}
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

func LoadTilemap(tmd *tileset.TileMapData, cropAssets map[string]crop.StripImg, tilesize int) Tilemap {
	var img = rl.LoadImage("./resources/map/tilesets.png")
	defer rl.UnloadImage(img)
	scale := tilesize / tmd.TileWidth
	rl.ImageResizeNN(img, img.Width*int32(scale), img.Height*int32(scale))

	var tm Tilemap
	tm.tilesetCols = int(img.Width) / tilesize
	tm.tilesetRows = int(img.Height) / tilesize
	tm.Cols = tmd.Width
	tm.Rows = tmd.Height
	tm.TileLayers = []map[rl.Vector2]Tile{}
	tm.Tilesize = tilesize
	tm.tilesetAsset = rl.LoadTextureFromImage(img)
	tm.Obstacles = map[rl.Vector2]bool{}
	tm.Objects = []Tile{}
	tm.FarmTiles = map[rl.Vector2]FarmTile{}
	tm.TileScale = scale
	tm.CropAssets = cropAssets
	tm.Beds = map[rl.Vector2]bool{}

	var width = tmd.Width
	sort.SliceStable(tmd.Layers, func(i, j int) bool {
		return tileset.LayerGetProp(tmd.Layers[i], "z") < tileset.LayerGetProp(tmd.Layers[j], "z")
	})

	for _, layer := range tmd.Layers {
		z := tileset.LayerGetProp(layer, "z")
		tiles := map[rl.Vector2]Tile{}
		for i, id := range layer.Data {
			x := i % width
			y := i / width
			cellpos := rl.NewVector2(float32(x), float32(y))
			if layer.Name == "obstacles" && id > 0 {
				tm.Obstacles[cellpos] = true
				continue
			}
			if layer.Name == "bed" && id > 0 {
				tm.Beds[cellpos] = true
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

type Player struct {
	Pos             rl.Vector2
	HitAreaOffset   rl.Rectangle
	AssetSize       rl.Vector2
	TileSize        int
	Size            rl.Vector2
	AnimStyles      anim.AnimStyles
	AnimState       string
	BaseAnimations  map[string]anim.Animation
	ToolAnimations  map[string]anim.Animation
	StyleAnimations map[string]anim.Animation
	Flipped         bool
	Tool            string
	Tools           []string
	ToolCounter     float32
}

func GetSrcRects(stripCount int, imgSize rl.Vector2) ([]rl.Rectangle, []rl.Rectangle) {
	original := []rl.Rectangle{}
	flipped := []rl.Rectangle{}
	for i := range stripCount {
		original = append(original, rl.NewRectangle(float32(i)*imgSize.X, 0, imgSize.X, imgSize.Y))
		flipped = append(flipped, rl.NewRectangle(float32(i)*imgSize.X, imgSize.Y, imgSize.X, imgSize.Y))
	}
	return original, flipped
}

func NewPlayer(pos rl.Vector2, tilesize int, scale int, animStyles anim.AnimStyles, tools []string, style string) Player {
	playerImg := animStyles["IDLE"].Base
	stripCount := animStyles["IDLE"].StripCount

	assetSize := rl.NewVector2(float32(playerImg.Width)/float32(stripCount), float32(playerImg.Height)/2)
	size := rl.NewVector2(float32(assetSize.X)*float32(scale), float32(assetSize.Y)*float32(scale))

	hitboxSize := assetSize.X * 0.4

	hitRect := rl.NewRectangle(size.X/2-hitboxSize/2, size.Y/2-hitboxSize/2, hitboxSize, hitboxSize)

	baseAnimations := map[string]anim.Animation{}
	toolAnimations := map[string]anim.Animation{}
	styleAnimations := map[string]anim.Animation{}
	animSpeed := 12
	for a, animStyle := range animStyles {
		baseSrcRects, baseSrcRectsFlipped := GetSrcRects(animStyle.StripCount, assetSize)
		baseAnimations[a] = anim.Animation{
			Image:          animStyle.Base,
			AssetSize:      assetSize,
			X:              0,
			Speed:          float32(animSpeed),
			StripCount:     float32(animStyle.StripCount),
			SrcRects:       baseSrcRects,
			SrcRectFlipped: baseSrcRectsFlipped,
		}

		if v, ok := animStyle.Variants[style]; ok {
			srcRects, srcRectsFlipped := GetSrcRects(animStyle.StripCount, assetSize)
			styleAnimations[a] = anim.Animation{
				Image:          v,
				AssetSize:      assetSize,
				X:              0,
				Speed:          float32(animSpeed),
				StripCount:     float32(animStyle.StripCount),
				SrcRects:       srcRects,
				SrcRectFlipped: srcRectsFlipped,
			}
		}
		if v, ok := animStyle.Variants["tools"]; ok {
			srcRects, srcRectsFlipped := GetSrcRects(animStyle.StripCount, assetSize)
			toolAnimations[a] = anim.Animation{
				Image:          v,
				AssetSize:      assetSize,
				X:              0,
				Speed:          float32(animSpeed),
				StripCount:     float32(animStyle.StripCount),
				SrcRects:       srcRects,
				SrcRectFlipped: srcRectsFlipped,
			}
		}
	}

	return Player{
		Pos:             pos,
		HitAreaOffset:   hitRect,
		AssetSize:       assetSize,
		Size:            size,
		TileSize:        tilesize,
		AnimStyles:      animStyles,
		AnimState:       "IDLE",
		BaseAnimations:  baseAnimations,
		ToolAnimations:  toolAnimations,
		StyleAnimations: styleAnimations,
		Flipped:         false,
		Tool:            "water",
		Tools:           tools,
		ToolCounter:     0,
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

	rl.DrawTexturePro(baseImg, baseAnim.SrcRect(p.Flipped), destRect, rl.NewVector2(0, 0), 0, rl.White)
	styleAnim := p.StyleAnimations[p.AnimState]
	styleImg := styleAnim.Image
	rl.DrawTexturePro(styleImg, styleAnim.SrcRect(p.Flipped), destRect, rl.NewVector2(0, 0), 0, rl.White)
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
		rl.DrawTexturePro(img, toolAnim.SrcRect(p.Flipped), destRect, rl.NewVector2(0, 0), 0, rl.White)
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
		return int(a.Center().Y - b.Center().Y)
	})
	for _, sprite := range sprites {
		sprite.Draw(offset, drawRoof)
	}
}

func LoadToolUIAsset() map[string]rl.Texture2D {
	res := map[string]rl.Texture2D{}
	res["axe"] = rl.LoadTexture("./resources/UI/axe.png")
	res["shovel"] = rl.LoadTexture("./resources/UI/shovel.png")
	res["water"] = rl.LoadTexture("./resources/UI/water.png")
	return res
}

func LoadUIAsset() map[string]rl.Texture2D {
	res := map[string]rl.Texture2D{}
	res["selectbox_bl"] = rl.LoadTexture("./resources/UI/selectbox_bl.png")
	res["selectbox_br"] = rl.LoadTexture("./resources/UI/selectbox_br.png")
	res["selectbox_tl"] = rl.LoadTexture("./resources/UI/selectbox_tl.png")
	res["selectbox_tr"] = rl.LoadTexture("./resources/UI/selectbox_tr.png")
	return res
}
func UnloadTextureMap[K comparable](assets map[K]rl.Texture2D) {
	for _, tex := range assets {
		rl.UnloadTexture(tex)
	}
}

func main() {
	const WIDTH = 1280
	const HEIGHT = 720
	rl.InitWindow(WIDTH, HEIGHT, "Farm sim")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)
	originalTilesize := 16

	// tileset id
	var crops = []string{"carrot", "cauliflower", "pumpkin", "sunflower", "radish", "parsnip", "potato", "cabbage", "beetroot", "wheat", "kale"}
	cropAssets, err := crop.LoadCropAssets("./resources/elements/Crops", append(crops, "soil"))
	if err != nil {
		return
	}
	defer crop.UnloadCropAssets(cropAssets)

	uiAssets := LoadUIAsset()
	defer UnloadTextureMap(uiAssets)

	tmd, _ := tileset.ParseMap("./resources/map/0.tmj")
	tm := LoadTilemap(&tmd, cropAssets, 48)
	defer tm.Unload()

	const cropTilesetStartId = 691
	getFullCropTileId := func(cropName string) int {
		idx := slices.Index(crops, cropName)
		return cropTilesetStartId + idx
	}

	currentSeed := "carrot"
	seedUiPos := rl.NewVector2(
		float32(tm.Tilesize),
		HEIGHT-80,
	)

	supportedStyles := []string{"IDLE", "WALKING", "WATERING", "DIG", "AXE"}
	humanAnimStyles := anim.NewAnimStyles("./resources/characters/Human", supportedStyles)
	defer anim.UnloadAnimStyles(humanAnimStyles)

	toolsUIAsset := LoadToolUIAsset()
	defer UnloadTextureMap(toolsUIAsset)
	tools := []string{}
	for t := range toolsUIAsset {
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
					if t.Type == "house_decor" {
						c := t.Center(float32(tm.Tilesize))
						return rl.NewVector2(c.X, c.Y-float32(tm.Tilesize)*2)
					}
					return t.Center(float32(tm.Tilesize))
				},
			})
		}
	}

	depthSprites = append(depthSprites, Sprite{
		Draw: func(offset rl.Vector2, drawRoof bool) {
			player.Draw(offset)
		},
		Center: func() rl.Vector2 {
			return rl.NewVector2(player.Center().X, player.Pos.Y+float32(tm.Tilesize)*0.5)
		},
	})

	playerInventory := inventory.NewInventory(cropAssets)
	defer playerInventory.DeinitInventory()
	inventoryContainer := rl.NewRectangle(WIDTH*0.5-800*0.5, HEIGHT*0.5-600*0.5, 800, 600)
	inventoryId := ""
	showInventory := false

	const padding float32 = 28.0
	slotSize := float32(tm.Tilesize)
	colCount := float32(math.Floor(float64(inventoryContainer.Width / (slotSize + padding))))

	var camScroll = rl.NewVector2(0, 0)
	var day int = 0
	transitionCounter := 0.0
	for !rl.WindowShouldClose() {
		playerMoveX := []float32{0, 0}
		playerMoveY := []float32{0, 0}
		dt := rl.GetFrameTime()
		if transitionCounter > 0 {
			transitionCounter = math.Max(0, transitionCounter-200.0*float64(dt))
		} else if showInventory {
			if rl.IsKeyPressed(rl.KeyI) {
				showInventory = false
			} else if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
				for i, item := range playerInventory.Items() {
					mpos := rl.GetMousePosition()
					irect := InventorySlotRect(inventoryContainer, i, padding, slotSize, colCount)
					if rl.CheckCollisionPointRec(mpos, irect) {
						inventoryId = item.Name
					}
				}

			}
		} else {
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
			} else if rl.IsKeyPressed(rl.KeyC) && player.ToolCounter == 0 {
				if player.Tool == "shovel" {
					hp := player.ToolHitPoint()
					rects := tm.GetFarmRectsAround(hp)
					idx := slices.IndexFunc(rects, func(r rl.Rectangle) bool {
						return rl.CheckCollisionCircleRec(hp, 5, r)
					})
					if idx != -1 {
						r := rects[idx]
						p := world.GetCellPos(rl.NewVector2(r.X, r.Y), float64(tm.Tilesize))
						if ft, ok := tm.FarmTiles[p]; ok && ft.State == "empty" {
							player.UseTool()
						}
					}
				} else if player.Tool == "water" {
					hp := player.ToolHitPoint()
					rects := tm.GetFarmRectsAround(hp)
					idx := slices.IndexFunc(rects, func(r rl.Rectangle) bool {
						return rl.CheckCollisionCircleRec(hp, 5, r)
					})
					if idx != -1 {
						player.UseTool()
						tm.AddWetTile(player.ToolHitPoint())
					}
				}
			}
			if rl.IsKeyPressed(rl.KeyD) {
				if idx := slices.Index(crops, currentSeed); idx != -1 {
					idx = (idx + 1) % len(crops)
					currentSeed = crops[idx]
				}
			}
			if rl.IsKeyPressed(rl.KeyX) {
				hp := player.ToolHitPoint()
				rects := tm.GetFarmRectsAround(hp)
				idx := slices.IndexFunc(rects, func(r rl.Rectangle) bool {
					return rl.CheckCollisionCircleRec(hp, 5, r)
				})
				if idx != -1 {
					cp := world.GetCellPos(rl.NewVector2(rects[idx].X, rects[idx].Y), float64(tm.Tilesize))
					if ft, ok := tm.FarmTiles[cp]; ok && ft.State == "digged" {
						ft.State = currentSeed
						tm.FarmTiles[cp] = ft
					}
				}
			}
			if rl.IsKeyPressed(rl.KeySpace) {
				hp := player.ToolHitPoint()
				chp := world.GetCellPos(hp, float64(tm.Tilesize))
				if _, ok := tm.Beds[chp]; ok {
					day += 1
					// start transition. block all inputs
					transitionCounter = 512
					// add plant age if soil is wet, reset soil to dry
					for p, ft := range tm.FarmTiles {
						if ft.IsWet {
							ft.PlantAge = ft.PlantAge + 1
						}
						ft.IsWet = false
						tm.FarmTiles[p] = ft
					}
				}
			}
			if rl.IsKeyPressed(rl.KeyI) {
				showInventory = !showInventory
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
		tm.DrawFarmTiles(camScroll)

		for _, t := range tm.GetTiles(tm.Objects, []string{"house_walls"}) {
			tm.DrawTile(t, camScroll)
		}

		idx := slices.IndexFunc(tm.TileLayers, func(l map[rl.Vector2]Tile) bool {
			for _, t := range l {
				if t.Type != "house_floor" {
					return false
				}
				return true
			}
			return false
		})
		_, inHouse := tm.TileLayers[idx][world.GetCellPos(player.Center(), float64(tm.Tilesize))]

		if !inHouse {
			for _, t := range tm.GetFloatingRoofs() {
				tm.DrawTile(t, camScroll)
			}
		}

		DrawDepth(camScroll, depthSprites, true)
		if player.ToolCounter > 0 {
			player.DrawTool(camScroll)
		}

		if !inHouse {
			tm.DrawRoof(camScroll)
		}

		// draw ui
		rl.DrawText(fmt.Sprintf("Day %d", day), 10, 10, 32, rl.White)
		DrawTilesetId(tm.tilesetAsset, getFullCropTileId(currentSeed), seedUiPos, tm.tilesetCols, float32(tm.Tilesize))

		toolTex := toolsUIAsset[player.Tool]
		DrawTextureCenterV(toolTex, rl.NewVector2(float32(tm.Tilesize)*2, HEIGHT-80), float32(tm.Tilesize), float32(tm.TileScale))
		if transitionCounter > 256 {
			rl.DrawRectangle(0, 0, WIDTH, HEIGHT, rl.NewColor(0, 0, 0, uint8(512-transitionCounter)))
		} else if transitionCounter > 0 {
			rl.DrawRectangle(0, 0, WIDTH, HEIGHT, rl.NewColor(0, 0, 0, uint8(transitionCounter)))
		}

		// draw inventory
		if showInventory {
			rl.DrawRectangleRec(inventoryContainer, rl.Beige)
			rl.DrawText("Inventory", int32(inventoryContainer.X)+20, int32(inventoryContainer.Y)+10, 30, rl.White)
			items := playerInventory.Items()
			inventoryIdx := 0
			for i, item := range items {
				rect := InventorySlotRect(inventoryContainer, i, padding, slotSize, colCount)
				rl.DrawRectangleRec(rect, rl.Brown)
				tx := rect.X + slotSize*0.5 - float32(item.Image.Width)*float32(tm.TileScale)*0.5
				ty := rect.Y + slotSize*0.5 - float32(item.Image.Height)*float32(tm.TileScale)*0.5
				rl.DrawTextureEx(item.Image, rl.NewVector2(tx, ty), 0, float32(tm.TileScale), rl.White)
				if inventoryId == item.Name {
					inventoryIdx = i
					shift := slotSize * 0.25
					stl := rl.NewVector2(rect.X-shift, rect.Y-shift)
					str := rl.NewVector2(rect.X+slotSize-shift, rect.Y-shift)
					sbl := rl.NewVector2(rect.X-shift, rect.Y+slotSize-shift)
					sbr := rl.NewVector2(rect.X+slotSize-shift, rect.Y+slotSize-shift)
					rl.DrawTextureEx(uiAssets["selectbox_tl"], stl, 0, float32(tm.TileScale), rl.White)
					rl.DrawTextureEx(uiAssets["selectbox_tr"], str, 0, float32(tm.TileScale), rl.White)
					rl.DrawTextureEx(uiAssets["selectbox_br"], sbr, 0, float32(tm.TileScale), rl.White)
					rl.DrawTextureEx(uiAssets["selectbox_bl"], sbl, 0, float32(tm.TileScale), rl.White)
				}
			}
			descRect := rl.NewRectangle(inventoryContainer.X+padding, inventoryContainer.Y+inventoryContainer.Height-padding-180, inventoryContainer.Width-padding*2, 180)
			rl.DrawRectangleRec(descRect, rl.White)
			rl.DrawText(items[inventoryIdx].Name, int32(descRect.X+padding), int32(descRect.Y+padding*0.5), 25, rl.Black)

			priceText := fmt.Sprintf("$%d", items[inventoryIdx].SellPrice)
			priceTextW := rl.MeasureText(priceText, 25)
			rl.DrawText(priceText, int32(descRect.X+descRect.Width-padding-float32(priceTextW)), int32(descRect.Y+padding*0.5), 25, rl.DarkGray)

			// description
			desc := items[inventoryIdx].Description
			words := strings.Split(desc, " ")
			fontsize := 20
			descWidth := rl.MeasureText(desc, int32(fontsize))
			if descWidth > int32(descRect.Width-padding*2) {
				gap := 10
				rl.DrawText(strings.Join(words[0:gap], " "), int32(descRect.X+padding), int32(descRect.Y+padding*2), int32(fontsize), rl.Gray)
				i := 0
				h := 3
				for {
					start := i + gap
					end := start + gap
					breaking := false
					if len(words) < start {
						start = len(words)
					}
					if len(words) < end {
						end = len(words)
						breaking = true
					}
					w := strings.Join(words[start:end], " ")
					rl.DrawText(w, int32(descRect.X+padding), int32(descRect.Y+padding*float32(h)), int32(fontsize), rl.Gray)
					i += gap
					h += 1
					if breaking {
						break
					}
				}

			} else {
				rl.DrawText(desc, int32(descRect.X+padding), int32(descRect.Y+padding*2), int32(fontsize), rl.Gray)
			}
		}
		rl.EndDrawing()
	}
}

func InventorySlotRect(inventoryContainer rl.Rectangle, i int, padding float32, slotsize float32, colCount float32) rl.Rectangle {
	x := inventoryContainer.X + padding + ((padding + slotsize) * (float32(math.Mod(float64(i), float64(colCount)))))
	y := inventoryContainer.Y + padding*2 + ((padding + slotsize) * (float32(math.Floor(float64(i) / float64(colCount)))))
	rect := rl.NewRectangle(x, y, slotsize, slotsize)
	return rect
}

func DrawTextureCenterV(tex rl.Texture2D, pos rl.Vector2, tilesize float32, tilescale float32) {
	rl.DrawTextureEx(
		tex,
		rl.NewVector2(
			pos.X+tilesize*0.5-float32(tex.Width)*tilescale*0.5,
			pos.Y+tilesize*0.5-float32(tex.Height)*tilescale*0.5,
		),
		0,
		tilescale,
		rl.White,
	)

}
