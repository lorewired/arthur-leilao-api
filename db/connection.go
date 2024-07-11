package db

import (
	"database/sql"
	"fmt"

	"github.com/Nier704/arthur-leilao-server/config"
	_ "github.com/lib/pq"
)

func NewPostgreConnection() (*sql.DB, error) {
	cfg := config.NewDBConfig()

	dns := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Pass, cfg.Name)

	db, err := sql.Open("postgres", dns)
	if err != nil {
		return nil, err
	}

	if err = createTable(db); err != nil {
		return nil, err
	}

	return db, nil
}

func createTable(db *sql.DB) error {
	sql := `CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`
	if _, err := db.Exec(sql); err != nil {
		return err
	}

	sql = `
	CREATE TABLE IF NOT EXISTS accounts (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		username VARCHAR(30) NOT NULL,
		password VARCHAR(255) NOT NULL,
		UNIQUE(username)
	);`

	if _, err := db.Exec(sql); err != nil {
		return err
	}

	sql = `
	CREATE TABLE IF NOT EXISTS products (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		account_id UUID NOT NULL,
		title VARCHAR(255) NOT NULL,
		description VARCHAR(255) NOT NULL,
		price NUMERIC(7, 2) NOT NULL,
		image_url TEXT
	);`

	if _, err := db.Exec(sql); err != nil {
		return err
	}

	sql = `
	CREATE TABLE IF NOT EXISTS account_product (
		account_id UUID NOT NULL,
		product_id UUID NOT NULL,
		PRIMARY KEY (account_id, product_id),
		FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
		FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
	);`

	if _, err := db.Exec(sql); err != nil {
		return err
	}

	sql = `
	CREATE TABLE IF NOT EXISTS account_bid (
		account_id UUID NOT NULL,
		product_id UUID NOT NULL,
		bid_value NUMERIC(5, 2) NOT NULL,
		bid_message TEXT NOT NULL,
		FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
		FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
	);`

	if _, err := db.Exec(sql); err != nil {
		return err
	}

	return nil
}
