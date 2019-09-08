package coinone

type TradeHistory struct {
	Trades []Trade `json:"trades"`
}

type Trade struct {
	Timestamp     uint64 `json:"timestamp"`
	Price         string `json:"price"`
	Volume        string `json:"volume"`
	IsSellerMaker bool   `json:"is_seller_maker"`
}
