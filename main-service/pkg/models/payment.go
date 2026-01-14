// Package models содержит структуры данных заказов
package models

// Payment содержит информацию об оплате
type Payment struct {
	Transaction  string `json:"transaction" validate:"required,min=10,max=50"`
	RequestID    string `json:"request_id" validate:"max=50"`
	Currency     string `json:"currency" validate:"required,iso4217"`
	Provider     string `json:"provider" validate:"required,max=50"`
	Amount       int    `json:"amount" validate:"required,min=0,max=10000000"`
	PaymentDT    int64  `json:"payment_dt" validate:"required,min=0"`
	Bank         string `json:"bank" validate:"required,max=50"`
	DeliveryCost int    `json:"delivery_cost" validate:"min=0,max=1000000"`
	GoodsTotal   int    `json:"goods_total" validate:"min=0,max=1000000"`
	CustomFee    int    `json:"custom_fee" validate:"min=0,max=1000000"`
}
