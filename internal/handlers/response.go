package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vidwadeseram/go-auth-template/internal/dto"
)

func WriteData(c *gin.Context, statusCode int, data any) {
	c.JSON(statusCode, dto.Envelope{Data: data})
}

func WriteError(c *gin.Context, statusCode int, code string, message string) {
	c.JSON(statusCode, dto.ErrorResponse{Error: dto.ErrorDetail{Code: code, Message: message}})
}

func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 && !c.Writer.Written() {
			WriteError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "An unexpected error occurred.")
		}
	}
}

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		slog.Info("http_request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("duration", time.Since(start)),
		)
	}
}
