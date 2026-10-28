package model

import "time"

const LowStockLimit = 10

type Product struct {
	ID int `json:"id"`
	SKU string `json:"sku"`
	Name string `json:"name"`
	Price int `json:"price"`
	Stock int `json:"stock"`
	IsActive bool `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	LastMovementAt *time.Time `json:"last_movement_at"`
}

type ProductList struct {
	Data []Product `json:"data"`
	Page int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

type ProductInput struct {
	SKU string `json:"sku"`
	Name string `json:"name"`
	Price *int `json:"price"`
	IsActive *bool `json:"is_active"`
}

type Movement struct {
	ID int `json:"id"`
	ProductID int `json:"product_id"`
	Type string `json:"type"`
	Qty int `json:"qty"`
	Note *string `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

type StockInput struct {
	Type string `json:"type"`
	Qty int `json:"qty"`
	Note string `json:"note"`
}

type Order struct {
	ID int `json:"id"`
	Customer string `json:"customer"`
	Status string `json:"status"`
	Total int `json:"total"`
	CreatedAt time.Time `json:"created_at"`
	CancelledAt *time.Time `json:"cancelled_at"`
	Items []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
	SKU string `json:"sku"`
	Name string `json:"name"`
	Qty int `json:"qty"`
	UnitPrice int `json:"unit_price"`
}

type OrderInput struct {
	Customer string `json:"customer"`
	Items []OrderItemInput `json:"items"`
}

type OrderItemInput struct {
	SKU string `json:"sku"`
	Qty int `json:"qty"`
}

type Summary struct {
	TotalProducts int `json:"total_products"`
	ActiveProducts int `json:"active_products"`
	TotalStockValue int64 `json:"total_stock_value"`
	LowStockCount int `json:"low_stock_count"`
}
