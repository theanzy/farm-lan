package crop

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/theanzy/farmsim/internal/strip"
)

func LoadCropAssets(dirpath string, crops []string) (map[string]strip.StripImg, error) {
	r := regexp.MustCompile(`[a-z](\d+)\.png`)

	res := map[string]strip.StripImg{}
	err := filepath.Walk(dirpath, func(path string, info fs.FileInfo, err error) error {
		if filepath.Ext(path) == ".png" {
			ftokens := strings.Split(info.Name(), "_")
			cropName := ftokens[0]
			if slices.Contains(crops, cropName) {
				s := r.FindStringSubmatch(info.Name())[1]
				stripCount, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					return err
				}
				res[cropName] = strip.NewStripImg(rl.LoadTexture(path), int32(stripCount))
			}
		}
		return nil
	})
	if err != nil {
		return map[string]strip.StripImg{}, err
	}

	return res, nil
}
