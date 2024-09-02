package common

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
)

type config struct{}

func NewConfig() config {
	c := config{}
	c.Load()
	return c
}

func (c *config) Load() {
	fmt.Printf("Starting with %s\n", os.Getenv("APP_ENV"))
	godotenv.Load(".env." + os.Getenv("APP_ENV"))
}

func (c *config) Get(key string) string {
	if res, ok := os.LookupEnv(key); ok {
		fmt.Printf("Used env: %s, val: %s\n", key, res)
		return res
	}
	log.Fatalf("%s - no such key in ENV", key)
	return ""
}

var Config config
