package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func GetEnvVar(key string) string {
	err := godotenv.Load(".env")

	if err != nil {
		log.Println("No .env file found, proceeding with system environment variables")
	}
	return os.Getenv(key)
}
