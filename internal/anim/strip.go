package anim

import (
	"log"
	"math"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func LoadStripAnimation(path string, animSpeed float32) StripAnimation {
	basename := filepath.Base(path)
	cfg, err := parseAnimConfig(basename)
	if err != nil {
		log.Fatalln(err)
	}

	img := rl.LoadTexture(path)

	w := float32(img.Width) / float32(cfg.StripCount)
	h := float32(img.Height)
	return NewStripAnimation(
		img,
		rl.NewVector2(w, h),
		animSpeed,
		float32(cfg.StripCount),
	)
}

type StripAnimation struct {
	AssetSize      rl.Vector2
	Image          rl.Texture2D
	X              float32
	Speed          float32
	StripCount     float32
	SrcRects       []rl.Rectangle
	SrcRectFlipped []rl.Rectangle
}

// img is an image with strip facing right on the upper part, and strip facing left on the lower part
func NewStripAnimation(img rl.Texture2D, assetSize rl.Vector2, animationSpeed float32, stripCount float32) StripAnimation {
	rects, rectsFlipped := getSrcRects(int(stripCount), assetSize)
	return StripAnimation{
		Image:          img,
		AssetSize:      assetSize,
		X:              0,
		Speed:          animationSpeed,
		StripCount:     stripCount,
		SrcRects:       rects,
		SrcRectFlipped: rectsFlipped,
	}
}

func getSrcRects(stripCount int, imgSize rl.Vector2) ([]rl.Rectangle, []rl.Rectangle) {
	original := []rl.Rectangle{}
	flipped := []rl.Rectangle{}
	for i := range stripCount {
		original = append(original, rl.NewRectangle(float32(i)*imgSize.X, 0, imgSize.X, imgSize.Y))
		flipped = append(flipped, rl.NewRectangle(float32(i)*imgSize.X, imgSize.Y, imgSize.X, imgSize.Y))
	}
	return original, flipped
}

func (a *StripAnimation) Update(dt float32) {
	a.X += dt * a.Speed
	if a.X >= a.StripCount {
		a.X = 0
	}
}

func (a *StripAnimation) Reset() {
	a.X = 0
}

func (a StripAnimation) SrcRect(flipped bool) rl.Rectangle {
	x := int(math.Floor(float64(a.X)))

	if flipped {
		return a.SrcRectFlipped[x]
	}
	return a.SrcRects[x]
}
