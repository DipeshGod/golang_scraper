package main

import (
	"log"

	"github.com/dipeshgod/go-scraper/internal/server"
)

func main() {
	if err := server.StartServer(":8080"); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
