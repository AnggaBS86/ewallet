package main

import (
	"log"
	"net/http"

	"ewallet/internal/config"
	"ewallet/internal/db"
	"ewallet/internal/router"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Cannot load .env file, please check your configuration")
	}

	cfg := config.Load()
	database, err := db.New(cfg)
	if err != nil {
		log.Fatalf("db connection failed: %v", err)
	}

	e := echo.New()
	router.Register(e, database, cfg)

	addr := ":" + cfg.Port
	if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
