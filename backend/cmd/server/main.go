// Command server runs the Sentinel Infrastructure Event Intelligence HTTP service.
package main

import (
	"log"

	"github.com/aadityya4real/sentinel/backend/internal/config"
	"github.com/aadityya4real/sentinel/backend/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	app, err := server.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
