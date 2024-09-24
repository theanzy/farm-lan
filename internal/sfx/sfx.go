package sfx

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type ItemDrop struct {
	moveSpeed    float32
	pos          rl.Vector2
	dir          rl.Vector2
	img          rl.Texture2D
	fadeCounter  float32
	fadeDuration float32
}

func NewItemDrop(img rl.Texture2D, fadeDuration float32) ItemDrop {
	return ItemDrop{
		img:          img,
		fadeDuration: fadeDuration,
		fadeCounter:  0,
		moveSpeed:    0,
	}
}

func (d *ItemDrop) Start(pos rl.Vector2, moveSpeed float32, dir rl.Vector2) {
	d.fadeCounter = d.fadeDuration
	d.pos = pos
	d.moveSpeed = moveSpeed
	d.dir = dir
}

func (d *ItemDrop) Update(dt float32) {
	if d.moveSpeed > 0 {
		d.pos.X += d.moveSpeed * d.dir.X
		d.pos.Y += d.moveSpeed * d.dir.Y
		d.moveSpeed = max(d.moveSpeed-0.5, 0)
	} else if d.fadeCounter > 0 {
		d.fadeCounter = max(d.fadeCounter-0.5, 0)
	}
}

func (d *ItemDrop) Draw(offset rl.Vector2, scale float32) {
	if d.fadeCounter > 0 {
		rl.DrawTextureEx(d.img, rl.NewVector2(d.pos.X-offset.X, d.pos.Y-offset.Y), 0, scale, rl.White)
	}
}
