package config

import (
	"os"

	"github.com/joho/godotenv"
)

type DBConfig struct {
	Host string
	Port string
	User string
	Pass string
	Name string
}

func NewDBConfig() *DBConfig {
	godotenv.Load()

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	name := os.Getenv("DB_NAME")

	return &DBConfig{
		Host: host,
		Port: port,
		User: user,
		Pass: pass,
		Name: name,
	}
}
