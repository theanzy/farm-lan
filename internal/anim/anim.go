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

type AnimStyle = struct {
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
			if variantName == "base" {
				style.Base = rl.LoadTexture(imgPath)
			} else {
				style.Variants[variantName] = rl.LoadTexture(imgPath)
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

type Animation struct {
	AssetSize    rl.Vector2
	Image        rl.Texture2D
	ImageFlipped rl.Texture2D
	X            float32
	Speed        float32
	StripCount   float32
}

func (a *Animation) Update(dt float32) {
	a.X += dt * a.Speed
	if a.X >= a.StripCount {
		a.X = 0
	}
}
func (a *Animation) Reset() {
	a.X = 0
}

func (a Animation) SrcRect() rl.Rectangle {
	x := float32(math.Floor(float64(a.X)))
	return rl.NewRectangle(x*a.AssetSize.X, 0, a.AssetSize.X, a.AssetSize.Y)
}
