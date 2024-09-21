package inventory

import (
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/theanzy/farmsim/internal/crop"
)

type Item struct {
	Type        string
	BuyPrice    int
	SellPrice   int
	Name        string
	Description string
	Image       rl.Texture2D
}

type InventoryItem struct {
	Item
	Quantity int
}

type Inventory struct {
	items []InventoryItem
}

func cropStrip(img crop.StripImg, idx int) rl.Texture2D {
	image := rl.LoadImageFromTexture(img.Img)
	defer rl.UnloadImage(image)
	rl.ImageCrop(image, img.SrcRects[idx])
	return rl.LoadTextureFromImage(image)
}

func NewInventory(assets map[string]crop.StripImg) Inventory {
	items := []InventoryItem{
		{
			Item: Item{
				Type:        "seed",
				BuyPrice:    12,
				SellPrice:   10,
				Name:        "Beetroot seed",
				Description: "Beetroot is a cool-season crop that thrives in well-drained soil with regular watering. It has roots and greens for harvest and sale",
				Image:       cropStrip(assets["beetroot"], 3),
			},
			Quantity: 0,
		},
		{
			Item: Item{
				Type:        "seed",
				BuyPrice:    15,
				SellPrice:   12,
				Name:        "Cabbage seed",
				Description: "Cabbage is a hardy, cool-season, thriving in fertile, well-drained soil with plenty of sunlight",
				Image:       cropStrip(assets["cabbage"], 3),
			},
			Quantity: 0,
		},
		{
			Item: Item{
				Type:        "seed",
				BuyPrice:    15,
				SellPrice:   12,
				Name:        "Carrot seed",
				Description: "Carrots are a crunchy root vegetable that thrive in loose, sandy soil. They need plenty of sunlight and regular watering for best results.",
				Image:       cropStrip(assets["carrot"], 3),
			},
			Quantity: 1,
		},
		{
			Item: Item{
				Type:        "seed",
				BuyPrice:    15,
				SellPrice:   12,
				Name:        "Cauliflower seed",
				Description: "Cauliflower is a white and crunchy vegetable. It likes sunny spots and lots of water to grow big and healthy.",
				Image:       cropStrip(assets["cauliflower"], 3),
			},
			Quantity: 1,
		},
		{
			Item: Item{
				Type:        "seed",
				BuyPrice:    15,
				SellPrice:   12,
				Name:        "Kale seed",
				Description: "Kale is a versatile leafy green superfood packed with robust flavor. Its earthy, slightly bitter taste and hearty texture make it adaptable for everything from raw salads and smoothies to comforting soups and even chips.",
				Image:       cropStrip(assets["kale"], 3),
			},
			Quantity: 1,
		},
		{
			Item: Item{
				Type:        "seed",
				BuyPrice:    15,
				SellPrice:   12,
				Name:        "Parsnip seed",
				Description: "A parsnip is a pale, tapered root vegetable that resembles a white carrot. The resemblance makes sense, because parsnips and carrots are cousins.",
				Image:       cropStrip(assets["parsnip"], 3),
			},
			Quantity: 1,
		},
		{
			Item: Item{
				Type:        "seed",
				BuyPrice:    15,
				SellPrice:   12,
				Name:        "Potato seed",
				Description: "It grows well in cool climates. Potatoes are often boiled, fried, or baked.",
				Image:       cropStrip(assets["potato"], 3),
			},
			Quantity: 5,
		},
		{
			Item: Item{
				Type:        "seed",
				BuyPrice:    15,
				SellPrice:   12,
				Name:        "Pumpkin seed",
				Description: "Pumpkin is a plump, nutritious orange vegetable, and a highly nutrient dense food. It is low in calories but rich in vitamins and minerals.",
				Image:       cropStrip(assets["pumpkin"], 3),
			},
			Quantity: 5,
		},
		{
			Item: Item{
				Type:        "seed",
				BuyPrice:    15,
				SellPrice:   12,
				Name:        "Radish seed",
				Description: "That slightly bitter, crunchy vegetable you pulled out of the garden bed is a radish. Many people love to eat sliced radishes on salads or buttered toast.",
				Image:       cropStrip(assets["radish"], 3),
			},
			Quantity: 5,
		},
		{
			Item: Item{
				Type:        "seed",
				BuyPrice:    15,
				SellPrice:   12,
				Name:        "Sunflower seed",
				Description: "The sunflower always faces toward the sun. The sunflower plant is 1 to 4 metres tall, but in good soil, it grows up to 5 metres.",
				Image:       cropStrip(assets["sunflower"], 3),
			},
			Quantity: 5,
		},
		{
			Item: Item{
				Type:        "seed",
				BuyPrice:    15,
				SellPrice:   12,
				Name:        "Wheat seed",
				Description: "A cereal grain that yields a fine white flour used chiefly in breads, baked goods, and pastas.",
				Image:       cropStrip(assets["wheat"], 3),
			},
			Quantity: 5,
		},
	}
	return Inventory{items: items}
}

func (i *Inventory) DeinitInventory() {
	for _, x := range i.items {
		rl.UnloadTexture(x.Image)
	}
}

func (i *Inventory) Increase(name string, quantity int) {
	idx := slices.IndexFunc(i.items, func(x InventoryItem) bool {
		return x.Name == name
	})
	if idx != -1 {
		i.items[idx].Quantity += quantity
	}
}

func (i *Inventory) Decrease(name string, quantity int) {
	idx := slices.IndexFunc(i.items, func(x InventoryItem) bool {
		return x.Name == name
	})
	if idx != -1 {
		i.items[idx].Quantity -= quantity
		if i.items[idx].Quantity <= 0 {
			i.items[idx].Quantity = 0
		}
	}
}

func (i *Inventory) Items() []InventoryItem {
	res := []InventoryItem{}
	for _, item := range i.items {
		if item.Quantity > 0 {
			res = append(res, item)
		}
	}
	return res
}
