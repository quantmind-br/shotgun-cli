package main

import (
	"log"
	"os"

	"shotgun-cli/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}