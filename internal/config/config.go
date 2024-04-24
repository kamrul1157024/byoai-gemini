package config

import (
	"encoding/json"
	"github.com/kamrul1157024/byoai-gemini/internal/apperror"
	"os"
)

type Config struct {
	GEMINI_API_KEY string `json:"GEMINI_API_KEY"`
}


var AppConfig = Config{}

func LoadCofiguration() {
  configFileData, err := os.ReadFile("config.json")
  apperror.CheckAndPanic(err)

  err = json.Unmarshal(configFileData, &AppConfig)
  apperror.CheckAndPanic(err)
}
