package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vidwadeseram/go-auth-template/internal/dto"
	"github.com/vidwadeseram/go-auth-template/internal/repository"
	"github.com/vidwadeseram/go-auth-template/internal/services"
)

type AdminHandler struct {
	rbacRepo     *repository.RBACRepository
	authService  *services.AuthService
	tokenService *services.TokenService
}

func NewAdminHandler(rbacRepo *repository.RBACRepository, authService *services.AuthService, tokenService *services.TokenService) *AdminHandler {
	return &AdminHandler{rbacRepo: rbacRepo, authService: authService, tokenService: tokenService}
}

func (h *AdminHandler) requirePermission(c *gin.Context, permissionName string) (uuid.UUID, bool) {
	userID, exists := c.Get("currentUserID")
	if !exists {
		WriteError(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "Authentication required.")
		return uuid.Nil, false
	}
	uid := userID.(uuid.UUID)
	has, err := h.rbacRepo.UserHasPermission(c.Request.Context(), uid, permissionName)
	if err != nil || !has {
		WriteError(c, http.StatusForbidden, "FORBIDDEN", "Permission '"+permissionName+"' is required.")
		return uuid.Nil, false
	}
	return uid, true
}

func (h *AdminHandler) ListRoles(c *gin.Context) {
	if _, ok := h.requirePermission(c, "roles.manage"); !ok {
		return
	}
	roles, err := h.rbacRepo.ListRoles(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to list roles.")
		return
	}
	WriteData(c, http.StatusOK, roles)
}

func (h *AdminHandler) ListPermissions(c *gin.Context) {
	if _, ok := h.requirePermission(c, "roles.manage"); !ok {
		return
	}
	perms, err := h.rbacRepo.ListPermissions(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to list permissions.")
		return
	}
	WriteData(c, http.StatusOK, perms)
}

func (h *AdminHandler) GetRolePermissions(c *gin.Context) {
	if _, ok := h.requirePermission(c, "roles.manage"); !ok {
		return
	}
	roleID, err := uuid.Parse(c.Param("role_id"))
	if err != nil {
		WriteError(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid role ID.")
		return
	}
	perms, err := h.rbacRepo.GetRolePermissions(c.Request.Context(), roleID)
	if err != nil {
		WriteError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to get role permissions.")
		return
	}
	WriteData(c, http.StatusOK, perms)
}

func (h *AdminHandler) AssignPermissionToRole(c *gin.Context) {
	if _, ok := h.requirePermission(c, "roles.manage"); !ok {
		return
	}
	var input dto.RolePermissionRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid request payload.")
		return
	}
	if err := h.rbacRepo.AssignPermissionToRole(c.Request.Context(), input.RoleID, input.PermissionID); err != nil {
		WriteError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to assign permission.")
		return
	}
	WriteData(c, http.StatusOK, dto.MessageData{Message: "Permission assigned to role."})
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	if _, ok := h.requirePermission(c, "users.read"); !ok {
		return
	}
	users, err := h.authService.ListUsers(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to list users.")
		return
	}
	WriteData(c, http.StatusOK, users)
}

func (h *AdminHandler) GetUser(c *gin.Context) {
	if _, ok := h.requirePermission(c, "users.read"); !ok {
		return
	}
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		WriteError(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid user ID.")
		return
	}
	user, err := h.authService.GetUser(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	WriteData(c, http.StatusOK, dto.NewUserResponse(user))
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	if _, ok := h.requirePermission(c, "users.delete"); !ok {
		return
	}
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		WriteError(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid user ID.")
		return
	}
	if err := h.authService.DeleteUser(c.Request.Context(), userID); err != nil {
		h.handleError(c, err)
		return
	}
	WriteData(c, http.StatusOK, dto.MessageData{Message: "User deleted."})
}

func (h *AdminHandler) AssignRoleToUser(c *gin.Context) {
	if _, ok := h.requirePermission(c, "roles.manage"); !ok {
		return
	}
	var input dto.UserRoleRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid request payload.")
		return
	}
	if err := h.rbacRepo.AssignRoleToUser(c.Request.Context(), input.UserID, input.RoleID); err != nil {
		WriteError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to assign role.")
		return
	}
	WriteData(c, http.StatusOK, dto.MessageData{Message: "Role assigned to user."})
}

func (h *AdminHandler) GetUserPermissions(c *gin.Context) {
	if _, ok := h.requirePermission(c, "users.read"); !ok {
		return
	}
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		WriteError(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid user ID.")
		return
	}
	perms, err := h.rbacRepo.GetUserPermissions(c.Request.Context(), userID)
	if err != nil {
		WriteError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to get user permissions.")
		return
	}
	WriteData(c, http.StatusOK, perms)
}

func (h *AdminHandler) handleError(c *gin.Context, err error) {
	var appErr *dto.AppError
	if err == dto.ErrNotFound {
		WriteError(c, http.StatusNotFound, "NOT_FOUND", "User not found.")
		return
	}
	if errors.As(err, &appErr) {
		WriteError(c, appErr.StatusCode, appErr.Code, appErr.Message)
		return
	}
	slog.Error("request failed", slog.String("error", err.Error()))
	WriteError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "An unexpected error occurred.")
}
