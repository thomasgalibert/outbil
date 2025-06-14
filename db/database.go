package db

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"outbil/models"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	conn *sql.DB
}

func New(dbPath string) (*Database, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &Database{conn: conn}
	if err := db.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return db, nil
}

func (db *Database) Close() error {
	return db.conn.Close()
}

func (db *Database) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS companies (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT,
			phone TEXT,
			address TEXT,
			city TEXT,
			postal_code TEXT,
			country TEXT,
			tax_id TEXT,
			logo BLOB,
			website TEXT,
			currency TEXT DEFAULT 'EUR',
			tax_rate REAL DEFAULT 20.0
		)`,
		`CREATE TABLE IF NOT EXISTS clients (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT,
			phone TEXT,
			address TEXT,
			city TEXT,
			postal_code TEXT,
			country TEXT,
			company TEXT,
			tax_id TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS quotes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			quote_number TEXT UNIQUE NOT NULL,
			client_id INTEGER NOT NULL,
			date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			valid_until TIMESTAMP,
			status TEXT DEFAULT 'draft',
			notes TEXT,
			terms TEXT,
			total_amount REAL DEFAULT 0,
			tax_amount REAL DEFAULT 0,
			discount REAL DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (client_id) REFERENCES clients(id)
		)`,
		`CREATE TABLE IF NOT EXISTS quote_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			quote_id INTEGER NOT NULL,
			description TEXT NOT NULL,
			quantity REAL DEFAULT 1,
			unit_price REAL NOT NULL,
			tax_rate REAL DEFAULT 0,
			amount REAL NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (quote_id) REFERENCES quotes(id) ON DELETE CASCADE
		)`,
	}

	for _, query := range queries {
		if _, err := db.conn.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

func (db *Database) CreateClient(client *models.Client) error {
	query := `INSERT INTO clients (name, email, phone, address, city, postal_code, country, company, tax_id) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	result, err := db.conn.Exec(query, client.Name, client.Email, client.Phone, client.Address, 
		client.City, client.PostalCode, client.Country, client.Company, client.TaxID)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	client.ID = int(id)
	client.CreatedAt = time.Now()
	client.UpdatedAt = time.Now()
	
	return nil
}

func (db *Database) GetClient(id int) (*models.Client, error) {
	query := `SELECT id, name, email, phone, address, city, postal_code, country, company, tax_id, created_at, updated_at 
			  FROM clients WHERE id = ?`
	
	client := &models.Client{}
	err := db.conn.QueryRow(query, id).Scan(
		&client.ID, &client.Name, &client.Email, &client.Phone, &client.Address,
		&client.City, &client.PostalCode, &client.Country, &client.Company, &client.TaxID,
		&client.CreatedAt, &client.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("client not found")
	}
	
	return client, err
}

func (db *Database) ListClients() ([]models.Client, error) {
	query := `SELECT id, name, email, phone, address, city, postal_code, country, company, tax_id, created_at, updated_at 
			  FROM clients ORDER BY name`
	
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []models.Client
	for rows.Next() {
		var client models.Client
		err := rows.Scan(
			&client.ID, &client.Name, &client.Email, &client.Phone, &client.Address,
			&client.City, &client.PostalCode, &client.Country, &client.Company, &client.TaxID,
			&client.CreatedAt, &client.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}

	return clients, nil
}

func (db *Database) UpdateClient(client *models.Client) error {
	query := `UPDATE clients SET name=?, email=?, phone=?, address=?, city=?, postal_code=?, 
			  country=?, company=?, tax_id=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`
	
	_, err := db.conn.Exec(query, client.Name, client.Email, client.Phone, client.Address,
		client.City, client.PostalCode, client.Country, client.Company, client.TaxID, client.ID)
	
	return err
}

func (db *Database) DeleteClient(id int) error {
	_, err := db.conn.Exec("DELETE FROM clients WHERE id = ?", id)
	return err
}

func (db *Database) CreateQuote(quote *models.Quote) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	quoteQuery := `INSERT INTO quotes (quote_number, client_id, date, valid_until, status, notes, terms, total_amount, tax_amount, discount) 
				   VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	result, err := tx.Exec(quoteQuery, quote.QuoteNumber, quote.ClientID, quote.Date, quote.ValidUntil,
		quote.Status, quote.Notes, quote.Terms, quote.TotalAmount, quote.TaxAmount, quote.Discount)
	if err != nil {
		return err
	}

	quoteID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	quote.ID = int(quoteID)

	for _, item := range quote.Items {
		itemQuery := `INSERT INTO quote_items (quote_id, description, quantity, unit_price, tax_rate, amount) 
					  VALUES (?, ?, ?, ?, ?, ?)`
		
		_, err := tx.Exec(itemQuery, quoteID, item.Description, item.Quantity, item.UnitPrice, item.TaxRate, item.Amount)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *Database) GetQuote(id int) (*models.Quote, error) {
	query := `SELECT q.id, q.quote_number, q.client_id, q.date, q.valid_until, q.status, q.notes, q.terms, 
			  q.total_amount, q.tax_amount, q.discount, q.created_at, q.updated_at,
			  c.id, c.name, c.email, c.phone, c.address, c.city, c.postal_code, c.country, c.company, c.tax_id
			  FROM quotes q
			  JOIN clients c ON q.client_id = c.id
			  WHERE q.id = ?`
	
	quote := &models.Quote{Client: &models.Client{}}
	err := db.conn.QueryRow(query, id).Scan(
		&quote.ID, &quote.QuoteNumber, &quote.ClientID, &quote.Date, &quote.ValidUntil,
		&quote.Status, &quote.Notes, &quote.Terms, &quote.TotalAmount, &quote.TaxAmount, &quote.Discount,
		&quote.CreatedAt, &quote.UpdatedAt,
		&quote.Client.ID, &quote.Client.Name, &quote.Client.Email, &quote.Client.Phone,
		&quote.Client.Address, &quote.Client.City, &quote.Client.PostalCode, &quote.Client.Country,
		&quote.Client.Company, &quote.Client.TaxID,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("quote not found")
	}
	if err != nil {
		return nil, err
	}

	itemsQuery := `SELECT id, quote_id, description, quantity, unit_price, tax_rate, amount, created_at 
				   FROM quote_items WHERE quote_id = ?`
	
	rows, err := db.conn.Query(itemsQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.QuoteItem
		err := rows.Scan(&item.ID, &item.QuoteID, &item.Description, &item.Quantity,
			&item.UnitPrice, &item.TaxRate, &item.Amount, &item.CreatedAt)
		if err != nil {
			return nil, err
		}
		quote.Items = append(quote.Items, item)
	}

	return quote, nil
}

func (db *Database) ListQuotes() ([]models.Quote, error) {
	query := `SELECT q.id, q.quote_number, q.client_id, q.date, q.valid_until, q.status, 
			  q.total_amount, q.created_at, c.name
			  FROM quotes q
			  JOIN clients c ON q.client_id = c.id
			  ORDER BY q.created_at DESC`
	
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quotes []models.Quote
	for rows.Next() {
		quote := models.Quote{Client: &models.Client{}}
		err := rows.Scan(
			&quote.ID, &quote.QuoteNumber, &quote.ClientID, &quote.Date, &quote.ValidUntil,
			&quote.Status, &quote.TotalAmount, &quote.CreatedAt, &quote.Client.Name,
		)
		if err != nil {
			return nil, err
		}
		quotes = append(quotes, quote)
	}

	return quotes, nil
}

func (db *Database) UpdateQuoteStatus(id int, status string) error {
	_, err := db.conn.Exec("UPDATE quotes SET status=?, updated_at=CURRENT_TIMESTAMP WHERE id=?", status, id)
	return err
}

func (db *Database) UpdateQuote(quote *models.Quote) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	quoteQuery := `UPDATE quotes SET client_id=?, valid_until=?, notes=?, terms=?, 
				   total_amount=?, tax_amount=?, discount=?, updated_at=CURRENT_TIMESTAMP 
				   WHERE id=?`
	
	_, err = tx.Exec(quoteQuery, quote.ClientID, quote.ValidUntil, quote.Notes, quote.Terms,
		quote.TotalAmount, quote.TaxAmount, quote.Discount, quote.ID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM quote_items WHERE quote_id = ?", quote.ID)
	if err != nil {
		return err
	}

	for _, item := range quote.Items {
		itemQuery := `INSERT INTO quote_items (quote_id, description, quantity, unit_price, tax_rate, amount) 
					  VALUES (?, ?, ?, ?, ?, ?)`
		
		_, err := tx.Exec(itemQuery, quote.ID, item.Description, item.Quantity, item.UnitPrice, item.TaxRate, item.Amount)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *Database) DeleteQuote(id int) error {
	_, err := db.conn.Exec("DELETE FROM quotes WHERE id = ?", id)
	return err
}

func (db *Database) DuplicateQuote(id int) (*models.Quote, error) {
	// Charger le devis source avec tous ses items
	sourceQuote, err := db.GetQuote(id)
	if err != nil {
		return nil, fmt.Errorf("impossible de charger le devis source: %w", err)
	}

	// Générer un nouveau numéro de devis
	newQuoteNumber, err := db.GetNextQuoteNumber()
	if err != nil {
		return nil, fmt.Errorf("impossible de générer un nouveau numéro de devis: %w", err)
	}

	// Créer le nouveau devis avec les données copiées
	newQuote := &models.Quote{
		QuoteNumber: newQuoteNumber,
		ClientID:    sourceQuote.ClientID,
		Date:        time.Now(),
		ValidUntil:  time.Now().AddDate(0, 1, 0), // Validité d'un mois par défaut
		Status:      "draft",
		Notes:       sourceQuote.Notes,
		Terms:       sourceQuote.Terms,
		TotalAmount: sourceQuote.TotalAmount,
		TaxAmount:   sourceQuote.TaxAmount,
		Discount:    sourceQuote.Discount,
		Items:       []models.QuoteItem{}, // On va copier les items après création
	}

	// Créer le nouveau devis dans la base
	err = db.CreateQuote(newQuote)
	if err != nil {
		return nil, fmt.Errorf("impossible de créer le devis dupliqué: %w", err)
	}

	// Copier tous les items du devis source
	for _, item := range sourceQuote.Items {
		newItem := models.QuoteItem{
			QuoteID:     newQuote.ID,
			Description: item.Description,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			TaxRate:     item.TaxRate,
			Amount:      item.Amount,
		}
		newQuote.Items = append(newQuote.Items, newItem)
	}

	// Mettre à jour le devis avec les items (utilise UpdateQuote qui gère la suppression/recréation des items)
	err = db.UpdateQuote(newQuote)
	if err != nil {
		// En cas d'erreur, supprimer le devis créé
		db.DeleteQuote(newQuote.ID)
		return nil, fmt.Errorf("impossible de copier les items: %w", err)
	}

	// Recharger le devis complet avec le client
	return db.GetQuote(newQuote.ID)
}

func (db *Database) GetNextQuoteNumber() (string, error) {
	// Générer un identifiant unique de 8 caractères
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const length = 8
	
	for attempts := 0; attempts < 100; attempts++ {
		// Générer une chaîne aléatoire avec crypto/rand pour une meilleure entropie
		b := make([]byte, length)
		randomBytes := make([]byte, length)
		_, err := rand.Read(randomBytes)
		if err != nil {
			return "", fmt.Errorf("erreur lors de la génération aléatoire: %w", err)
		}
		
		for i := range b {
			b[i] = charset[int(randomBytes[i])%len(charset)]
		}
		quoteNumber := string(b)
		
		// Vérifier l'unicité
		var exists int
		err = db.conn.QueryRow("SELECT COUNT(*) FROM quotes WHERE quote_number = ?", quoteNumber).Scan(&exists)
		if err != nil {
			return "", err
		}
		
		if exists == 0 {
			return quoteNumber, nil
		}
	}
	
	return "", fmt.Errorf("impossible de générer un numéro unique après 100 tentatives")
}

func (db *Database) GetCompany() (*models.Company, error) {
	query := `SELECT id, name, email, phone, address, city, postal_code, country, tax_id, logo, website, currency, tax_rate 
			  FROM companies LIMIT 1`
	
	company := &models.Company{}
	err := db.conn.QueryRow(query).Scan(
		&company.ID, &company.Name, &company.Email, &company.Phone, &company.Address,
		&company.City, &company.PostalCode, &company.Country, &company.TaxID,
		&company.Logo, &company.Website, &company.Currency, &company.TaxRate,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	
	return company, err
}

func (db *Database) SaveCompany(company *models.Company) error {
	existingCompany, err := db.GetCompany()
	if err != nil {
		return err
	}

	if existingCompany == nil {
		query := `INSERT INTO companies (name, email, phone, address, city, postal_code, country, tax_id, logo, website, currency, tax_rate) 
				  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		
		result, err := db.conn.Exec(query, company.Name, company.Email, company.Phone, company.Address,
			company.City, company.PostalCode, company.Country, company.TaxID,
			company.Logo, company.Website, company.Currency, company.TaxRate)
		if err != nil {
			return err
		}

		id, _ := result.LastInsertId()
		company.ID = int(id)
		return nil
	}

	query := `UPDATE companies SET name=?, email=?, phone=?, address=?, city=?, postal_code=?, 
			  country=?, tax_id=?, logo=?, website=?, currency=?, tax_rate=? WHERE id=?`
	
	_, err = db.conn.Exec(query, company.Name, company.Email, company.Phone, company.Address,
		company.City, company.PostalCode, company.Country, company.TaxID,
		company.Logo, company.Website, company.Currency, company.TaxRate, existingCompany.ID)
	
	return err
}