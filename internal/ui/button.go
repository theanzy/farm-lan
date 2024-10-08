package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type ButtonState int

const (
	BtnEnabled ButtonState = iota
	BtnPressed
	BtnDisabled
)

type Button struct {
	Rect           rl.Rectangle
	State          ButtonState
	pressedCounter int
}

func (b *Button) Update() {
	if b.State == BtnPressed {
		if b.pressedCounter > 0 {
			b.pressedCounter = max(0, b.pressedCounter-1)
		} else {
			b.State = BtnEnabled
		}
	}
}

func (b *Button) Press() bool {
	if b.State == BtnPressed || b.State == BtnDisabled {
		return false
	}
	b.State = BtnPressed
	b.pressedCounter = 15
	return true
}

type TextButton struct {
	Button
	text      string
	textWidth float32
	fontsize  int32
	BgColor   rl.Color
	TextColor rl.Color
}

func NewTextButton(rect rl.Rectangle, text string, fontsize int32, backgroundColor rl.Color) TextButton {
	return TextButton{
		Button: Button{
			Rect:  rect,
			State: BtnEnabled,
		},
		text:      text,
		fontsize:  fontsize,
		textWidth: float32(rl.MeasureText(text, fontsize)),
		BgColor:   backgroundColor,
		TextColor: rl.White,
	}
}

func (b *TextButton) SetText(text string) {
	b.text = text
	b.textWidth = float32(rl.MeasureText(text, b.fontsize))
}

func (b *TextButton) Draw() {

	var bgColor rl.Color = b.BgColor
	var diffcolor uint8 = 0
	if b.State == BtnPressed {
		diffcolor = 30
		bgColor = rl.NewColor(b.BgColor.R-diffcolor, b.BgColor.G-diffcolor, b.BgColor.B-diffcolor, 255)
	} else if b.State == BtnDisabled {
		bgColor = rl.NewColor(211, 211, 211, 255)
	}
	rl.DrawRectangleRounded(b.Rect, 0.2, 20, bgColor)
	btnText := b.text
	rl.DrawText(
		btnText,
		int32(b.Rect.X+b.Rect.Width*0.5-b.textWidth*0.5),
		int32(b.Rect.Y+b.Rect.Height*0.5-float32(b.fontsize)*0.5),
		b.fontsize,
		rl.White,
	)

}

type ImgButton struct {
	Button
	img       rl.Texture2D
	tilescale float32
}

func NewImgButton(topleft rl.Vector2, img rl.Texture2D, tilescale float32) ImgButton {
	return ImgButton{
		Button: Button{
			pressedCounter: 0,
			Rect:           rl.NewRectangle(topleft.X, topleft.Y, float32(img.Width)*tilescale, float32(img.Height)*tilescale),
			State:          BtnEnabled,
		},
		img:       img,
		tilescale: tilescale,
	}
}

func (b *ImgButton) Draw() {
	if b.State == BtnPressed {
		rl.DrawTextureEx(
			b.img,
			rl.NewVector2(b.Rect.X+b.Rect.Width*0.5-b.Rect.Width*0.5*0.8, b.Rect.Y+b.Rect.Height*0.5-b.Rect.Height*0.5*0.8),
			0,
			b.tilescale*0.8,
			rl.White,
		)
	} else {
		rl.DrawTextureEx(
			b.img,
			rl.NewVector2(b.Rect.X, b.Rect.Y),
			0,
			b.tilescale,
			rl.White,
		)

	}
}
