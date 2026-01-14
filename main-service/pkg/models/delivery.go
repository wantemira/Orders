// Package models содержит структуры данных заказов
package models

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
