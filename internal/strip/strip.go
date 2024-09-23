package strip

import rl "github.com/gen2brain/raylib-go/raylib"

type StripImg struct {
	Img        rl.Texture2D
	StripCount int
	SrcRects   []rl.Rectangle
}

func NewStripImg(tex rl.Texture2D, stripcount int32) StripImg {
	rects := []rl.Rectangle{}
	unitWidth := tex.Width / stripcount
	for i := range stripcount {
		rect := rl.NewRectangle(float32(i)*float32(unitWidth), 0, float32(unitWidth), float32(tex.Height))
		rects = append(rects, rect)
	}
	return StripImg{
		Img:        tex,
		StripCount: int(stripcount),
		SrcRects:   rects,
	}
}

func UnloadMapStripImg(assets map[string]StripImg) {
	for _, s := range assets {
		rl.UnloadTexture(s.Img)
	}
}
