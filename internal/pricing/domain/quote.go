package domain

// QuoteItem representa un item cotizado con precio unitario y total.
type QuoteItem struct {
	ItemType  ItemType
	ItemID    string
	Qty       int
	UnitPrice Money
	LineTotal Money
}

// Quote agrupa items cotizados con subtotal.
type Quote struct {
	Items    []QuoteItem
	Subtotal Money
}
