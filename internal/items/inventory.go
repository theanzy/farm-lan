package items

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type InventoryItem struct {
	Item
	Quantity int
}

type Inventory struct {
	items []InventoryItem
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
		irect := ItemSlotRect(ui.container, i, ui.padding, ui.slotsize, ui.colcount)
		if rl.CheckCollisionPointRec(mpos, irect) {
			ui.InventoryId = item.Name
		}
	}
}

func (ui *InventoryUI) ItemHover(inventory *Inventory, mpos rl.Vector2) {
	hovered := false
	for i, item := range inventory.Items() {
		irect := ItemSlotRect(ui.container, i, ui.padding, ui.slotsize, ui.colcount)
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
	imgScale := tilescale * 0.8
	for i, item := range items {
		rect := ItemSlotRect(ui.container, i, padding, ui.slotsize, ui.colcount)
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
