package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vidwadeseram/go-auth-template/internal/config"
	"github.com/vidwadeseram/go-auth-template/internal/database"
	"github.com/vidwadeseram/go-auth-template/internal/handlers"
	"github.com/vidwadeseram/go-auth-template/internal/mailer"
	"github.com/vidwadeseram/go-auth-template/internal/middleware"
	"github.com/vidwadeseram/go-auth-template/internal/repository"
	"github.com/vidwadeseram/go-auth-template/internal/services"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		logger.Error("failed to connect database", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := database.RunMigrations(cfg); err != nil {
		logger.Error("failed to run migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}

	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	tokenService := services.NewTokenService(cfg)
	mailClient := mailer.New(cfg)
	authService := services.NewAuthService(userRepo, tokenRepo, tokenService, mailClient, db, cfg)
	authHandler := handlers.NewAuthHandler(authService)
	healthHandler := handlers.NewHealthHandler(db)
	authMiddleware := middleware.NewAuthMiddleware(tokenService, userRepo)
	rbacRepo := repository.NewRBACRepository(db)
	adminHandler := handlers.NewAdminHandler(rbacRepo, authService, tokenService)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(handlers.RequestLogger())
	router.Use(handlers.ErrorMiddleware())

	router.GET("/health", healthHandler.Health)

	router.StaticFile("/openapi.json", "./static/openapi.json")
	router.StaticFile("/docs", "./static/swagger.html")

	authGroup := router.Group("/api/v1/auth")
	authGroup.Use(middleware.RateLimit(1/60.0, 5))
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/logout", authHandler.Logout)
	authGroup.POST("/refresh", authHandler.Refresh)
	authGroup.GET("/me", authMiddleware.RequireAuth(), authHandler.Me)
	authGroup.POST("/verify-email", authHandler.VerifyEmail)
	authGroup.POST("/forgot-password", authHandler.ForgotPassword)
	authGroup.POST("/reset-password", authHandler.ResetPassword)

	adminGroup := router.Group("/api/v1/admin")
	adminGroup.Use(authMiddleware.RequireAuth())
	adminGroup.GET("/roles", adminHandler.ListRoles)
	adminGroup.GET("/permissions", adminHandler.ListPermissions)
	adminGroup.GET("/roles/:role_id/permissions", adminHandler.GetRolePermissions)
	adminGroup.POST("/roles/permissions", adminHandler.AssignPermissionToRole)
	adminGroup.GET("/users", adminHandler.ListUsers)
	adminGroup.GET("/users/:user_id", adminHandler.GetUser)
	adminGroup.DELETE("/users/:user_id", adminHandler.DeleteUser)
	adminGroup.GET("/users/:user_id/permissions", adminHandler.GetUserPermissions)
	adminGroup.POST("/users/roles", adminHandler.AssignRoleToUser)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.AppPort),
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("starting server", slog.Int("port", cfg.AppPort))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped unexpectedly", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("server stopped")
}
