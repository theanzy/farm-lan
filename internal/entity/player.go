package entity

import (
	"slices"

	"github.com/theanzy/farmsim/internal/anim"

	rl "github.com/gen2brain/raylib-go/raylib"
)

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
