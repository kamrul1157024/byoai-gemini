package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func StreamingHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Status(http.StatusOK)
		c.Header("Content-Type", "text/event-stream; charset=utf-8")
		c.Header("Transfer-Encoding", "chunked")
		c.Writer.WriteHeaderNow()
		c.Next()
	}
}
