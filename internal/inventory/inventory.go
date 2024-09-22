package inventory

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

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
			Quantity: 0,
		}, {
			Item: Item{
				Type:        "crop",
				BuyPrice:    12,
				SellPrice:   20,
				Name:        "Beetroot",
				Description: "Beetroot is a cool-season crop that thrives in well-drained soil with regular watering. It has roots and greens for harvest and sale",
				Image:       cropStrip(assets["beetroot"], 4),
			},
			Quantity: 0,
		},
		{
			Item: Item{
				Type:        "crop",
				BuyPrice:    15,
				SellPrice:   22,
				Name:        "Cabbage",
				Description: "Cabbage is a hardy, cool-season, thriving in fertile, well-drained soil with plenty of sunlight",
				Image:       cropStrip(assets["cabbage"], 4),
			},
			Quantity: 0,
		},
		{
			Item: Item{
				Type:        "crop",
				BuyPrice:    15,
				SellPrice:   25,
				Name:        "Carrot",
				Description: "Carrots are a crunchy root vegetable that thrive in loose, sandy soil. They need plenty of sunlight and regular watering for best results.",
				Image:       cropStrip(assets["carrot"], 4),
			},
			Quantity: 1,
		},
		{
			Item: Item{
				Type:        "crop",
				BuyPrice:    15,
				SellPrice:   32,
				Name:        "Cauliflower",
				Description: "Cauliflower is a white and crunchy vegetable. It likes sunny spots and lots of water to grow big and healthy.",
				Image:       cropStrip(assets["cauliflower"], 4),
			},
			Quantity: 1,
		},
		{
			Item: Item{
				Type:        "crop",
				BuyPrice:    15,
				SellPrice:   30,
				Name:        "Kale",
				Description: "Kale is a versatile leafy green superfood packed with robust flavor. Its earthy, slightly bitter taste and hearty texture make it adaptable for everything from raw salads and smoothies to comforting soups and even chips.",
				Image:       cropStrip(assets["kale"], 4),
			},
			Quantity: 1,
		},
		{
			Item: Item{
				Type:        "crop",
				BuyPrice:    15,
				SellPrice:   12,
				Name:        "Parsnip",
				Description: "A parsnip is a pale, tapered root vegetable that resembles a white carrot. The resemblance makes sense, because parsnips and carrots are cousins.",
				Image:       cropStrip(assets["parsnip"], 4),
			},
			Quantity: 1,
		},
		{
			Item: Item{
				Type:        "crop",
				BuyPrice:    15,
				SellPrice:   12,
				Name:        "Potato",
				Description: "It grows well in cool climates. Potatoes are often boiled, fried, or baked.",
				Image:       cropStrip(assets["potato"], 4),
			},
			Quantity: 5,
		},
		{
			Item: Item{
				Type:        "crop",
				BuyPrice:    15,
				SellPrice:   12,
				Name:        "Pumpkin",
				Description: "Pumpkin is a plump, nutritious orange vegetable, and a highly nutrient dense food. It is low in calories but rich in vitamins and minerals.",
				Image:       cropStrip(assets["pumpkin"], 4),
			},
			Quantity: 5,
		},
		{
			Item: Item{
				Type:        "crop",
				BuyPrice:    15,
				SellPrice:   12,
				Name:        "Radish",
				Description: "That slightly bitter, crunchy vegetable you pulled out of the garden bed is a radish. Many people love to eat sliced radishes on salads or buttered toast.",
				Image:       cropStrip(assets["radish"], 4),
			},
			Quantity: 5,
		},
		{
			Item: Item{
				Type:        "crop",
				BuyPrice:    15,
				SellPrice:   12,
				Name:        "Sunflower",
				Description: "The sunflower always faces toward the sun. The sunflower plant is 1 to 4 metres tall, but in good soil, it grows up to 5 metres.",
				Image:       cropStrip(assets["sunflower"], 4),
			},
			Quantity: 5,
		},
		{
			Item: Item{
				Type:        "crop",
				BuyPrice:    15,
				SellPrice:   12,
				Name:        "Wheat",
				Description: "A cereal grain that yields a fine white flour used chiefly in breads, baked goods, and pastas.",
				Image:       cropStrip(assets["wheat"], 4),
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
		item := i.items[idx]
		item.Quantity += quantity
		i.items[idx] = item
	}
}

func (i *Inventory) Decrease(name string, quantity int) int {
	idx := slices.IndexFunc(i.items, func(x InventoryItem) bool {
		return x.Name == name
	})
	if idx != -1 {
		item := i.items[idx]
		item.Quantity -= quantity
		if item.Quantity <= 0 {
			item.Quantity = 0
		}
		i.items[idx] = item
		return item.Quantity
	}
	return -1
}

func (i *Inventory) AvailableSeeds() []string {
	res := []string{}
	for _, item := range i.items {
		if item.Quantity > 0 && item.Type == "seed" {
			res = append(res, strings.ToLower(strings.Split(item.Name, " ")[0]))
		}
	}
	return res
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

type InventoryUI struct {
	container   rl.Rectangle
	padding     float32
	slotsize    float32
	colcount    float32
	InventoryId string
	hoverId     string
}

func NewInventoryUI(screenWidth float32, screenHeight float32, tilesize float32) InventoryUI {
	var w float32 = 800.0
	var h float32 = 600.0
	container := rl.NewRectangle(screenWidth*0.5-w*0.5, screenHeight*0.5-h*0.5, w, h)
	const padding float32 = 28.0
	slotsize := tilesize
	colcount := float32(math.Floor(float64(container.Width / (slotsize + padding))))
	return InventoryUI{
		container:   container,
		padding:     padding,
		slotsize:    slotsize,
		colcount:    colcount,
		InventoryId: "",
	}

}

func (ui *InventoryUI) ItemClick(inventory *Inventory, mpos rl.Vector2) {
	for i, item := range inventory.Items() {
		irect := InventorySlotRect(ui.container, i, ui.padding, ui.slotsize, ui.colcount)
		if rl.CheckCollisionPointRec(mpos, irect) {
			ui.InventoryId = item.Name
		}
	}
}

func (ui *InventoryUI) ItemHover(inventory *Inventory, mpos rl.Vector2) {
	hovered := false
	for i, item := range inventory.Items() {
		irect := InventorySlotRect(ui.container, i, ui.padding, ui.slotsize, ui.colcount)
		if rl.CheckCollisionPointRec(mpos, irect) && item.Name != ui.InventoryId {
			hovered = true
			ui.hoverId = item.Name
		}
	}
	if !hovered {
		ui.hoverId = ""
	}
}

func (ui *InventoryUI) DrawSlotSelection(rect rl.Rectangle, tilescale float32, uiAssets map[string]rl.Texture2D, alpha uint8) {
	shift := ui.slotsize * 0.25
	stl := rl.NewVector2(rect.X-shift, rect.Y-shift)
	str := rl.NewVector2(rect.X+ui.slotsize-shift, rect.Y-shift)
	sbl := rl.NewVector2(rect.X-shift, rect.Y+ui.slotsize-shift)
	sbr := rl.NewVector2(rect.X+ui.slotsize-shift, rect.Y+ui.slotsize-shift)
	tint := rl.NewColor(255, 255, 255, alpha)
	rl.DrawTextureEx(uiAssets["selectbox_tl"], stl, 0, tilescale, tint)
	rl.DrawTextureEx(uiAssets["selectbox_tr"], str, 0, tilescale, tint)
	rl.DrawTextureEx(uiAssets["selectbox_br"], sbr, 0, tilescale, tint)
	rl.DrawTextureEx(uiAssets["selectbox_bl"], sbl, 0, tilescale, tint)
}

func (ui *InventoryUI) Draw(inventory *Inventory, uiAssets map[string]rl.Texture2D, tilescale float32) {
	rl.DrawRectangleRec(ui.container, rl.Beige)
	rl.DrawText("Inventory", int32(ui.container.X)+20, int32(ui.container.Y)+10, 30, rl.White)
	items := inventory.Items()
	inventoryIdx := 0
	padding := ui.padding
	imgScale := tilescale * 0.5
	for i, item := range items {
		rect := InventorySlotRect(ui.container, i, padding, ui.slotsize, ui.colcount)
		rl.DrawRectangleRec(rect, rl.Brown)
		tx := rect.X + ui.slotsize*0.5 - float32(item.Image.Width)*imgScale*0.5
		ty := rect.Y + ui.slotsize*0.5 - float32(item.Image.Height)*imgScale*0.5
		rl.DrawTextureEx(item.Image, rl.NewVector2(tx, ty), 0, imgScale, rl.White)
		if ui.InventoryId == item.Name {
			inventoryIdx = i
			ui.DrawSlotSelection(rect, tilescale, uiAssets, 255)
		} else if ui.hoverId == item.Name {
			ui.DrawSlotSelection(rect, tilescale, uiAssets, 100)
		}
		// quantity
		var qfontsize int32 = 15
		qText := strconv.Itoa(item.Quantity)
		qWidth := rl.MeasureText(qText, qfontsize) + 5
		rl.DrawText(qText, int32(rect.X+ui.slotsize)-qWidth, int32(rect.Y+ui.slotsize)-qfontsize, qfontsize, rl.White)
	}
	// name
	descRect := rl.NewRectangle(ui.container.X+padding, ui.container.Y+ui.container.Height-padding-180, ui.container.Width-padding*2, 180)
	rl.DrawRectangleRec(descRect, rl.White)
	rl.DrawText(items[inventoryIdx].Name, int32(descRect.X+padding), int32(descRect.Y+padding*0.5), 25, rl.Black)

	// price
	priceText := fmt.Sprintf("$%d", items[inventoryIdx].SellPrice)
	priceTextW := rl.MeasureText(priceText, 25)
	rl.DrawText(priceText, int32(descRect.X+descRect.Width-padding-float32(priceTextW)), int32(descRect.Y+padding*0.5), 25, rl.DarkGray)

	// description
	desc := items[inventoryIdx].Description
	words := strings.Split(desc, " ")
	fontsize := 20
	descWidth := rl.MeasureText(desc, int32(fontsize))
	if descWidth > int32(descRect.Width-padding*2) {
		gap := 10
		rl.DrawText(strings.Join(words[0:gap], " "), int32(descRect.X+padding), int32(descRect.Y+padding*2), int32(fontsize), rl.Gray)
		i := 0
		h := 3
		for {
			start := i + gap
			end := start + gap
			breaking := false
			if len(words) < start {
				start = len(words)
			}
			if len(words) < end {
				end = len(words)
				breaking = true
			}
			w := strings.Join(words[start:end], " ")
			rl.DrawText(w, int32(descRect.X+padding), int32(descRect.Y+padding*float32(h)), int32(fontsize), rl.Gray)
			i += gap
			h += 1
			if breaking {
				break
			}
		}

	} else {
		rl.DrawText(desc, int32(descRect.X+padding), int32(descRect.Y+padding*2), int32(fontsize), rl.Gray)
	}
}
func InventorySlotRect(container rl.Rectangle, i int, padding float32, slotsize float32, colCount float32) rl.Rectangle {
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
