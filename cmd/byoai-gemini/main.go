package main

import (
	"github.com/gin-gonic/gin"
	"github.com/kamrul1157024/byoai-gemini/apis/handler"
)

func main() {
	engine := gin.Default()
	apis.AddRoutesForGeminiAI(engine)
	apis.AddRoutesForHealthCheck(engine)

	engine.SetTrustedProxies(nil)
	engine.Run("0.0.0.0:8000")
}
