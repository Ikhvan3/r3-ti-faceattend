package main

import (
	"log"

	"r3-ti-faceattend/backend/internal/app"
	"r3-ti-faceattend/backend/internal/config"
)

func main() {
	if err := config.LoadDotEnv(".env"); err != nil {
		log.Fatalf("load environment: %v", err)
	}

	app.Run()
}
