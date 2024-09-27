package main

import (
	"fmt"
	"math"
	rand "math/rand/v2"
	"slices"
	"sort"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/theanzy/farmsim/internal/anim"
	"github.com/theanzy/farmsim/internal/crop"
	"github.com/theanzy/farmsim/internal/items"
	"github.com/theanzy/farmsim/internal/render"
	"github.com/theanzy/farmsim/internal/sfx"
	"github.com/theanzy/farmsim/internal/strip"
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
	IsWet   bool
	State   string
	CropAge int
}

type Tree struct {
	State         string
	Img           strip.StripImg
	Pos           rl.Vector2
	Hitbox        rl.Rectangle
	Size          rl.Vector2
	Center        rl.Vector2
	frame         float32
	hunkImg       rl.Texture2D
	hunkSize      rl.Vector2
	shakeDuration float32
	WoodCount     int
}

func NewTree(img strip.StripImg, hunkImg rl.Texture2D, cellpos rl.Vector2, tilesize float32, tilescale float32, woodcount int) Tree {
	size := rl.NewVector2(float32(img.Img.Width/int32(img.StripCount))*tilescale, float32(img.Img.Height)*tilescale)
	pos := rl.NewVector2(cellpos.X*tilesize, cellpos.Y*tilesize)
	hunkSize := rl.NewVector2(float32(hunkImg.Width)*tilescale, float32(hunkImg.Height)*tilescale)
	return Tree{
		State:         "idle",
		Img:           img,
		Pos:           pos,
		Size:          size,
		Hitbox:        NewTreeHitbox(img, cellpos, tilesize, tilescale),
		Center:        rl.NewVector2(pos.X+size.X/2, pos.Y+size.Y/2),
		shakeDuration: 0,
		hunkImg:       hunkImg,
		hunkSize:      hunkSize,
		WoodCount:     woodcount,
	}
}

func NewTreeHitbox(img strip.StripImg, pos rl.Vector2, tilesize float32, tilescale float32) rl.Rectangle {
	hitWidth := tilesize * 0.8
	hitHeight := tilesize * 0.8
	width := float32(img.Img.Width/int32(img.StripCount)) * tilescale
	height := float32(img.Img.Height) * tilescale
	return rl.NewRectangle(
		pos.X*tilesize+width*0.5-hitWidth*0.5,
		pos.Y*tilesize+height-hitHeight,
		hitWidth,
		hitHeight,
	)
}

func (t *Tree) Update(dt float32) {
	if t.State == "shaking" {
		t.shakeDuration -= 100 * dt
		if t.shakeDuration <= 0 {
			t.shakeDuration = 0
			t.State = "dead"
		} else {
			t.frame += dt * 5
			if t.frame >= float32(t.Img.StripCount) {
				t.frame = 0
			}
		}

	}
}

func (t *Tree) Shake(duration float32) {
	t.State = "shaking"
	t.shakeDuration = duration
}

func (t *Tree) Draw(offset rl.Vector2) {
	if t.State == "dead" {
		hunkSize := t.hunkSize
		x := t.Pos.X + t.Size.X*0.5 - hunkSize.X*0.5 - offset.X
		y := t.Pos.Y + t.Size.Y - hunkSize.Y - offset.Y
		src := rl.NewRectangle(0, 0, float32(t.hunkImg.Width), float32(t.hunkImg.Height))
		dest := rl.NewRectangle(x, y, hunkSize.X, hunkSize.Y)
		rl.DrawTexturePro(t.hunkImg, src, dest, rl.NewVector2(0, 0), 0, rl.White)
		return
	}
	dest := rl.NewRectangle(
		t.Pos.X-offset.X,
		t.Pos.Y-offset.Y,
		t.Size.X,
		t.Size.Y,
	)
	x := int(math.Floor(float64(t.frame)))
	rl.DrawTexturePro(t.Img.Img, t.Img.SrcRects[x], dest, rl.NewVector2(0, 0), 0, rl.White)

	// hitbox := t.Hitbox
	// hitbox.X -= offset.X
	// hitbox.Y -= offset.Y
	// rl.DrawRectangleRec(hitbox, rl.Red)
}

type Tilemap struct {
	TileLayers   []map[rl.Vector2]Tile
	Objects      []Tile
	Obstacles    map[rl.Vector2]bool
	Trees        []Tree
	Beds         map[rl.Vector2]bool
	tilesetAsset rl.Texture2D
	Tilesize     int
	tilesetCols  int
	tilesetRows  int
	Cols         int
	Rows         int
	Roofs        []Tile
	FarmTiles    map[rl.Vector2]FarmTile
	CropAssets   map[string]strip.StripImg
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
			age := min(ft.CropAge, ca.StripCount-1)
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
	treeRects := []rl.Rectangle{}
	for _, t := range tm.Trees {
		treeRects = append(treeRects, t.Hitbox)
	}
	return append(world.GetTileRectsAround(tm.Obstacles, pos, float32(tm.Tilesize)), treeRects...)
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

func GetCollidedTreeIdx(trees []Tree, hitpoint rl.Vector2) int {
	return slices.IndexFunc(trees, func(t Tree) bool {
		return rl.CheckCollisionPointRec(hitpoint, t.Hitbox)
	})
}

func LoadImgWithScale(imgPath string, scale int32) rl.Texture2D {
	var img = rl.LoadImage(imgPath)
	defer rl.UnloadImage(img)
	rl.ImageResizeNN(img, img.Width*scale, img.Height*scale)
	return rl.LoadTextureFromImage(img)
}

func LoadTilemap(tmd *tileset.TileMapData, cropAssets map[string]strip.StripImg, treeAssets map[string]strip.StripImg, treeHunkImg rl.Texture2D, tilesize int) Tilemap {
	scale := tilesize / tmd.TileWidth

	var tm Tilemap
	tm.tilesetAsset = LoadImgWithScale("./resources/map/tilesets.png", int32(scale))
	tm.Tilesize = tilesize
	tm.tilesetCols = int(tm.tilesetAsset.Width) / tilesize
	tm.tilesetRows = int(tm.tilesetAsset.Height) / tilesize
	tm.Trees = []Tree{}
	tm.Cols = tmd.Width
	tm.Rows = tmd.Height
	tm.TileLayers = []map[rl.Vector2]Tile{}
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
			if id == 0 {
				continue
			}
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
					Pos:     cellpos,
					State:   "empty",
					CropAge: 0,
				}
			}
			if layer.Name == "tree_real" && id > 0 {
				if id == 4102 {
					tm.Trees = append(tm.Trees, NewTree(treeAssets["tree_01"], treeHunkImg, cellpos, float32(tm.Tilesize), float32(tm.TileScale), 5))
				} else if id == 4103 {
					tm.Trees = append(tm.Trees, NewTree(treeAssets["tree_02"], treeHunkImg, cellpos, float32(tm.Tilesize), float32(tm.TileScale), 2))
				}
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
	BaseAnimations  map[string]anim.StripAnimation
	ToolAnimations  map[string]anim.StripAnimation
	StyleAnimations map[string]anim.StripAnimation
	Flipped         bool
	Tool            string
	Tools           []string
	ToolCounter     float32
}

func NewPlayer(pos rl.Vector2, tilesize int, scale int, animStyles anim.AnimStyles, tools []string, style string) Player {
	playerImg := animStyles["IDLE"].Base
	assetSize := rl.NewVector2(
		float32(playerImg.Width)/float32(animStyles["IDLE"].StripCount),
		float32(playerImg.Height)/2,
	)
	size := rl.NewVector2(
		float32(assetSize.X)*float32(scale),
		float32(assetSize.Y)*float32(scale),
	)

	hitboxSize := assetSize.X * 0.4

	hitRect := rl.NewRectangle(size.X/2-hitboxSize/2, size.Y/2-hitboxSize/2, hitboxSize, hitboxSize)

	baseAnimations := map[string]anim.StripAnimation{}
	toolAnimations := map[string]anim.StripAnimation{}
	styleAnimations := map[string]anim.StripAnimation{}
	animSpeed := 12
	for a, animStyle := range animStyles {
		baseAnimations[a] = anim.NewStripAnimation(
			animStyle.Base,
			assetSize,
			float32(animSpeed),
			float32(animStyle.StripCount),
		)

		if v, ok := animStyle.Variants[style]; ok {
			styleAnimations[a] = anim.NewStripAnimation(
				v,
				assetSize,
				float32(animSpeed),
				float32(animStyle.StripCount),
			)
		}
		if v, ok := animStyle.Variants["tools"]; ok {
			toolAnimations[a] = anim.NewStripAnimation(
				v,
				assetSize,
				float32(animSpeed),
				float32(animStyle.StripCount),
			)
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

func (p *Player) UseTool(duration float32) {
	if p.ToolCounter > 0 {
		return
	}
	p.ToolCounter = duration
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

func LoadToolUIAsset() map[string]rl.Texture2D {
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

func main() {
	const WIDTH = 1280
	const HEIGHT = 720
	rl.InitWindow(WIDTH, HEIGHT, "Farm sim")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)
	originalTilesize := 16

	// tileset id
	var crops = []string{"carrot", "cauliflower", "pumpkin", "sunflower", "radish", "parsnip", "potato", "cabbage", "beetroot", "wheat", "kale"}
	cropAssets, err := crop.LoadCropAssets("./resources/elements/Crops", append(crops, "soil", "wood"))
	if err != nil {
		return
	}
	defer strip.UnloadMapStripImg(cropAssets)
	woodDropSfx := sfx.NewItemDrop(cropAssets["wood"].Img, 50)

	uiAssets := map[string]rl.Texture2D{
		"selectbox_bl": rl.LoadTexture("./resources/UI/selectbox_bl.png"),
		"selectbox_br": rl.LoadTexture("./resources/UI/selectbox_br.png"),
		"selectbox_tl": rl.LoadTexture("./resources/UI/selectbox_tl.png"),
		"selectbox_tr": rl.LoadTexture("./resources/UI/selectbox_tr.png"),
		"arrow_left":   rl.LoadTexture("./resources/UI/arrow_left.png"),
		"arrow_right":  rl.LoadTexture("./resources/UI/arrow_right.png"),
	}
	defer UnloadTextureMap(uiAssets)
	treeAssets := map[string]strip.StripImg{
		"tree_01": strip.NewStripImg(rl.LoadTexture("./resources/elements/Plants/spr_deco_tree_01_strip4.png"), 4),
		"tree_02": strip.NewStripImg(rl.LoadTexture("./resources/elements/Plants/spr_deco_tree_02_strip4.png"), 4),
	}
	defer strip.UnloadMapStripImg(treeAssets)
	treeHunkImg := rl.LoadTexture("./resources/elements/Plants/tree_hunk.png")
	defer rl.UnloadTexture(treeHunkImg)

	tmd, _ := tileset.ParseMap("./resources/map/0.tmj")
	tm := LoadTilemap(&tmd, cropAssets, treeAssets, treeHunkImg, 48)
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

	depthRenderer := render.NewDepthRenderer(20)
	for _, t := range tm.Objects {
		if t.Type != "house_walls" {
			depthRenderer.Sprites = append(depthRenderer.Sprites, render.Sprite{
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
	for i := range tm.Trees {
		depthRenderer.Sprites = append(depthRenderer.Sprites, render.Sprite{
			Draw: func(offset rl.Vector2, drawRoof bool) {
				tm.Trees[i].Draw(offset)
			},
			Center: func() rl.Vector2 {
				return tm.Trees[i].Center
			},
		})

	}

	depthRenderer.Sprites = append(depthRenderer.Sprites, render.Sprite{
		Draw: func(offset rl.Vector2, drawRoof bool) {
			player.Draw(offset)
		},
		Center: func() rl.Vector2 {
			return rl.NewVector2(player.Center().X, player.Pos.Y+float32(tm.Tilesize)*0.5)
		},
	})

	allItems := items.LoadItems(cropAssets)
	defer items.UnloadItems(allItems)

	playerInventory := items.NewInventory(allItems)
	inventoryUI := items.NewInventoryUI(WIDTH, HEIGHT, float32(tm.Tilesize))
	showInventory := false
	seedShop := items.NewSeedShop("Seed merchant", allItems)
	seedShopUI := items.NewShopUI(rl.NewVector2(WIDTH, HEIGHT), float32(tm.Tilesize), uiAssets)
	showShop := true

	var camScroll = rl.NewVector2(0, 0)
	var day int = 0
	transitionCounter := 0.0
	overlays := []rl.Color{
		rl.NewColor(255, 255, 255, 0),
		rl.NewColor(247, 228, 160, 30),
		rl.NewColor(255, 151, 89, 80),
		rl.NewColor(70, 63, 103, 100),
		rl.NewColor(4, 26, 54, 150),
		rl.NewColor(84, 88, 131, 80),
	}
	overlayIdx := 1
	overlayColor := overlays[0]
	overlayCounter := 0

	for !rl.WindowShouldClose() {
		playerMoveX := []float32{0, 0}
		playerMoveY := []float32{0, 0}
		dt := rl.GetFrameTime()

		if transitionCounter > 0 {
			transitionCounter = math.Max(0, transitionCounter-200.0*float64(dt))
		} else if showInventory {
			if rl.IsKeyPressed(rl.KeyI) {
				showInventory = false
			} else if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
				inventoryUI.ItemClick(&playerInventory, rl.GetMousePosition())
			}
			inventoryUI.ItemHover(&playerInventory, rl.GetMousePosition())
		} else if showShop {
			if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
				seedShopUI.Click(rl.GetMousePosition(), &playerInventory, &seedShop)
			}
			seedShopUI.ItemHover(rl.GetMousePosition(), &playerInventory, &seedShop)
			seedShopUI.Update()
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
							player.UseTool(300)
						}
					}
				} else if player.Tool == "water" {
					hp := player.ToolHitPoint()
					rects := tm.GetFarmRectsAround(hp)
					idx := slices.IndexFunc(rects, func(r rl.Rectangle) bool {
						return rl.CheckCollisionCircleRec(hp, 5, r)
					})
					if idx != -1 {
						player.UseTool(100)
						tm.AddWetTile(player.ToolHitPoint())
					}
				} else if player.Tool == "axe" {
					hp := player.ToolHitPoint()
					if idx := GetCollidedTreeIdx(tm.Trees, hp); idx != -1 && tm.Trees[idx].State == "idle" {
						var duration float32 = 500
						player.UseTool(duration)
						tree := tm.Trees[idx]
						tree.Shake(duration)
						tm.Trees[idx] = tree
						playerInventory.Increase("Wood", tree.WoodCount)
					}
				}
			}
			if rl.IsKeyPressed(rl.KeyD) {
				seeds := playerInventory.AvailableSeeds()
				if idx := slices.Index(seeds, currentSeed); idx != -1 {
					idx = (idx + 1) % len(seeds)
					currentSeed = seeds[idx]
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
						if q := playerInventory.Decrease(items.CropToSeedName(currentSeed), 1); q == 0 {
							seeds := playerInventory.AvailableSeeds()
							if len(seeds) > 0 {
								currentSeed = playerInventory.AvailableSeeds()[0]
							}
						}
					}
				}
			}
			if rl.IsKeyPressed(rl.KeySpace) {
				hp := player.ToolHitPoint()
				chp := world.GetCellPos(hp, float64(tm.Tilesize))
				if ft, ok := GetFullyGrownCrop(chp, tm.FarmTiles, cropAssets); ok {
					playerInventory.Increase(items.CropToCropName(ft.State), 1)
					ft.State = "digged"
					ft.CropAge = 0
					tm.FarmTiles[ft.Pos] = ft
					// TODO add sfx for harvest

				} else if _, ok := tm.Beds[chp]; ok {
					day += 1
					// start transition. block all inputs
					transitionCounter = 512
					// add plant age if soil is wet, reset soil to dry
					for p, ft := range tm.FarmTiles {
						if ft.IsWet {
							ft.CropAge = ft.CropAge + 1
						}
						ft.IsWet = false
						tm.FarmTiles[p] = ft
					}
				}
			}
			if rl.IsKeyPressed(rl.KeyI) {
				showInventory = !showInventory
			}
			overlayCounter += 1
			if overlayCounter >= 30 {
				overlayCounter = 0
				destColor := overlays[overlayIdx]
				diffR := int(destColor.R) - int(overlayColor.R)
				diffG := int(destColor.G) - int(overlayColor.G)
				diffB := int(destColor.B) - int(overlayColor.B)
				diffA := int(destColor.A) - int(overlayColor.A)
				if diffR != 0 {
					overlayColor.R = uint8(int(overlayColor.R) + diffR/int(math.Abs(float64(diffR))))
				}
				if diffG != 0 {
					overlayColor.G = uint8(int(overlayColor.G) + diffG/int(math.Abs(float64(diffG))))
				}
				if diffB != 0 {
					overlayColor.B = uint8(int(overlayColor.B) + diffB/int(math.Abs(float64(diffB))))
				}
				if diffA != 0 {
					overlayColor.A = uint8(int(overlayColor.A) + diffA/int(math.Abs(float64(diffA))))
				}
				if diffR == 0 && diffG == 0 && diffB == 0 && diffA == 0 {
					overlayIdx = (overlayIdx + 1) % len(overlays)
				}
			}
		}

		camScrollDest := rl.NewVector2(player.Pos.X-WIDTH/2, player.Pos.Y-HEIGHT/2)
		dCamScroll := rl.NewVector2((camScrollDest.X-camScroll.X)*2, (camScrollDest.Y-camScroll.Y)*2)

		camScroll.X += dCamScroll.X * dt
		camScroll.Y += dCamScroll.Y * dt
		player.Update(dt, rl.NewVector2(playerMoveX[1]-playerMoveX[0], playerMoveY[1]-playerMoveY[0]), tm.GetObstaclesAround, tm.AddFarmHole)
		for i, t := range tm.Trees {
			prevState := t.State
			t.Update(dt)
			tm.Trees[i] = t
			if prevState == "shaking" && t.State == "dead" {
				woodDropSfx.Start(
					rl.NewVector2(t.Pos.X+t.Size.X/2, t.Pos.Y+t.Size.Y/2),
					7,
					rl.Vector2Normalize(rl.NewVector2(rand.Float32()*2-1, rand.Float32()*2-1)),
				)
			}
		}
		woodDropSfx.Update(dt)
		depthRenderer.Update()

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

		depthRenderer.Draw(camScroll, true)
		if player.ToolCounter > 0 {
			player.DrawTool(camScroll)
		}

		if !inHouse {
			tm.DrawRoof(camScroll)
		}

		// draw ui
		rl.DrawRectangle(0, 0, WIDTH, HEIGHT, overlayColor)

		rl.DrawText(fmt.Sprintf("Day %d", day), 10, 10, 32, rl.White)
		DrawTilesetId(tm.tilesetAsset, getFullCropTileId(currentSeed), seedUiPos, tm.tilesetCols, float32(tm.Tilesize))

		toolTex := toolsUIAsset[player.Tool]
		DrawTextureCenterV(toolTex, rl.NewVector2(float32(tm.Tilesize)*2, HEIGHT-80), float32(tm.Tilesize), float32(tm.TileScale))
		if transitionCounter > 256 {
			rl.DrawRectangle(0, 0, WIDTH, HEIGHT, rl.NewColor(0, 0, 0, uint8(512-transitionCounter)))
		} else if transitionCounter > 0 {
			rl.DrawRectangle(0, 0, WIDTH, HEIGHT, rl.NewColor(0, 0, 0, uint8(transitionCounter)))
		}

		if showShop {
			seedShopUI.Draw(&seedShop, &playerInventory, uiAssets, float32(tm.TileScale))
		}
		woodDropSfx.Draw(camScroll, float32(tm.TileScale))
		// draw inventory
		if showInventory {
			inventoryUI.Draw(&playerInventory, uiAssets, float32(tm.TileScale))
		}
		rl.EndDrawing()
	}
}

func GetFullyGrownCrop(cellpos rl.Vector2, farmTiles map[rl.Vector2]FarmTile, cropAssets map[string]strip.StripImg) (FarmTile, bool) {
	if ft, ok := farmTiles[cellpos]; ok {
		if ca, ok := cropAssets[ft.State]; ok && ft.CropAge >= ca.StripCount {
			return ft, true
		}
	}
	return FarmTile{}, false
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
