package main

import (
	"log"

	"github.com/sraj/everest/internal/app"
)

func main() {
	a, err := app.New()
	if err != nil {
		log.Fatalf("failed to bootstrap: %v", err)
	}
	defer a.Close()

	if err := a.Run(); err != nil {
		log.Fatalf("server stopped with error: %v", err)
	}
}
