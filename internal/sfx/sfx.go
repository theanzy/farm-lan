package sfx

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type ItemDrop struct {
	Name         string
	moveSpeed    float32
	pos          rl.Vector2
	dir          rl.Vector2
	img          rl.Texture2D
	fadeCounter  float32
	fadeDuration float32
	Fade         bool
}

func NewItemDrop(img rl.Texture2D, name string, fadeDuration float32, fade bool) ItemDrop {
	return ItemDrop{
		img:          img,
		Name:         name,
		fadeDuration: fadeDuration,
		fadeCounter:  0,
		moveSpeed:    0,
		Fade:         fade,
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
	} else if d.fadeCounter > 0 && d.Fade {
		d.fadeCounter = max(d.fadeCounter-0.5, 0)
	}
}

func (d *ItemDrop) Draw(offset rl.Vector2, scale float32) {
	if d.fadeCounter > 0 || !d.Fade {
		w := d.img.Width * int32(scale)
		h := d.img.Height * int32(scale)
		pos := rl.NewVector2(d.pos.X-offset.X, d.pos.Y-offset.Y)
		rl.DrawCircleV(rl.NewVector2(pos.X, pos.Y+float32(h)/3), 14, rl.NewColor(0, 0, 0, 50))
		rl.DrawTextureEx(d.img, rl.NewVector2(pos.X-float32(w)/2, pos.Y-float32(h)/2), 0, scale, rl.White)
	}
}

func (d *ItemDrop) Hitbox(offset rl.Vector2, scale float32) rl.Rectangle {
	w := float32(d.img.Width) * scale
	h := float32(d.img.Height) * scale
	return rl.NewRectangle(
		d.pos.X-offset.X-w/2,
		d.pos.Y-offset.Y-h/2,
		w,
		h,
	)
}

func CheckCollision(ls []ItemDrop, rec rl.Rectangle, scale float32) int {
	for i := range ls {
		if rl.CheckCollisionRecs(ls[i].Hitbox(rl.NewVector2(0, 0), scale), rec) {
			return i
		}
	}
	return -1
}
