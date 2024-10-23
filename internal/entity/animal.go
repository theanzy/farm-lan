package entity

import (
	"math/rand/v2"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/theanzy/farmsim/internal/anim"
)

type Animal struct {
	Pos           rl.Vector2
	Size          rl.Vector2
	tilesize      int
	scale         int
	Species       string
	frameMovement rl.Vector2
	steps         int
	Flipped       bool
	animations    anim.StripAnimation
}

func NewAnimal(pos rl.Vector2, tilesize int, scale int, species string, animStyle anim.AnimStyle) Animal {
	assetSize := rl.NewVector2(
		float32(animStyle.Base.Width)/float32(animStyle.StripCount),
		float32(animStyle.Base.Height)/2,
	)
	return Animal{
		Pos:      pos,
		Size:     rl.NewVector2(float32(tilesize), float32(tilesize)),
		tilesize: tilesize,
		scale:    scale,
		Species:  species,
		Flipped:  false,
		animations: anim.NewStripAnimation(
			animStyle.Base,
			assetSize,
			6,
			float32(animStyle.StripCount),
		),
	}
}

func (a *Animal) Update(dt float32, getObstacles func(pos rl.Vector2) []rl.Rectangle) {
	a.animations.Update(dt)

	if a.steps == 0 {
		if rand.Float64()*10 < 3 {
			a.frameMovement.X = float32(rand.Float64())*2 - 1
			a.frameMovement.Y = float32(rand.Float64())*2 - 1
			a.frameMovement = rl.Vector2Normalize(a.frameMovement)
		} else {
			a.frameMovement = rl.NewVector2(0, 0)
		}

		a.steps = rand.IntN(50) + 150
	} else {
		a.steps -= 1
		a.Pos.X += a.frameMovement.X * dt * 35
		for _, obstacle := range getObstacles(a.Center()) {
			hitbox := a.Hitbox(rl.NewVector2(0, 0))
			if rl.CheckCollisionRecs(hitbox, obstacle) {
				if a.frameMovement.X > 0 {
					a.Pos.X = obstacle.X - hitbox.Width
				} else if a.frameMovement.X < 0 {
					a.Pos.X = obstacle.X + obstacle.Width
				}
				a.steps = 0
			}
		}

		a.Pos.Y += a.frameMovement.Y * dt * 35
		for _, obstacle := range getObstacles(a.Center()) {
			hitbox := a.Hitbox(rl.NewVector2(0, 0))
			if rl.CheckCollisionRecs(hitbox, obstacle) {
				if a.frameMovement.Y > 0 {
					a.Pos.Y = obstacle.Y - hitbox.Height
				} else if a.frameMovement.Y < 0 {
					a.Pos.Y = obstacle.Y + obstacle.Height
				}
				a.steps = 0
			}
		}
	}

	if a.frameMovement.X > 0 {
		a.Flipped = true
	} else if a.frameMovement.X < 0 {
		a.Flipped = false
	}
}

func (a Animal) Draw(offset rl.Vector2) {
	size := a.Size.X * 1.5
	destRect := rl.NewRectangle(
		a.Pos.X-offset.X-size/2+a.Size.X/2,
		a.Pos.Y-offset.Y-size/2+a.Size.Y/2,
		size,
		size,
	)
	rl.DrawTexturePro(
		a.animations.Image,
		a.animations.SrcRect(a.Flipped),
		destRect,
		rl.NewVector2(0, 0),
		0,
		rl.White,
	)

}

func (a *Animal) Center() rl.Vector2 {
	return rl.NewVector2(a.Pos.X+a.Size.X*0.5, a.Pos.Y+a.Size.Y*0.5-float32(a.tilesize))
}
func (a *Animal) Hitbox(offset rl.Vector2) rl.Rectangle {
	return rl.NewRectangle(
		a.Pos.X-offset.X,
		a.Pos.Y-offset.Y,
		a.Size.X,
		a.Size.Y,
	)
}
