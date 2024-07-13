package jwt

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Nier704/arthur-leilao-server/internal/domain/account"
	"github.com/Nier704/arthur-leilao-server/internal/utils"
	"github.com/golang-jwt/jwt/v5"
)

var secret_key = []byte(os.Getenv("JWT_SECRET"))

type Jwt struct {
	DB *sql.DB
}

func NewJwt(db *sql.DB) *Jwt {
	return &Jwt{
		DB: db,
	}
}

func (jwt *Jwt) Authenticate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cookie, err := r.Cookie("jwt")
	if err != nil {
		log.Printf("Error getting jwt cookie")
		http.Error(w, "error getting jwt cookie", 500)
		return
	}

	tokenString := cookie.Value

	token, err := verifyToken(tokenString)
	if err != nil {
		log.Printf("token verification failed")
		http.Error(w, "token verification failed", 500)
		return
	}

	id, err := token.Claims.GetSubject()
	if err != nil {
		log.Printf("error getting token subject")
		http.Error(w, "error getting token subject", 500)
		return
	}

	sql := `
		SELECT * FROM accounts WHERE id = $1;
	`
	ctx := context.Background()

	stmt, err := jwt.DB.PrepareContext(ctx, sql)
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		http.Error(w, "error preparing statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	var acc account.Account

	if err = stmt.QueryRowContext(ctx, id).Scan(&acc.ID, &acc.Username, &acc.Password); err != nil {
		log.Printf("Error getting account: %v", err)
		http.Error(w, "error getting account", http.StatusNotFound)
		return
	}

	w.WriteHeader(200)

	if err = json.NewEncoder(w).Encode(acc); err != nil {
		log.Printf("Error encoding response")
		http.Error(w, "error encoding response", 500)
	}
}

func generateToken(id string, secret_key []byte) (string, error) {
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": id,
		"iss": "arthurleilao",
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	token_string, err := claims.SignedString(secret_key)
	if err != nil {
		return "", err
	}

	return token_string, nil
}

func verifyToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secret_key, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	exp_time, err := token.Claims.GetExpirationTime()
	if err != nil {
		return nil, err
	}

	if time.Now().Unix() > exp_time.Unix() {
		return nil, errors.New("token expired")
	}

	return token, nil
}

func (jwt *Jwt) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var body account.Account
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("Error decoding body: %v", err)
		http.Error(w, "invalid body", http.StatusInternalServerError)
		return
	}

	if ok := account.ValidateCredentials(&body); !ok {
		log.Printf("invalid credentials")
		http.Error(w, "invalid credentials", http.StatusBadRequest)
		return
	}

	acc, err := tryLogin(&body, jwt.DB)
	if err != nil {
		log.Printf("not found: %v", err)
		http.Error(w, "not found", 404)
		return
	}

	token, err := generateToken(acc.ID.String(), secret_key)
	if err != nil {
		log.Printf("error generating jwt token: %v", err)
		http.Error(w, "error generating jwt token", 500)
		return
	}

	// create and set cookie
	cookie := http.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(time.Hour * 24),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}

	http.SetCookie(w, &cookie)

	res := map[string]string{
		"message": "success",
	}

	w.WriteHeader(200)

	if err = json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "error encoding response", 500)
	}
}

func tryLogin(body *account.Account, db *sql.DB) (*account.Account, error) {
	sql := `
			SELECT * FROM accounts
			WHERE username = $1; 
		`
	ctx := context.Background()

	stmt, err := db.PrepareContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var acc account.Account

	err = stmt.QueryRowContext(ctx, body.Username).
		Scan(&acc.ID, &acc.Username, &acc.Password)
	if err != nil {
		return nil, err
	}

	if err := utils.CompareHashAndPassword(acc.Password, body.Password); err != nil {
		return nil, err
	}

	return &acc, nil
}

func (jwt *Jwt) Logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cookie := http.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-(time.Hour * 24)),
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)

	res := map[string]string{
		"status":  "disconnected",
		"message": "success",
	}

	w.WriteHeader(200)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Error encoding message", 500)
	}
}

func (jwt *Jwt) Signup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var body account.Account
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("Error decoding body: %v", err)
		http.Error(w, "invalid body", http.StatusInternalServerError)
		return
	}

	if ok := account.ValidateCredentials(&body); !ok {
		log.Printf("invalid credentials")
		http.Error(w, "invalid credentials", http.StatusBadRequest)
		return
	}

	err := jwt.account_exists(&body)
	if err == nil {
		log.Printf("Error searching existing account: %v", err)
		http.Error(w, "Account already exists", http.StatusInternalServerError)
		return
	}

	hash, err := utils.HashPassword(body.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		http.Error(w, "error hashing password", http.StatusInternalServerError)
		return
	}

	body.Password = string(hash)

	inserted_id, err := jwt.insert_account(&body)
	if err != nil {
		log.Println(err)
		http.Error(w, "error creating new account", http.StatusInternalServerError)
		return
	}

	res := map[string]string{
		"status":      "created",
		"inserted_id": inserted_id,
	}

	w.WriteHeader(201)

	if err = json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "error encoding response", http.StatusInternalServerError)
	}
}

func (jwt *Jwt) insert_account(body *account.Account) (string, error) {
	sql := `
		INSERT INTO accounts
		(username, password)
		VALUES ($1, $2)
		RETURNING id;
	`
	ctx := context.Background()

	stmt, err := jwt.DB.PrepareContext(ctx, sql)
	if err != nil {
		return "", errors.New("error preparing context")
	}
	defer stmt.Close()

	var inserted_id string
	err = stmt.QueryRowContext(ctx, body.Username, body.Password).Scan(&inserted_id)
	if err != nil {
		return "", err
	}

	return inserted_id, nil
}

func (jwt *Jwt) account_exists(body *account.Account) error {
	sql := `
	SELECT username FROM accounts
	WHERE username = $1;
	`
	ctx := context.Background()

	stmt, err := jwt.DB.PrepareContext(ctx, sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	var existing_username string
	err = stmt.QueryRowContext(ctx, body.Username).Scan(&existing_username)
	if err != nil {
		return err
	}

	return err
}
