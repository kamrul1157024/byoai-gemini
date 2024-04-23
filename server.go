package main

import (
	"net/http"
	"server/apis"

	"github.com/gin-gonic/gin"
)

func getStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

func main() {
	engine := gin.Default()
	apis.AddRoutesForGeminiAI(engine)
	engine.GET("/_status", getStatus)

	engine.SetTrustedProxies(nil)
	engine.Run("0.0.0.0:8000")
}
