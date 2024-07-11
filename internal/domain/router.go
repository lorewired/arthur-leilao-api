package domain

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/Nier704/arthur-leilao-server/internal/domain/account"
	"github.com/Nier704/arthur-leilao-server/internal/domain/jwt"
	"github.com/Nier704/arthur-leilao-server/internal/domain/product"
	"github.com/Nier704/arthur-leilao-server/internal/middlewares"
	"github.com/gorilla/handlers"
)

type Router struct {
	accountHandler *account.AccountHandler
	jwt            *jwt.Jwt
	productHandler *product.ProductHandler
	mux            *http.ServeMux
	port           string
}

func NewRouter() *Router {
	return &Router{
		accountHandler: nil,
		productHandler: nil,
		jwt:            nil,
		mux:            http.NewServeMux(),
		port:           "3000",
	}
}

func (r *Router) Init(db *sql.DB) {
	ah := account.NewAccountHandler(db)
	ph := product.NewProductHandler(db)
	jwt := jwt.NewJwt(db)

	r.accountHandler = ah
	r.productHandler = ph
	r.jwt = jwt

	r.setAccountsRoutes()
	r.setProductsRoutes()
}

func (r *Router) Start() {
	fmt.Println("server is already running at PORT: " + r.port)

	if err := http.ListenAndServe("0.0.0.0:"+r.port, handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:5173"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
		handlers.AllowCredentials(),
	)(r.mux)); err != nil {
		log.Fatal(err)
	}
}

func (r *Router) setAccountsRoutes() {
	r.mux.Handle("GET /api/accounts", middlewares.Log(http.HandlerFunc(r.accountHandler.GetAll)))
	r.mux.Handle("GET /api/account/{accountId}", middlewares.Log(http.HandlerFunc(r.accountHandler.GetById)))
	r.mux.Handle("GET /api/account/auth", middlewares.Log(http.HandlerFunc(r.jwt.Authenticate)))
	r.mux.Handle("POST /api/account/signup", middlewares.Log(http.HandlerFunc(r.jwt.Signup)))
	r.mux.Handle("POST /api/account/login", middlewares.Log(http.HandlerFunc(r.jwt.Login)))
	r.mux.Handle("POST /api/account/logout", middlewares.Log(http.HandlerFunc(r.jwt.Logout)))
	r.mux.Handle("PUT /api/account/{accountId}", middlewares.Log(http.HandlerFunc(r.accountHandler.Update)))
	r.mux.Handle("DELETE /api/account/{accountId}", middlewares.Log(http.HandlerFunc(r.accountHandler.Delete)))
}

func (r *Router) setProductsRoutes() {
	r.mux.Handle("GET /api/products", middlewares.Log(http.HandlerFunc(r.productHandler.GetAll)))
	r.mux.Handle("GET /api/product/{productId}", middlewares.Log(http.HandlerFunc(r.productHandler.GetById)))
	r.mux.Handle("POST /api/product", middlewares.Log(http.HandlerFunc(r.productHandler.Create)))
	r.mux.Handle("PUT /api/product/{productId}", middlewares.Log(http.HandlerFunc(r.productHandler.Update)))
	r.mux.Handle("DELETE /api/product/{productId}", middlewares.Log(http.HandlerFunc(r.productHandler.Delete)))

	r.mux.Handle("POST /api/account/{accountId}/product/{productId}", middlewares.Log(http.HandlerFunc(r.productHandler.AssociateProductWithAccount)))
	r.mux.Handle("GET /api/bid/account/{accountId}", middlewares.Log(http.HandlerFunc(r.productHandler.GetAllBids)))
	r.mux.Handle("GET /api/account/{accountId}/bids", middlewares.Log(http.HandlerFunc(r.productHandler.GetAllAccountBids)))
	r.mux.Handle("POST /api/bid/account/{accountId}/product/{productId}", middlewares.Log(http.HandlerFunc(r.productHandler.AddBid)))
	r.mux.Handle("GET /api/bid/account/{accountId}/product/{productId}", middlewares.Log(http.HandlerFunc(r.productHandler.GetBidById)))
}
