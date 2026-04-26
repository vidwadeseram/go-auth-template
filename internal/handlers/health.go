package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Health(c *gin.Context) {
	if h.db != nil {
		if sqlDB, err := h.db.DB(); err == nil {
			if pingErr := sqlDB.PingContext(c.Request.Context()); pingErr != nil {
				WriteError(c, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "Database is unavailable.")
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
