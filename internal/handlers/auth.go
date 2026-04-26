package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vidwadeseram/go-auth-template/internal/dto"
	"github.com/vidwadeseram/go-auth-template/internal/services"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input dto.RegisterRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid request payload.")
		return
	}

	user, err := h.authService.Register(c.Request.Context(), input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	WriteData(c, http.StatusCreated, dto.RegisterData{
		User:    dto.NewUserResponse(user),
		Message: "Registration successful. Verification email sent.",
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input dto.LoginRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid request payload.")
		return
	}

	tokens, err := h.authService.Login(c.Request.Context(), input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	WriteData(c, http.StatusOK, tokens)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var input dto.LogoutRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid request payload.")
		return
	}

	if err := h.authService.Logout(c.Request.Context(), input.RefreshToken); err != nil {
		h.handleError(c, err)
		return
	}

	WriteData(c, http.StatusOK, dto.MessageData{Message: "Logout successful."})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var input dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid request payload.")
		return
	}

	tokens, err := h.authService.Refresh(c.Request.Context(), input.RefreshToken)
	if err != nil {
		h.handleError(c, err)
		return
	}

	WriteData(c, http.StatusOK, tokens)
}

func (h *AuthHandler) Me(c *gin.Context) {
	currentUserID, exists := c.Get("currentUserID")
	if !exists {
		WriteError(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "Authentication credentials were not provided.")
		return
	}

	userID, ok := currentUserID.(uuid.UUID)
	if !ok {
		WriteError(c, http.StatusUnauthorized, "INVALID_TOKEN", "Token subject is invalid.")
		return
	}

	user, err := h.authService.Me(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	WriteData(c, http.StatusOK, dto.NewUserResponse(user))
}

func (h *AuthHandler) handleError(c *gin.Context, err error) {
	var appErr *dto.AppError
	if errors.As(err, &appErr) {
		WriteError(c, appErr.StatusCode, appErr.Code, appErr.Message)
		return
	}
	slog.Error("request failed", slog.String("error", err.Error()))
	WriteError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "An unexpected error occurred.")
}
