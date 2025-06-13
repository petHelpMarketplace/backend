package app

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	errEnvFileUnavailable = errors.New("env: .env file unavailable")
)

const (
	envFileName = ".env"
)

func Envs() {
	if env := os.Getenv("APP_STAGE"); env != "" && env == "local" {
		if _, err := os.Stat(envFileName); err == nil {
			if err := godotenv.Load(envFileName); err != nil {
				log.Fatalf("Failed to load .env file: %s", err.Error())
			}
		} else {
			panic(errEnvFileUnavailable)
		}

		log.Printf("Loaded .env file: %s", envFileName)
	}
}
