package models

import (
	"time"
)

type Quote struct {
	ID           int         `json:"id"`
	QuoteNumber  string      `json:"quote_number"`
	ClientID     int         `json:"client_id"`
	Client       *Client     `json:"client,omitempty"`
	Date         time.Time   `json:"date"`
	ValidUntil   time.Time   `json:"valid_until"`
	Status       string      `json:"status"`
	Notes        string      `json:"notes"`
	Terms        string      `json:"terms"`
	TotalAmount  float64     `json:"total_amount"`
	TaxAmount    float64     `json:"tax_amount"`
	Discount     float64     `json:"discount"`
	Items        []QuoteItem `json:"items,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

type QuoteItem struct {
	ID          int     `json:"id"`
	QuoteID     int     `json:"quote_id"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TaxRate     float64 `json:"tax_rate"`
	Amount      float64 `json:"amount"`
	CreatedAt   time.Time `json:"created_at"`
}

const (
	StatusDraft    = "draft"
	StatusSent     = "sent"
	StatusAccepted = "accepted"
	StatusRejected = "rejected"
	StatusExpired  = "expired"
)