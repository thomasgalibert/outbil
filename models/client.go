package models

import (
	"time"
)

type Client struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	City      string    `json:"city"`
	PostalCode string   `json:"postal_code"`
	Country   string    `json:"country"`
	Company   string    `json:"company"`
	TaxID     string    `json:"tax_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}