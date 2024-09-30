package items

import (
	"fmt"
	"math"
	"slices"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type InventoryItem struct {
	Item
	Quantity int
}

type Inventory struct {
	items   []InventoryItem
	deposit float32
}

func NewInventory(items []Item) Inventory {
	iItems := []InventoryItem{}
	for _, item := range items {
		q := 0
		if item.Name == "Wheat seed" {
			q = 5
		}
		iItems = append(iItems, InventoryItem{Item: item, Quantity: q})
	}
	return Inventory{items: iItems}
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

func (i *Inventory) Count(name string) int {
	result := 0
	for _, item := range i.items {
		if item.Name == name {
			result = item.Quantity
		}
	}
	return result
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
		irect := itemSlotRect(ui.container, i, ui.padding, ui.slotsize, ui.colcount)
		if rl.CheckCollisionPointRec(mpos, irect) {
			ui.InventoryId = item.Name
		}
	}
}

func (ui *InventoryUI) ItemHover(inventory *Inventory, mpos rl.Vector2) {
	hovered := false
	for i, item := range inventory.Items() {
		irect := itemSlotRect(ui.container, i, ui.padding, ui.slotsize, ui.colcount)
		if rl.CheckCollisionPointRec(mpos, irect) && item.Name != ui.InventoryId {
			hovered = true
			ui.hoverId = item.Name
		}
	}
	if !hovered {
		ui.hoverId = ""
	}
}

func (ui *InventoryUI) Draw(inventory *Inventory, uiAssets map[string]rl.Texture2D, tilescale float32) {
	lineColor := rl.NewColor(rl.Beige.R-20, rl.Beige.G-20, rl.Beige.B-20, 255)
	rl.DrawRectangleRec(ui.container, rl.Beige)
	rl.DrawRectangleLinesEx(ui.container, 2, lineColor)
	rl.DrawText("Inventory", int32(ui.container.X)+20, int32(ui.container.Y)+10, 30, rl.White)
	items := inventory.Items()
	inventoryIdx := -1
	padding := ui.padding
	imgScale := tilescale
	for i, item := range items {
		rect := itemSlotRect(ui.container, i, padding, ui.slotsize, ui.colcount)
		DrawItem(rect, item.Image, imgScale, item.Quantity)
		if ui.InventoryId == item.Name {
			inventoryIdx = i
			drawSlotSelection(rect, tilescale, uiAssets, 255)
		} else if ui.hoverId == item.Name {
			drawSlotSelection(rect, tilescale, uiAssets, 100)
		}
	}
	if inventoryIdx >= 0 {

		// name
		descRect := rl.NewRectangle(ui.container.X+padding, ui.container.Y+ui.container.Height-padding-180, ui.container.Width-padding*2, 180)
		rl.DrawRectangleRec(descRect, rl.White)
		rl.DrawText(items[inventoryIdx].Name, int32(descRect.X+padding), int32(descRect.Y+padding*0.5), 25, rl.Black)

		// price
		priceText := fmt.Sprintf("$%d", items[inventoryIdx].SellPrice)
		priceTextW := rl.MeasureText(priceText, 25)
		rl.DrawText(priceText, int32(descRect.X+descRect.Width-padding-float32(priceTextW)), int32(descRect.Y+padding*0.5), 25, rl.DarkGray)

		// description
		DrawMultilineText(
			items[inventoryIdx].Description,
			rl.NewVector2(descRect.X+padding, descRect.Y+padding*2),
			20,
			int32(descRect.Width-6*padding),
			8,
		)
	}
}
