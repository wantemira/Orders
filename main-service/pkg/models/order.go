// Package models содержит структуры данных заказов
package models

import "time"

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
