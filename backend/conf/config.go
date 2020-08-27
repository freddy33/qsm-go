package conf

import (
	"github.com/freddy33/qsm-go/m3util"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string

	ServerPort string
}

func NewDBConfig() Config {
	config := Config{
		DBHost:     m3util.GetCompulsoryEnv("DB_HOST"),
		DBPort:     m3util.GetCompulsoryEnvInt("DB_PORT"),
		DBUser:     m3util.GetCompulsoryEnv("DB_USER"),
		DBPassword: m3util.GetCompulsoryEnv("DB_PASSWORD"),
		DBName:     m3util.GetCompulsoryEnv("DB_NAME"),
	}

	return config
}

func NewServerConfig() Config {
	config := Config{
		ServerPort: m3util.GetCompulsoryEnv("SERVER_PORT"),
	}

	return config
}
