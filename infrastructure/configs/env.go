package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: No .env file found - using system environment variables")
	}
}

func GetMongoURI() string {
	return os.Getenv("MONGODB_URI")
}

func GetDatabaseName() string {
	return os.Getenv("DATABASE_NAME")
}

func GetServerPort() string {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		return "8080"
	}
	return port
}
