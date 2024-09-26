package items

import (
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ShopItem struct {
	Item
	Quantity int
}

type Shop struct {
	name  string
	Items []ShopItem
}

func NewShop(name string, items []Item, quantities map[string]int) Shop {
	sitems := []ShopItem{}
	for _, item := range items {
		quantity := 0
		if q, ok := quantities[item.Name]; ok {
			quantity = q
		}
		sitems = append(sitems, ShopItem{Item: item, Quantity: quantity})
	}
	return Shop{name: name, Items: sitems}
}

func NewSeedShop(name string, items []Item) Shop {
	q := map[string]int{}
	seeds := []Item{}
	for _, item := range items {
		if item.Type == "seed" {
			q[item.Name] = 99
			seeds = append(seeds, item)
		}
	}
	return NewShop(name, seeds, q)
}

func (s *Shop) Increase(name string, quantity int) {
	idx := slices.IndexFunc(s.Items, func(x ShopItem) bool {
		return x.Name == name
	})
	if idx != -1 {
		item := s.Items[idx]
		item.Quantity += quantity
		s.Items[idx] = item
	}
}

func (s *Shop) Decrease(name string, quantity int) int {
	idx := slices.IndexFunc(s.Items, func(x ShopItem) bool {
		return x.Name == name
	})
	if idx != -1 {
		item := s.Items[idx]
		item.Quantity -= quantity
		if item.Quantity <= 0 {
			item.Quantity = 0
		}
		s.Items[idx] = item
		return item.Quantity
	}
	return -1
}

type Selection struct {
	side string
	id   string
}

type ShopUI struct {
	container          rl.Rectangle
	inventoryContainer rl.Rectangle
	shopContainer      rl.Rectangle
	padding            float32
	slotsize           float32
	colcount           float32
	selection          Selection
	hoverId            Selection
	selectionRect      rl.Rectangle
	hoverRect          rl.Rectangle
}

func NewShopUI(screenSize rl.Vector2, tilesize float32) ShopUI {
	var w float32 = 1000
	var h float32 = 600

	container := rl.NewRectangle(screenSize.X*0.5-w*0.5, screenSize.Y*0.5-h*0.5, w, h)

	const padding float32 = 28.0
	slotsize := tilesize
	colcount := 6

	sectionWidth := padding*float32(colcount) + slotsize*float32(colcount)
	inventoryContainer := rl.NewRectangle(container.X, container.Y, sectionWidth, container.Height)

	shopX := container.X + container.Width - sectionWidth - padding
	shopContainer := rl.NewRectangle(shopX, container.Y, container.Width, container.Height)
	return ShopUI{
		container:          container,
		padding:            padding,
		slotsize:           slotsize,
		colcount:           float32(colcount),
		inventoryContainer: inventoryContainer,
		shopContainer:      shopContainer,
	}
}

func (ui *ShopUI) Click(mpos rl.Vector2, inventory *Inventory, shop *Shop) {
	for i, item := range inventory.Items() {
		rect := itemSlotRect(ui.inventoryContainer, i, ui.padding, ui.slotsize, ui.colcount)
		if rl.CheckCollisionPointRec(mpos, rect) {
			ui.selection.id = item.Name
			ui.selection.side = "inventory"
			ui.selectionRect = rect
			return
		}
	}
	for i, item := range shop.Items {
		rect := itemSlotRect(ui.shopContainer, i, ui.padding, ui.slotsize, ui.colcount)
		if rl.CheckCollisionPointRec(mpos, rect) {
			ui.selection.id = item.Name
			ui.selection.side = "shop"
			ui.selectionRect = rect
			return
		}
	}
}

func (ui *ShopUI) Draw(shop *Shop, inventory *Inventory, uiAssets map[string]rl.Texture2D, tilescale float32) {
	lineColor := rl.NewColor(rl.Beige.R-20, rl.Beige.G-20, rl.Beige.B-20, 255)
	rl.DrawRectangleRec(ui.container, rl.Beige)
	rl.DrawRectangleLinesEx(ui.container, 2, lineColor)

	ui.drawInventory(inventory, ui.inventoryContainer, uiAssets, tilescale)

	midX := ui.container.X + ui.container.Width*0.5
	rl.DrawLineEx(rl.NewVector2(midX, ui.container.Y), rl.NewVector2(midX, ui.container.Y+ui.container.Height), 2, lineColor)

	// shop
	ui.drawShop(shop, ui.shopContainer, uiAssets, tilescale)
	if ui.selection.id != "" {
		drawSlotSelection(ui.selectionRect, tilescale, uiAssets, 255)
	}
	if ui.hoverId.id != "" && ui.hoverId.id != ui.selection.id {
		drawSlotSelection(ui.hoverRect, tilescale, uiAssets, 100)

	}
}

func (ui *ShopUI) ItemHover(mpos rl.Vector2, inventory *Inventory, shop *Shop) {
	for i, item := range inventory.Items() {
		rect := itemSlotRect(ui.inventoryContainer, i, ui.padding, ui.slotsize, ui.colcount)
		if rl.CheckCollisionPointRec(mpos, rect) {
			ui.hoverId.id = item.Name
			ui.hoverId.side = "inventory"
			ui.hoverRect = rect
			return
		}
	}
	for i, item := range shop.Items {
		rect := itemSlotRect(ui.shopContainer, i, ui.padding, ui.slotsize, ui.colcount)
		if rl.CheckCollisionPointRec(mpos, rect) {
			ui.hoverId.id = item.Name
			ui.hoverId.side = "shop"
			ui.hoverRect = rect
			return
		}
	}
}

func (ui *ShopUI) drawInventory(inventory *Inventory, container rl.Rectangle, uiAssets map[string]rl.Texture2D, scale float32) {
	rl.DrawText("Inventory", int32(container.X)+20, int32(container.Y)+10, 30, rl.White)
	items := inventory.Items()
	padding := ui.padding
	for i, item := range items {
		rect := itemSlotRect(container, i, padding, ui.slotsize, ui.colcount)
		DrawItem(rect, item.Image, scale, item.Quantity)
	}
}

func (ui *ShopUI) drawShop(shop *Shop, container rl.Rectangle, uiAssets map[string]rl.Texture2D, scale float32) {
	rl.DrawText(shop.name, int32(container.X)+20, int32(container.Y)+10, 30, rl.White)
	items := shop.Items
	padding := ui.padding
	for i, item := range items {
		rect := itemSlotRect(container, i, padding, ui.slotsize, ui.colcount)
		DrawItem(rect, item.Image, scale, item.Quantity)
	}
}
