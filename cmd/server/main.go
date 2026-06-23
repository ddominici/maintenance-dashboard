package main

import (
	"fmt"
	"log"
	"os"

	"maintenance-dashboard/internal/bootstrap"
	"maintenance-dashboard/internal/infra/crypto"
)

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "encrypt" {
		runEncrypt()
		return
	}

	app, err := bootstrap.NewApp()
	if err != nil {
		log.Fatalf("bootstrap failed: %v", err)
	}
	if err := app.Run(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func runEncrypt() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: maintenance-dashboard encrypt <plaintext>")
		fmt.Fprintln(os.Stderr, "       ENCRYPTION_KEY must be set")
		os.Exit(1)
	}
	key, err := crypto.LoadKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if key == nil {
		fmt.Fprintln(os.Stderr, "error: ENCRYPTION_KEY environment variable is not set")
		os.Exit(1)
	}
	encrypted, err := crypto.Encrypt(key, os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(encrypted)
}
