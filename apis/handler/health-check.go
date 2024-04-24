package apis

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func getAppStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

func AddRoutesForHealthCheck(engine *gin.Engine) {
	engine.GET("/_status", getAppStatus)
}
