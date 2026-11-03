package main

import (
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	token := os.Getenv("GITHUB_TOKEN")
}
