package models

type Company struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Address    string `json:"address"`
	City       string `json:"city"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
	TaxID      string `json:"tax_id"`
	Logo       []byte `json:"logo,omitempty"`
	Website    string `json:"website"`
	Currency   string `json:"currency"`
	TaxRate    float64 `json:"tax_rate"`
}