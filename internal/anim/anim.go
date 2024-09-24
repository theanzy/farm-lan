package anim

import (
	"log"
	"math"
	"os"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type AnimStyle struct {
	Variants   map[string]rl.Texture2D
	Base       rl.Texture2D
	StripCount int
}

type AnimStyles = map[string]AnimStyle

func NewAnimStyles(dirpath string, supportedStyles []string) AnimStyles {
	styles := AnimStyles{}

	entries, err := os.ReadDir(dirpath)
	if err != nil {
		log.Fatal(err)
	}
	r := regexp.MustCompile(`[a-z](\d+)\.png`)

	for _, e := range entries {
		if !slices.Contains(supportedStyles, e.Name()) {
			continue
		}
		var style = AnimStyle{
			Variants:   map[string]rl.Texture2D{},
			StripCount: 0,
		}

		fullpath := path.Join(dirpath, e.Name())
		files, err := os.ReadDir(fullpath)
		if err != nil {
			log.Fatal(err)
		}

		for _, f := range files {
			variantName := strings.Split(f.Name(), "_")[0]
			s := r.FindStringSubmatch(f.Name())[1]
			strip, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				log.Fatal(err)
			}
			style.StripCount = int(strip)
			imgPath := path.Join(fullpath, f.Name())
			img := rl.LoadTexture(imgPath)
			if variantName == "base" {
				style.Base = img
			} else {
				style.Variants[variantName] = img
			}
		}
		styles[e.Name()] = style
	}
	return styles
}

func UnloadAnimStyles(s AnimStyles) {
	for _, style := range s {
		rl.UnloadTexture(style.Base)
		for _, variant := range style.Variants {
			rl.UnloadTexture(variant)
		}
	}
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
