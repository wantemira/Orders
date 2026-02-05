// Package models содержит структуры данных заказов
package models

// Item представляет товар в заказе
type Item struct {
	ID          uint   `json:"id" validate:"-"` // - означает "не валидировать"
	ChrtID      int64  `json:"chrt_id" validate:"required,min=1,max=2147483647"`
	TrackNumber string `json:"track_number" validate:"required,alphanumunicode,max=50"`
	Price       int    `json:"price" validate:"required,min=0,max=1000000"`
	RID         string `json:"rid" validate:"required,max=50"`
	Name        string `json:"name" validate:"required,max=100"`
	Sale        int    `json:"sale" validate:"min=0,max=100"`
	Size        string `json:"size" validate:"max=10"`
	TotalPrice  int    `json:"total_price" validate:"required,min=0,max=1000000"`
	NmID        int64  `json:"nm_id" validate:"required,min=1,max=2147483647"`
	Brand       string `json:"brand" validate:"max=100"`
	Status      int    `json:"status" validate:"min=0,max=999"`
}
