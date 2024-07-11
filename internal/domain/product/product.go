package product

import "github.com/google/uuid"

type Product struct {
	ID          uuid.UUID `json:"id"`
	AccountID   uuid.UUID `json:"account_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	ImageURL    string    `json:"image_url"`
}
