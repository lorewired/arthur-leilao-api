package product

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type ProductHandler struct {
	DB *sql.DB
}

func NewProductHandler(db *sql.DB) *ProductHandler {
	return &ProductHandler{
		DB: db,
	}
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var body Product
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("Error getting product body: %v", err)
		http.Error(w, "error getting product body", http.StatusInternalServerError)
		return
	}

	if ok := validateCredentials(&body); ok {

		sql := `
		INSERT INTO products
		(title, account_id ,description, price, image_url)
		VALUES
		($1, $2, $3, $4, $5)
		RETURNING id;
		`
		ctx := context.Background()

		stmt, err := h.DB.PrepareContext(ctx, sql)
		if err != nil {
			log.Printf("Error preparing context: %v", err)
			http.Error(w, "error preparing context", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		var inserted_id string
		err = stmt.QueryRowContext(ctx, body.Title, body.AccountID, body.Description, body.Price, body.ImageURL).
			Scan(&inserted_id)
		if err != nil {
			log.Printf("Error creating product: %v", err)
			http.Error(w, "error creating product", http.StatusInternalServerError)
			return
		}

		res := map[string]string{
			"status":      "created",
			"inserted_id": inserted_id,
		}

		w.WriteHeader(201)

		if err := json.NewEncoder(w).Encode(res); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "error encoding response", 500)
		}

		return
	}

	fmt.Println("Invalid Body")
	http.Error(w, "invalid body", http.StatusBadRequest)
}

func (h *ProductHandler) GetById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.PathValue("productId")
	if id == "" {
		log.Printf("Invalid id")
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	sql := `
		SELECT id, account_id, title, description, price, image_url FROM products WHERE id = $1;
	`
	ctx := context.Background()

	stmt, err := h.DB.PrepareContext(ctx, sql)
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		http.Error(w, "error preparing statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	var product Product

	if err = stmt.QueryRowContext(ctx, id).
		Scan(&product.ID, &product.AccountID, &product.Title, &product.Description, &product.Price, &product.ImageURL); err != nil {
		log.Printf("not found: %v", err)
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	_, err = uuid.Parse(product.ID.String())
	if err != nil {
		log.Printf("Invalid accountId format: %v", err)
		http.Error(w, "invalid accountId format", 400)
		return
	}

	w.WriteHeader(200)

	if err := json.NewEncoder(w).Encode(product); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "error encoding response", 500)
	}
}

func (h *ProductHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sql := `
		SELECT id, account_id, title, description, price, image_url FROM products;
	`
	ctx := context.Background()

	stmt, err := h.DB.PrepareContext(ctx, sql)
	if err != nil {
		log.Printf("Error preparing context: %v", err)
		http.Error(w, "error preparing context", http.StatusInternalServerError)
		return
	}

	products := make([]Product, 0)
	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		log.Printf("Error getting all products: %v", err)
		http.Error(w, "error getting products", http.StatusInternalServerError)
		return
	}

	for rows.Next() {
		var product Product
		if err = rows.Scan(&product.ID, &product.AccountID, &product.Title, &product.Description, &product.Price, &product.ImageURL); err != nil {
			log.Printf("Error scanning account: %v", err)
			http.Error(w, "error scanning account", http.StatusInternalServerError)
			break
		}
		products = append(products, product)
	}

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(products); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "error encoding response", http.StatusInternalServerError)
	}
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.PathValue("productId")

	if id == "" {
		log.Println("Invalid id")
		http.Error(w, "invalid id", 400)
		return
	}

	var body Product
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("Error decoding body: %v", err)
		http.Error(w, "invalid body", 500)
		return
	}

	if ok := validateCredentials(&body); ok {

		sql := `
		UPDATE products
		SET title = $1,
			description = $2,
			price = $3,
			image_url = $4
		WHERE id = $5
		RETURNING id, title, description, price, image_url;
		`
		ctx := context.Background()

		stmt, err := h.DB.PrepareContext(ctx, sql)
		if err != nil {
			log.Printf("Error preparing statement: %v", err)
			http.Error(w, "error preparing statement", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		var product Product

		if err = stmt.QueryRowContext(ctx, body.Title, body.Description, body.Price, body.ImageURL, id).
			Scan(&product.ID, &product.Title, &product.Description, &product.Price, &product.ImageURL); err != nil {
			log.Printf("not found: %v", err)
			http.Error(w, "not found", 404)
			return
		}

		w.WriteHeader(200)

		if err := json.NewEncoder(w).Encode(product); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "error encoding response", 500)
			return
		}

		return
	}

	log.Printf("Invalid credentials")
	http.Error(w, "invalid credentials", http.StatusBadRequest)
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.PathValue("productId")

	if id == "" {
		log.Println("Invalid id")
		http.Error(w, "invalid id", 400)
		return
	}

	sql := `DELETE FROM products WHERE id = $1`

	ctx := context.Background()

	stmt, err := h.DB.PrepareContext(ctx, sql)
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		http.Error(w, "error preparing statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, id)
	if err != nil {
		log.Printf("Error deleting product: %v", err)
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		http.Error(w, "error getting rows affected", http.StatusInternalServerError)
		return
	}

	if rows > 1 {
		log.Printf("Error more than one line was affected: %v", err)
		http.Error(w, "error: more than one line was affected", http.StatusInternalServerError)
		return
	}

	res := map[string]string{
		"message": "product deleted",
	}

	w.WriteHeader(200)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "error encoding response", 500)
	}
}

// AssociateProductWithAccount associates a product with an account
func (h *ProductHandler) AssociateProductWithAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accountID := r.PathValue("accountId")
	productID := r.PathValue("productId")

	if accountID == "" || productID == "" {
		log.Println("Invalid accountId or productId")
		http.Error(w, "invalid accountId or productId", 400)
		return
	}

	_, err := uuid.Parse(accountID)
	if err != nil {
		log.Printf("Invalid accountId format: %v", err)
		http.Error(w, "invalid accountId format", 400)
		return
	}

	_, err = uuid.Parse(productID)
	if err != nil {
		log.Printf("Invalid accountId format: %v", err)
		http.Error(w, "invalid accountId format", 400)
		return
	}

	sql := `
		INSERT INTO account_product (account_id, product_id)
		VALUES ($1, $2);
	`
	ctx := context.Background()

	stmt, err := h.DB.PrepareContext(ctx, sql)
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		http.Error(w, "error preparing statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, accountID, productID); err != nil {
		log.Printf("Error associating product with account: %v", err)
		http.Error(w, "error associating product with account", http.StatusInternalServerError)
		return
	}

	res := map[string]string{
		"message": "product associated with account",
	}

	w.WriteHeader(201)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "error encoding response", 500)
	}
}

func (h *ProductHandler) GetAllAccountBids(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accountID := r.PathValue("accountId")

	if accountID == "" {
		log.Println("Invalid accountId")
		http.Error(w, "invalid accountId", 400)
		return
	}

	_, err := uuid.Parse(accountID)
	if err != nil {
		log.Printf("Invalid accountId format: %v", err)
		http.Error(w, "invalid accountId format", 400)
		return
	}

	sql := `
		SELECT product_id FROM account_product
		WHERE account_id = $1;
	`
	ctx := context.Background()

	stmt, err := h.DB.PrepareContext(ctx, sql)
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		http.Error(w, "error preparing statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, accountID)
	if err != nil {
		log.Printf("Error getting account product: %v", err)
		http.Error(w, "error getting product account product", http.StatusInternalServerError)
		return
	}

	var ids []string
	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			break
		}
		ids = append(ids, id)
	}

	var products []Product
	for _, id := range ids {
		sql = `
		SELECT id, account_id, title, description, price, image_url FROM products
		WHERE id = $1;
		`

		stmt, err = h.DB.PrepareContext(ctx, sql)
		if err != nil {
			log.Printf("Error preparing statement: %v", err)
			http.Error(w, "error preparing statement", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		var product Product
		if err = stmt.QueryRowContext(ctx, id).Scan(&product.ID, &product.AccountID, &product.Title, &product.Description, &product.Price, &product.ImageURL); err != nil {
			log.Printf("Error getting account product: %v", err)
			http.Error(w, "error getting product account product", http.StatusInternalServerError)
			return
		}

		products = append(products, product)
	}

	w.WriteHeader(200)

	if err := json.NewEncoder(w).Encode(products); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "error encoding response", 500)
	}
}

func (h *ProductHandler) AddBid(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accountID := r.PathValue("accountId")
	productID := r.PathValue("productId")

	if accountID == "" || productID == "" {
		log.Println("Invalid accountId or productId")
		http.Error(w, "invalid accountId or productId", 400)
		return
	}

	var body struct {
		BidValue   float64 `json:"bid_value"`
		BidMessage string  `json:"bid_message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Println("Error decoding body")
		http.Error(w, "error decoding body", 500)
		return
	}

	if body.BidValue <= 0 || body.BidMessage == "" {
		log.Println("Invalid Bid Value or Bid Message")
		http.Error(w, "invalid Bid Value or Bid Message", 400)
		return
	}

	sql := `
	INSERT INTO account_bid
	(account_id, product_id, bid_value, bid_message)
	VALUES ($1, $2, $3, $4);
	`
	ctx := context.Background()

	stmt, err := h.DB.PrepareContext(ctx, sql)
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		http.Error(w, "error preparing statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, accountID, productID, body.BidValue, body.BidMessage)
	if err != nil {
		log.Printf("Error creating a product bid: %v", err)
		http.Error(w, "error creating a product bid", http.StatusInternalServerError)
		return
	}

	res := map[string]string{
		"status": "bid created",
	}

	w.WriteHeader(http.StatusCreated)

	if err = json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func (h *ProductHandler) GetBidById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accountID := r.PathValue("accountId")
	productID := r.PathValue("productId")

	if accountID == "" || productID == "" {
		log.Println("Invalid accountId or productId")
		http.Error(w, "invalid accountId or productId", 400)
		return
	}

	sql := `
	SELECT * FROM account_bid
	WHERE account_id = $1 AND product_id = $2;
	`
	ctx := context.Background()

	stmt, err := h.DB.PrepareContext(ctx, sql)
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		http.Error(w, "error preparing statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	var productBid struct {
		AccountID  uuid.UUID `json:"account_id"`
		ProductID  uuid.UUID `json:"product_id"`
		BidValue   float64   `json:"bid_value"`
		BidMessage string    `json:"bid_message"`
	}

	if err = stmt.QueryRowContext(ctx, accountID, productID).
		Scan(&productBid.AccountID, &productBid.ProductID, &productBid.BidValue, &productBid.BidMessage); err != nil {
		log.Printf("Error getting product bid: %v", err)
		http.Error(w, "error getting product bid", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(productBid); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func (h *ProductHandler) GetAllBids(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accountId := r.PathValue("accountId")

	if accountId == "" {
		log.Println("Invalid account id")
		http.Error(w, "invalid id", 400)
		return
	}

	sql := `
	SELECT * FROM account_bid
	WHERE account_id = $1;
	`
	ctx := context.Background()

	stmt, err := h.DB.PrepareContext(ctx, sql)
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		http.Error(w, "error preparing statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	type bid struct {
		AccountID  uuid.UUID `json:"account_id"`
		ProductID  uuid.UUID `json:"product_id"`
		BidValue   float64   `json:"bid_value"`
		BidMessage string    `json:"bid_message"`
	}

	rows, err := stmt.QueryContext(ctx, accountId)
	if err != nil {
		log.Printf("Error getting account bids: %v", err)
		http.Error(w, "error getting account bids", http.StatusInternalServerError)
		return
	}

	bids := make([]bid, 0)

	for rows.Next() {
		var b bid
		err = rows.Scan(&b.AccountID, &b.ProductID, &b.BidValue, &b.BidMessage)
		if err != nil {
			break
		}
		bids = append(bids, b)
	}

	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(bids); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func validateCredentials(body *Product) bool {
	return body.Title != "" && body.Description != "" && body.Price > 0
}
