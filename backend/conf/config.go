package config

import (
	"fmt"
	"os"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
)

var (
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string

	ServerPort string
)

func LoadDBConfig() {
	DBHost = getCompulsoryEnv("DB_HOST")
	DBPort = getCompulsoryEnvInt("DB_PORT")
	DBUser = getCompulsoryEnv("DB_USER")
	DBPassword = getCompulsoryEnv("DB_PASSWORD")
	DBName = getCompulsoryEnv("DB_NAME")
}

func LoadServerConfig() {
	LoadDBConfig()
	ServerPort = getCompulsoryEnv("SERVER_PORT")
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
