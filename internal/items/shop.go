package items

import (
	"fmt"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/theanzy/farmsim/internal/ui"
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
	footerContainer    rl.Rectangle
	button             ui.TextButton
	increaseButton     ui.ImgButton
	decreaseButton     ui.ImgButton
	quantity           int
}

func NewShopUI(screenSize rl.Vector2, tilesize float32, uiAssets map[string]rl.Texture2D) ShopUI {
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
	footerContainer := rl.NewRectangle(container.X, container.Y+container.Height-150, container.Width, 150)
	btnRect := rl.NewRectangle(
		footerContainer.X+footerContainer.Width-padding-150,
		footerContainer.Y+footerContainer.Height-padding*0.5-40,
		150,
		40,
	)
	btn := ui.NewTextButton(btnRect, "BUY", 20, rl.Blue)

	rightArrow := uiAssets["arrow_right"]
	leftArrow := uiAssets["arrow_left"]
	buttonScale := tilesize / float32(rightArrow.Height) * 0.6

	increaseButton := ui.NewImgButton(
		rl.NewVector2(
			btn.Rect.X+btn.Rect.Width-float32(rightArrow.Width)*buttonScale,
			btn.Rect.Y-padding*0.25-float32(rightArrow.Height)*buttonScale,
		),
		rightArrow,
		buttonScale,
	)

	decreaseButton := ui.NewImgButton(
		rl.NewVector2(
			btn.Rect.X,
			btn.Rect.Y-padding*0.25-float32(leftArrow.Height)*buttonScale,
		),
		leftArrow,
		buttonScale,
	)
	return ShopUI{
		container:          container,
		padding:            padding,
		slotsize:           slotsize,
		colcount:           float32(colcount),
		inventoryContainer: inventoryContainer,
		shopContainer:      shopContainer,
		footerContainer:    footerContainer,
		button:             btn,
		increaseButton:     increaseButton,
		decreaseButton:     decreaseButton,
		quantity:           1,
	}
}

func (ui *ShopUI) Click(mpos rl.Vector2, inventory *Inventory, shop *Shop) {
	for i, item := range inventory.Items() {
		rect := itemSlotRect(ui.inventoryContainer, i, ui.padding, ui.slotsize, ui.colcount)
		if rl.CheckCollisionPointRec(mpos, rect) {
			ui.selection.id = item.Name
			ui.selection.side = "inventory"
			ui.selectionRect = rect
			ui.button.SetText("SELL")
			ui.button.BgColor = rl.Red
			ui.quantity = 1
			return
		}
	}
	for i, item := range shop.Items {
		rect := itemSlotRect(ui.shopContainer, i, ui.padding, ui.slotsize, ui.colcount)
		if rl.CheckCollisionPointRec(mpos, rect) {
			ui.selection.id = item.Name
			ui.selection.side = "shop"
			ui.selectionRect = rect
			ui.button.SetText("BUY")
			ui.button.BgColor = rl.Blue
			ui.quantity = 1
			return
		}
	}
	if rl.CheckCollisionPointRec(mpos, ui.button.Rect) {
		ui.button.Press()
	}
	if rl.CheckCollisionPointRec(mpos, ui.increaseButton.Rect) {
		maxQuantity := 1
		if ui.selection.side == "inventory" {
			if idx := slices.IndexFunc(inventory.Items(), func(x InventoryItem) bool { return x.Name == ui.selection.id }); idx != -1 {
				maxQuantity = inventory.Items()[idx].Quantity
			}

		}
		if ui.selection.side == "shop" {
			if idx := slices.IndexFunc(shop.Items, func(x ShopItem) bool { return x.Name == ui.selection.id }); idx != -1 {
				maxQuantity = shop.Items[idx].Quantity
			}
		}
		ui.quantity = ui.quantity + 1
		if ui.quantity > maxQuantity {
			ui.quantity = maxQuantity
		}
		ui.increaseButton.Press()
	}
	if rl.CheckCollisionPointRec(mpos, ui.decreaseButton.Rect) {
		ui.quantity = max(1, ui.quantity-1)
		ui.decreaseButton.Press()
	}
}

func (ui *ShopUI) Update() {
	ui.button.Update()
	ui.increaseButton.Update()
	ui.decreaseButton.Update()
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
		rl.DrawRectangleRec(ui.footerContainer, rl.White)
		rl.DrawRectangleLinesEx(ui.footerContainer, 2, lineColor)
		// name
		if ui.selection.side == "shop" {
			if idx := slices.IndexFunc(shop.Items, func(x ShopItem) bool {
				return x.Name == ui.selection.id
			}); idx != -1 {
				item := shop.Items[idx]
				drawShopFooter(
					ui.footerContainer,
					item.Name,
					item.Description,
					float32(item.BuyPrice),
					float32(ui.quantity),
					ui.padding,
					&ui.button,
					&ui.increaseButton,
					&ui.decreaseButton,
				)
			}
		} else if ui.selection.side == "inventory" {
			if idx := slices.IndexFunc(inventory.Items(), func(x InventoryItem) bool {
				return x.Name == ui.selection.id
			}); idx != -1 {
				item := inventory.Items()[idx]
				drawShopFooter(
					ui.footerContainer,
					item.Name,
					item.Description,
					float32(item.BuyPrice),
					float32(ui.quantity),
					ui.padding,
					&ui.button,
					&ui.increaseButton,
					&ui.decreaseButton,
				)
			}

		}
	}
	if ui.hoverId.id != "" && ui.hoverId.id != ui.selection.id {
		drawSlotSelection(ui.hoverRect, tilescale, uiAssets, 100)
	}

}

func drawShopFooter(container rl.Rectangle, name string, description string, price float32, quantity float32, padding float32, button *ui.TextButton, increaseButton *ui.ImgButton, decreaseButton *ui.ImgButton) {
	rl.DrawText(name, int32(container.X+padding), int32(container.Y+padding), 20, rl.Black)
	// price
	priceText := fmt.Sprintf("$%0.f", price*quantity)
	var priceFontsize int32 = 20
	priceTextWidth := rl.MeasureText(priceText, priceFontsize)
	rl.DrawText(
		priceText,
		int32(container.X+container.Width-padding-float32(priceTextWidth)),
		int32(container.Y+padding),
		priceFontsize,
		rl.Black,
	)

	// quantity
	quantityRect := rl.NewRectangle(
		button.Rect.X+button.Rect.Width*0.5-35,
		button.Rect.Y-padding*0.25-float32(increaseButton.Rect.Height),
		70,
		30,
	)
	rl.DrawRectangleRec(quantityRect, rl.RayWhite)
	qText := fmt.Sprintf("%0.f", quantity)
	var qFontSize int32 = 18
	qTextWidth := rl.MeasureText(qText, qFontSize)
	rl.DrawText(
		qText,
		int32(quantityRect.X+quantityRect.Width*0.5-float32(qTextWidth)*0.5),
		int32(quantityRect.Y+quantityRect.Height*0.5-float32(qFontSize)*0.5),
		qFontSize,
		rl.Black,
	)

	// increase button
	increaseButton.Draw()

	// decrease button
	decreaseButton.Draw()

	// description
	descRect := rl.NewRectangle(container.X, container.Y+container.Height-180, container.Width-padding*2, 180)
	DrawMultilineText(
		description,
		rl.NewVector2(descRect.X+padding, descRect.Y+padding*3.5),
		19,
		int32(descRect.Width-5*padding-button.Rect.Width),
		8,
	)
	button.Draw()
}

func (ui *ShopUI) ItemHover(mpos rl.Vector2, inventory *Inventory, shop *Shop) {
	ui.hoverId.id = ""
	ui.hoverId.side = ""
	for i, item := range inventory.Items() {
		rect := itemSlotRect(ui.inventoryContainer, i, ui.padding, ui.slotsize, ui.colcount)
		if rl.CheckCollisionPointRec(mpos, rect) {
			ui.hoverId.id = item.Name
			ui.hoverId.side = "inventory"
			ui.hoverRect = rect
			return
		} else {

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
