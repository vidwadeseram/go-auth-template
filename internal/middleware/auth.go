package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vidwadeseram/go-auth-template/internal/repository"
	"github.com/vidwadeseram/go-auth-template/internal/services"
)

type AuthMiddleware struct {
	tokenService *services.TokenService
	userRepo     *repository.UserRepository
}

func NewAuthMiddleware(tokenService *services.TokenService, userRepo *repository.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{tokenService: tokenService, userRepo: userRepo}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := strings.TrimSpace(c.GetHeader("Authorization"))
		parts := strings.SplitN(authorization, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "AUTHENTICATION_REQUIRED", "message": "Authentication credentials were not provided."}})
			c.Abort()
			return
		}

		claims, err := m.tokenService.ParseToken(parts[1], "access")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "INVALID_TOKEN", "message": "Token is invalid."}})
			c.Abort()
			return
		}

		userID, err := m.tokenService.Subject(claims)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "INVALID_TOKEN", "message": "Token subject is invalid."}})
			c.Abort()
			return
		}

		user, err := m.userRepo.GetByID(c.Request.Context(), userID)
		if err != nil || user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "USER_NOT_FOUND", "message": "Authenticated user was not found."}})
			c.Abort()
			return
		}

		c.Set("currentUserID", userID)
		c.Set("currentUser", user)
		c.Next()
	}
}
