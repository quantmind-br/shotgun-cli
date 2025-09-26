package main

import (
	"log"

	"example.com/sample-project/internal/handlers"
)

func main() {
	if err := handlers.Serve(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
