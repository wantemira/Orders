// Package models содержит структуры данных заказов
package models

import (
	"time"
)

// OrderJSON представляет заказ в формате JSON для API
type OrderJSON struct {
	OrderUID          string    `json:"order_uid" validate:"required,min=10,max=50"`
	TrackNumber       string    `json:"track_number" validate:"required,alphanumunicode,max=50"`
	Entry             string    `json:"entry" validate:"required,max=10"`
	Locale            string    `json:"locale" validate:"required,len=2"`
	InternalSignature string    `json:"internal_signature" validate:"max=255"`
	CustomerID        string    `json:"customer_id" validate:"max=50"`
	DeliveryService   string    `json:"delivery_service" validate:"max=50"`
	ShardKey          string    `json:"shardkey" validate:"max=10"`
	SmID              int       `json:"sm_id" validate:"min=0,max=999"`
	DateCreated       time.Time `json:"date_created" validate:"required"`
	OofShard          string    `json:"oof_shard" validate:"max=10"`

	Delivery Delivery `json:"delivery" validate:"required"`
	Payment  Payment  `json:"payment" validate:"required"`
	Items    []Item   `json:"items" validate:"required,min=1,max=100,dive"`
}

// Order представляет заказ в базе данных
type Order struct {
	OrderUID          string    `json:"order_uid" validate:"required,min=10,max=50"`
	TrackNumber       string    `json:"track_number" validate:"required,alphanumunicode,max=50"`
	Entry             string    `json:"entry" validate:"required,max=10"`
	Locale            string    `json:"locale" validate:"required,len=2"`
	InternalSignature string    `json:"internal_signature" validate:"max=255"`
	CustomerID        string    `json:"customer_id" validate:"max=50"`
	DeliveryService   string    `json:"delivery_service" validate:"max=50"`
	ShardKey          string    `json:"shardkey" validate:"max=10"`
	SmID              int       `json:"sm_id" validate:"min=0,max=999"`
	DateCreated       time.Time `json:"date_created" validate:"required"`
	OofShard          string    `json:"oof_shard" validate:"max=10"`
}

// Delivery содержит информацию о доставке
type Delivery struct {
	OrderUID string `json:"order_uid" validate:"required,min=10,max=50"`
	Name     string `json:"name" validate:"required,max=100"`
	Phone    string `json:"phone" validate:"required,max=20"`
	Zip      string `json:"zip" validate:"required,max=20"`
	City     string `json:"city" validate:"required,max=50"`
	Address  string `json:"address" validate:"required,max=100"`
	Region   string `json:"region" validate:"required,max=50"`
	Email    string `json:"email" validate:"required,email,max=100"`
}

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

// Item представляет товар в заказе
type Item struct {
	ID          uint   `json:"id" validate:"-"` // - означает "не валидировать"
	ChrtID      int64  `json:"chrt_id" validate:"required,min=1"`
	TrackNumber string `json:"track_number" validate:"required,alphanumunicode,max=50"`
	Price       int    `json:"price" validate:"required,min=0,max=1000000"`
	RID         string `json:"rid" validate:"required,max=50"`
	Name        string `json:"name" validate:"required,max=100"`
	Sale        int    `json:"sale" validate:"min=0,max=100"`
	Size        string `json:"size" validate:"max=10"`
	TotalPrice  int    `json:"total_price" validate:"required,min=0,max=1000000"`
	NmID        int64  `json:"nm_id" validate:"required,min=1"`
	Brand       string `json:"brand" validate:"max=100"`
	Status      int    `json:"status" validate:"min=0,max=999"`
}
