package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var (
	DbHost     string
	DbPort     int
	DbUser     string
	DbPassword string
	DbName     string
)

func LoadConfig() {
	godotenv.Load()
	DbHost = getCompulsoryEnv("DB_HOST")
	DbPort = getCompulsoryEnvInt("DB_PORT")
	DbUser = getCompulsoryEnv("DB_USER")
	DbPassword = getCompulsoryEnv("DB_PASSWORD")
	DbName = getCompulsoryEnv("DB_NAME")
}

func getCompulsoryEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		message := fmt.Sprintf("missing %s", key)
		panic(message)
	}

	return value
}

func getCompulsoryEnvInt(key string) int {
	valueString := getCompulsoryEnv(key)
	valueInt, err := strconv.Atoi(valueString)
	if err != nil {
		panic("error parsing %s to int")
	}

	return valueInt
}
