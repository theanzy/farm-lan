package anim

import rl "github.com/gen2brain/raylib-go/raylib"

type AnimatedTile struct {
	Pos       rl.Vector2
	Tilescale float32
	StripAnim StripAnimation
}

func (t *AnimatedTile) Update(dt float32) {
	t.StripAnim.Update(dt)
}

func (t *AnimatedTile) Draw(offset rl.Vector2) {
	srcRect := t.StripAnim.SrcRect(false)
	rl.DrawTexturePro(
		t.StripAnim.Image,
		srcRect,
		rl.NewRectangle(t.Pos.X-offset.X, t.Pos.Y-offset.Y, srcRect.Width*t.Tilescale, srcRect.Height*t.Tilescale),
		// rl.NewRectangle(0, 0, srcRect.Width*t.Tilescale, srcRect.Height*t.Tilescale),
		rl.NewVector2(0, 0),
		0,
		rl.White,
	)
}
