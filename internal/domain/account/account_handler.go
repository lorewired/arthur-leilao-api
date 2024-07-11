package account

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Nier704/arthur-leilao-server/internal/utils"
)

type AccountHandler struct {
	DB *sql.DB
}

func NewAccountHandler(db *sql.DB) *AccountHandler {
	return &AccountHandler{
		DB: db,
	}
}

func (h *AccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.PathValue("accountId")

	if id == "" {
		log.Println("Invalid id")
		http.Error(w, "invalid id", 400)
		return
	}

	sql := `DELETE FROM accounts WHERE id = $1`

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
		log.Printf("Error deleting account: %v", err)
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
		"message": "account deleted",
	}

	w.WriteHeader(200)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "error encoding response", 500)
	}
}

func (h *AccountHandler) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.PathValue("accountId")

	if id == "" {
		log.Println("Invalid id")
		http.Error(w, "invalid id", 400)
		return
	}

	var body Account
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("Error decoding body: %v", err)
		http.Error(w, "Error decoding body", 500)
		return
	}

	if ok := ValidateCredentials(&body); ok {

		hash, err := utils.HashPassword(body.Password)
		if err != nil {
			log.Printf("Error generating encrypted password: %v", err)
			http.Error(w, "Error generating encrypted password", 500)
			return
		}

		body.Password = string(hash)

		sql := `
		UPDATE users
		SET username = $1,
				password = $2
		WHERE id = $3
		RETURNING id, username, password;
	`
		ctx := context.Background()

		stmt, err := h.DB.PrepareContext(ctx, sql)
		if err != nil {
			log.Printf("Error preparing statement: %v", err)
			http.Error(w, "error preparing statement", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		var acc Account

		if err = stmt.QueryRowContext(ctx, body.Username, body.Password).Scan(&acc.ID, &acc.Username, &acc.Password); err != nil {
			log.Printf("not found: %v", err)
			http.Error(w, "not found", 404)
			return
		}

		w.WriteHeader(200)

		if err := json.NewEncoder(w).Encode(acc); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "error encoding response", 500)
			return
		}

		return
	}

	log.Printf("Invalid credentials")
	http.Error(w, "invalid credentials", http.StatusBadRequest)
}

func (h *AccountHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sql := `
		SELECT * FROM accounts;
	`
	ctx := context.Background()

	stmt, err := h.DB.PrepareContext(ctx, sql)
	if err != nil {
		log.Printf("Error preparing context: %v", err)
		http.Error(w, "error preparing context", http.StatusInternalServerError)
		return
	}

	accounts := make([]Account, 0)
	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		log.Printf("Error getting all accounts: %v", err)
		http.Error(w, "error getting accounts", http.StatusInternalServerError)
		return
	}

	for rows.Next() {
		var acc Account
		if err = rows.Scan(&acc.ID, &acc.Username, &acc.Password); err != nil {
			log.Printf("Error scanning account: %v", err)
			http.Error(w, "error scanning account", http.StatusInternalServerError)
			break
		}
		accounts = append(accounts, acc)
	}

	w.WriteHeader(200)

	if err := json.NewEncoder(w).Encode(accounts); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "error encoding response", http.StatusInternalServerError)
	}
}

func (h *AccountHandler) GetById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.PathValue("accountId")
	if id == "" {
		log.Printf("Invalid id")
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	sql := `
		SELECT * FROM accounts WHERE id = $1;
	`
	ctx := context.Background()

	stmt, err := h.DB.PrepareContext(ctx, sql)
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		http.Error(w, "error preparing statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	var acc Account

	if err = stmt.QueryRowContext(ctx, id).Scan(&acc.ID, &acc.Username, &acc.Password); err != nil {
		log.Printf("not found: %v", err)
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(200)

	if err := json.NewEncoder(w).Encode(acc); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "error encoding response", 500)
	}
}

func ValidateCredentials(acc *Account) bool {
	return acc.Username != "" && acc.Password != ""
}
