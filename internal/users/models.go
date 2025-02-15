package users

type InfoResponse struct {
	Coins       int             `json:"coins" db:"coins"`
	Inventory   []InventoryItem `json:"inventory"`
	CoinHistory CoinHistory     `json:"coinHistory"`
}

type InventoryItem struct {
	Type     string `json:"type" db:"item"`
	Quantity int    `json:"quantity" db:"quantity"`
}

type CoinHistory struct {
	Received []CoinTransaction `json:"received"`
	Sent     []CoinTransaction `json:"sent"`
}

type CoinTransaction struct {
	FromUser string `json:"fromUser,omitempty" db:"sender_id"`
	ToUser   string `json:"toUser,omitempty" db:"receiver_id"`
	Amount   int    `json:"amount" db:"amount"`
}

type ErrorResponse struct {
	Errors string `json:"errors"`
}
