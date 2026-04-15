package main

import (
	"log"

	"maintenance-dashboard/internal/bootstrap"
)

func main() {
	app, err := bootstrap.NewApp()
	if err != nil {
		log.Fatalf("bootstrap failed: %v", err)
	}
	if err := app.Run(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
