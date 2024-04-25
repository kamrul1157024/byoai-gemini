package config

import (
	"os"
)

type Config struct {
	GEMINI_API_KEY string `json:"GEMINI_API_KEY"`
}

var AppConfig = Config{GEMINI_API_KEY: os.Getenv("GEMINI_AI_API_KEY")}
