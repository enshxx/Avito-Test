package main

import (
	"avito/internal/app"
	"github.com/joho/godotenv"
	"log"
)

const configPath = "config/config.yaml"

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	app.Migrations()
	app.Run(configPath)
}
