package items

import (
	"math"
	"strconv"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/theanzy/farmsim/internal/strip"
)

type Item struct {
	Type        string
	BuyPrice    int
	SellPrice   int
	Name        string
	Description string
	Image       rl.Texture2D
}

func cropStrip(img strip.StripImg, idx int) rl.Texture2D {
	image := rl.LoadImageFromTexture(img.Img)
	defer rl.UnloadImage(image)
	rl.ImageCrop(image, img.SrcRects[idx])
	return rl.LoadTextureFromImage(image)
}

func LoadItems(assets map[string]strip.StripImg) []Item {
	items := []Item{
		{
			Type:        "seed",
			BuyPrice:    12,
			SellPrice:   10,
			Name:        "Beetroot seed",
			Description: "Beetroot is a cool-season crop that thrives in well-drained soil with regular watering. It has roots and greens for harvest and sale",
			Image:       cropStrip(assets["beetroot"], 3),
		},
		{
			Type:        "seed",
			BuyPrice:    15,
			SellPrice:   12,
			Name:        "Cabbage seed",
			Description: "Cabbage is a hardy, cool-season, thriving in fertile, well-drained soil with plenty of sunlight",
			Image:       cropStrip(assets["cabbage"], 3),
		},
		{
			Type:        "seed",
			BuyPrice:    15,
			SellPrice:   12,
			Name:        "Carrot seed",
			Description: "Carrots are a crunchy root vegetable that thrive in loose, sandy soil. They need plenty of sunlight and regular watering for best results.",
			Image:       cropStrip(assets["carrot"], 3),
		},
		{
			Type:        "seed",
			BuyPrice:    15,
			SellPrice:   12,
			Name:        "Cauliflower seed",
			Description: "Cauliflower is a white and crunchy vegetable. It likes sunny spots and lots of water to grow big and healthy.",
			Image:       cropStrip(assets["cauliflower"], 3),
		},
		{
			Type:        "seed",
			BuyPrice:    15,
			SellPrice:   12,
			Name:        "Kale seed",
			Description: "Kale is a versatile leafy green superfood packed with robust flavor. Its earthy, slightly bitter taste and hearty texture make it adaptable for everything from raw salads and smoothies to comforting soups and even chips.",
			Image:       cropStrip(assets["kale"], 3),
		},
		{
			Type:        "seed",
			BuyPrice:    15,
			SellPrice:   12,
			Name:        "Parsnip seed",
			Description: "A parsnip is a pale, tapered root vegetable that resembles a white carrot. The resemblance makes sense, because parsnips and carrots are cousins.",
			Image:       cropStrip(assets["parsnip"], 3),
		},
		{
			Type:        "seed",
			BuyPrice:    15,
			SellPrice:   12,
			Name:        "Potato seed",
			Description: "It grows well in cool climates. Potatoes are often boiled, fried, or baked.",
			Image:       cropStrip(assets["potato"], 3),
		},
		{
			Type:        "seed",
			BuyPrice:    15,
			SellPrice:   12,
			Name:        "Pumpkin seed",
			Description: "Pumpkin is a plump, nutritious orange vegetable, and a highly nutrient dense food. It is low in calories but rich in vitamins and minerals.",
			Image:       cropStrip(assets["pumpkin"], 3),
		},
		{
			Type:        "seed",
			BuyPrice:    15,
			SellPrice:   12,
			Name:        "Radish seed",
			Description: "That slightly bitter, crunchy vegetable you pulled out of the garden bed is a radish. Many people love to eat sliced radishes on salads or buttered toast.",
			Image:       cropStrip(assets["radish"], 3),
		},
		{
			Type:        "seed",
			BuyPrice:    15,
			SellPrice:   12,
			Name:        "Sunflower seed",
			Description: "The sunflower always faces toward the sun. The sunflower plant is 1 to 4 metres tall, but in good soil, it grows up to 5 metres.",
			Image:       cropStrip(assets["sunflower"], 3),
		},
		{
			Type:        "seed",
			BuyPrice:    15,
			SellPrice:   12,
			Name:        "Wheat seed",
			Description: "A cereal grain that yields a fine white flour used chiefly in breads, baked goods, and pastas.",
			Image:       cropStrip(assets["wheat"], 3),
		}, {
			Type:        "crop",
			BuyPrice:    12,
			SellPrice:   20,
			Name:        "Beetroot",
			Description: "Beetroot is a cool-season crop that thrives in well-drained soil with regular watering. It has roots and greens for harvest and sale",
			Image:       cropStrip(assets["beetroot"], 4),
		},
		{
			Type:        "crop",
			BuyPrice:    15,
			SellPrice:   22,
			Name:        "Cabbage",
			Description: "Cabbage is a hardy, cool-season, thriving in fertile, well-drained soil with plenty of sunlight",
			Image:       cropStrip(assets["cabbage"], 4),
		},
		{
			Type:        "crop",
			BuyPrice:    15,
			SellPrice:   25,
			Name:        "Carrot",
			Description: "Carrots are a crunchy root vegetable that thrive in loose, sandy soil. They need plenty of sunlight and regular watering for best results.",
			Image:       cropStrip(assets["carrot"], 4),
		},
		{
			Type:        "crop",
			BuyPrice:    15,
			SellPrice:   32,
			Name:        "Cauliflower",
			Description: "Cauliflower is a white and crunchy vegetable. It likes sunny spots and lots of water to grow big and healthy.",
			Image:       cropStrip(assets["cauliflower"], 4),
		},
		{
			Type:        "crop",
			BuyPrice:    15,
			SellPrice:   30,
			Name:        "Kale",
			Description: "Kale is a versatile leafy green superfood packed with robust flavor. Its earthy, slightly bitter taste and hearty texture make it adaptable for everything from raw salads and smoothies to comforting soups and even chips.",
			Image:       cropStrip(assets["kale"], 4),
		},
		{
			Type:        "crop",
			BuyPrice:    15,
			SellPrice:   12,
			Name:        "Parsnip",
			Description: "A parsnip is a pale, tapered root vegetable that resembles a white carrot. The resemblance makes sense, because parsnips and carrots are cousins.",
			Image:       cropStrip(assets["parsnip"], 4),
		},
		{
			Type:        "crop",
			BuyPrice:    15,
			SellPrice:   12,
			Name:        "Potato",
			Description: "It grows well in cool climates. Potatoes are often boiled, fried, or baked.",
			Image:       cropStrip(assets["potato"], 4),
		},
		{
			Type:        "crop",
			BuyPrice:    15,
			SellPrice:   12,
			Name:        "Pumpkin",
			Description: "Pumpkin is a plump, nutritious orange vegetable, and a highly nutrient dense food. It is low in calories but rich in vitamins and minerals.",
			Image:       cropStrip(assets["pumpkin"], 4),
		},
		{
			Type:        "crop",
			BuyPrice:    15,
			SellPrice:   12,
			Name:        "Radish",
			Description: "That slightly bitter, crunchy vegetable you pulled out of the garden bed is a radish. Many people love to eat sliced radishes on salads or buttered toast.",
			Image:       cropStrip(assets["radish"], 4),
		},
		{
			Type:        "crop",
			BuyPrice:    15,
			SellPrice:   12,
			Name:        "Sunflower",
			Description: "The sunflower always faces toward the sun. The sunflower plant is 1 to 4 metres tall, but in good soil, it grows up to 5 metres.",
			Image:       cropStrip(assets["sunflower"], 4),
		},
		{
			Type:        "crop",
			BuyPrice:    15,
			SellPrice:   12,
			Name:        "Wheat",
			Description: "A cereal grain that yields a fine white flour used chiefly in breads, baked goods, and pastas.",
			Image:       cropStrip(assets["wheat"], 4),
		},
		{
			Type:        "wood",
			BuyPrice:    15,
			SellPrice:   12,
			Name:        "Wood",
			Description: "Used for building",
			Image:       cropStrip(assets["wood"], 0),
		},
	}
	return items
}

func UnloadItems(items []Item) {
	for _, item := range items {
		rl.UnloadTexture(item.Image)
	}
}

func DrawItem(rect rl.Rectangle, img rl.Texture2D, scale float32, quantity int) {
	rl.DrawRectangleRec(rect, rl.Brown)
	slotsize := rect.Width
	tx := rect.X + slotsize*0.5 - float32(img.Width)*scale*0.5
	ty := rect.Y + slotsize*0.5 - float32(img.Height)*scale*0.5
	rl.DrawTextureEx(img, rl.NewVector2(tx, ty), 0, scale, rl.White)

	// quantity
	var qfontsize int32 = 15
	qText := strconv.Itoa(quantity)
	qWidth := rl.MeasureText(qText, qfontsize) + 5
	rl.DrawText(qText, int32(rect.X+slotsize)-qWidth, int32(rect.Y+slotsize)-qfontsize, qfontsize, rl.White)
}

func drawSlotSelection(rect rl.Rectangle, scale float32, uiAssets map[string]rl.Texture2D, alpha uint8) {
	shift := rect.Width * 0.25
	stl := rl.NewVector2(rect.X-shift, rect.Y-shift)
	str := rl.NewVector2(rect.X+rect.Width-shift, rect.Y-shift)
	sbl := rl.NewVector2(rect.X-shift, rect.Y+rect.Height-shift)
	sbr := rl.NewVector2(rect.X+rect.Width-shift, rect.Y+rect.Height-shift)
	tint := rl.NewColor(255, 255, 255, alpha)
	rl.DrawTextureEx(uiAssets["selectbox_tl"], stl, 0, scale, tint)
	rl.DrawTextureEx(uiAssets["selectbox_tr"], str, 0, scale, tint)
	rl.DrawTextureEx(uiAssets["selectbox_br"], sbr, 0, scale, tint)
	rl.DrawTextureEx(uiAssets["selectbox_bl"], sbl, 0, scale, tint)
}

func itemSlotRect(container rl.Rectangle, i int, padding float32, slotsize float32, colCount float32) rl.Rectangle {
	x := container.X + padding + ((padding + slotsize) * (float32(math.Mod(float64(i), float64(colCount)))))
	y := container.Y + padding*2 + ((padding + slotsize) * (float32(math.Floor(float64(i) / float64(colCount)))))
	rect := rl.NewRectangle(x, y, slotsize, slotsize)
	return rect
}

func CropToSeedName(cropName string) string {
	return strings.ToUpper(cropName[0:1]) + cropName[1:] + " seed"
}

func CropToCropName(cropName string) string {
	return strings.ToUpper(cropName[0:1]) + cropName[1:]
}
