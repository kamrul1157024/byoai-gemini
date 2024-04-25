package main

import (
	"github.com/gin-gonic/gin"
	"github.com/kamrul1157024/byoai-gemini/apis/handler"
	"github.com/kamrul1157024/byoai-gemini/apis/middlewares"
)

func main() {
	engine := gin.New()
	engine.Use(middlewares.JSONLogger())

	aiRouter := engine.Group("/")
	apis.AddRoutesForGeminiAI(aiRouter)

	healthRouter := engine.Group("/")
	apis.AddRoutesForHealthCheck(healthRouter)

	engine.Run("0.0.0.0:8000")
}
