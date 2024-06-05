package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var Env Environment

type Environment struct {
	GinMode   string
	DSN       string
	URL       string
	Port      string
	JWTSecret string

	AWSAccessKey       string
	AWSSecretAccessKey string
	AWSRegion          string
	AWSImageIdBase     string
	AWSKeyPairName     string

	GoogleClientId     string
	GoogleClientSecret string

	LocalStackEndpoint string
}

func LoadEnv() {
	env := os.Getenv("APP_ENV")
	var envFile string

	switch env {
	case "production":
		envFile = ".env.production"
	case "development":
		envFile = ".env.development"
	default:
		log.Fatal("APP_ENV has to be set to \"production\" or \"development\"")
	}

	log.Printf("APP_ENV set to \"%s\"", env)

	err := godotenv.Load(envFile)
	if err != nil {
		log.Fatalf("Error loading %s file", envFile)
	}

	Env = Environment{
		GinMode:   os.Getenv("GIN_MODE"),
		DSN:       os.Getenv("DSN"),
		URL:       os.Getenv("URL"),
		Port:      os.Getenv("PORT"),
		JWTSecret: os.Getenv("JWT_SECRET"),

		AWSAccessKey:       os.Getenv("AWS_ACCESS_KEY"),
		AWSSecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		AWSRegion:          os.Getenv("AWS_REGION"),
		AWSImageIdBase:     os.Getenv("AWS_IMAGE_ID_BASE"),
		AWSKeyPairName:     os.Getenv("AWS_KEY_PAIR_NAME"),

		GoogleClientId:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),

		LocalStackEndpoint: os.Getenv("LOCALSTACK_ENDPOINT"),
	}
}
