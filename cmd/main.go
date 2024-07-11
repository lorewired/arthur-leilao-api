package main

import (
	"log"

	"github.com/Nier704/arthur-leilao-server/db"
	"github.com/Nier704/arthur-leilao-server/internal/domain"
)

func main() {
	db, err := db.NewPostgreConnection()
	if err != nil {
		log.Fatal(err)
	}

	router := domain.NewRouter()
	router.Init(db)
	router.Start()
}
