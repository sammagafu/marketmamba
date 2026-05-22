package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"forex-bot/internal/adminseed"
	"forex-bot/internal/config"
	"forex-bot/internal/storage"
)

func main() {
	_ = godotenv.Load()
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("config: %v\n", err)
		os.Exit(1)
	}
	store, err := storage.NewPostgresStorage(cfg.Database.URL)
	if err != nil {
		fmt.Printf("database: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()
	if err := adminseed.Run(store); err != nil {
		fmt.Printf("seed admin: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Web admin created/updated. Use email login on the dashboard.")
}
