package common

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type config struct{}

func newConfig() config {
	c := config{}
	c.Load()
	return c
}

func (c *config) Load() {
	godotenv.Load(".env." + os.Getenv("APP_ENV"))
}

func (c *config) Get(key string) string {
	if res, ok := os.LookupEnv(key); ok {
		return res
	}
	log.Fatalf("%s - no such key in ENV", key)
	return ""
}

var Config = newConfig()
