package render

import (
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type DepthRenderer struct {
	counter      int
	sortInterval int
	Sprites      []Sprite
}

func NewDepthRenderer(sortInverval int) DepthRenderer {
	return DepthRenderer{
		counter:      sortInverval,
		sortInterval: sortInverval,
		Sprites:      []Sprite{},
	}
}

func (r *DepthRenderer) Update() {
	r.counter -= 1
	if r.counter <= 0 {
		r.counter = r.sortInterval
		slices.SortStableFunc(r.Sprites, func(a Sprite, b Sprite) int {
			return int(a.Center().Y - b.Center().Y)
		})
	}
}

func (r *DepthRenderer) Draw(offset rl.Vector2, drawRoof bool) {
	for _, sprite := range r.Sprites {
		sprite.Draw(offset, drawRoof)
	}
}

type Sprite struct {
	Draw   func(offset rl.Vector2, drawRoof bool)
	Center func() rl.Vector2
}
