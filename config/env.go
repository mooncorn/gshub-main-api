package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var Env Environment

type Environment struct {
	AppEnv    string
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

	LatestServerAPIVersion string
}

func LoadEnv() {
	env := os.Getenv("APP_ENV")
	log.Printf("APP_ENV set to \"%s\"", env)

	if !strings.EqualFold(env, "production") {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	Env = Environment{
		AppEnv:    os.Getenv("APP_ENV"),
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
