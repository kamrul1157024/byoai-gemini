package main

import (
	"github.com/gin-gonic/gin"
	"github.com/kamrul1157024/byoai-gemini/apis/handler"
	"github.com/kamrul1157024/byoai-gemini/internal/config"
)

func main() {
  config.LoadCofiguration();
	engine := gin.Default()
	apis.AddRoutesForGeminiAI(engine)
	apis.AddRoutesForHealthCheck(engine)

	engine.SetTrustedProxies(nil)
	engine.Run("0.0.0.0:8000")
}
