package crop

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type StripImg struct {
	Img        rl.Texture2D
	StripCount int
	SrcRects   []rl.Rectangle
}

func LoadCropAssets(dirpath string, crops []string) (map[string]StripImg, error) {
	r := regexp.MustCompile(`[a-z](\d+)\.png`)

	res := map[string]StripImg{}
	err := filepath.Walk(dirpath, func(path string, info fs.FileInfo, err error) error {
		if filepath.Ext(path) == ".png" {
			ftokens := strings.Split(info.Name(), "_")
			cropName := ftokens[0]
			if slices.Contains(crops, cropName) {
				s := r.FindStringSubmatch(info.Name())[1]
				strip, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					return err
				}
				stripCount := int(strip)
				img := rl.LoadTexture(path)
				rects := []rl.Rectangle{}
				unitWidth := img.Width / int32(stripCount)
				for i := range stripCount {
					rect := rl.NewRectangle(float32(i)*float32(unitWidth), 0, float32(unitWidth), float32(img.Height))
					rects = append(rects, rect)
				}
				res[cropName] = StripImg{
					Img:        rl.LoadTexture(path),
					StripCount: stripCount,
					SrcRects:   rects,
				}
			}
		}
		return nil
	})
	if err != nil {
		return map[string]StripImg{}, err
	}

	return res, nil
}

func UnloadCropAssets(assets map[string]StripImg) {
	for _, stripImg := range assets {
		rl.UnloadTexture(stripImg.Img)
	}
}
