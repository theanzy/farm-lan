package anim

import (
	"fmt"
	"log"
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

type AnimConfig struct {
	Parts      []string
	StripCount int
}

var stripRegExp = regexp.MustCompile(`[a-z](\d+)\.png`)

// example: base_attack_strip10.png
//
// parts = ['base', 'attack']
//
// stripcount = 10
func parseAnimConfig(filename string) (AnimConfig, error) {
	parts := strings.Split(filename, "_")
	if len(parts) < 3 {
		return AnimConfig{}, fmt.Errorf("Invalid filename %q", filename)
	}
	s := stripRegExp.FindStringSubmatch(filename)[1]
	strip, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return AnimConfig{}, err
	}
	return AnimConfig{
		Parts:      parts[0:2],
		StripCount: int(strip),
	}, nil
}

func NewAnimStyles(dirpath string, supportedStyles []string) AnimStyles {
	styles := AnimStyles{}

	entries, err := os.ReadDir(dirpath)
	if err != nil {
		log.Fatal(err)
	}

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
			cfg, err := parseAnimConfig(f.Name())
			if err != nil {
				log.Fatal(err)
			}
			imgPath := path.Join(fullpath, f.Name())
			img := rl.LoadTexture(imgPath)
			style.StripCount = cfg.StripCount
			if cfg.Parts[0] == "base" {
				style.Base = img
			} else {
				style.Variants[cfg.Parts[0]] = img
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
