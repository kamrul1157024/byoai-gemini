package middlewares

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kamrul1157024/byoai-gemini/internal/loggers"
)

func JSONLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()
		elapsed := time.Since(start)
		loggers.AppLogger.Info("Request Details:",
			slog.Int("status", ctx.Writer.Status()),
			slog.String("method", ctx.Request.Method),
			slog.String("path", ctx.Request.URL.Path),
			slog.Any("duration", elapsed.String()),
			slog.Any("query_params", ctx.Request.URL.Query()),
		)
	}
}
