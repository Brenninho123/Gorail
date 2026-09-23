package main

import (
	"log"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}

	app := NewProject("Gorail", port)
	if err := app.Run(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
