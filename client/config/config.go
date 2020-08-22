package config

import (
	"github.com/freddy33/qsm-go/m3util"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	BackendRootURL string
}

func NewConfig() Config {
	config := Config{
		BackendRootURL: m3util.GetCompulsoryEnv("BACKEND_ROOT_URL"),
	}

	return config
}
