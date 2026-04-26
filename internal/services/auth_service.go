package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/vidwadeseram/go-auth-template/internal/config"
	"github.com/vidwadeseram/go-auth-template/internal/dto"
	"github.com/vidwadeseram/go-auth-template/internal/mailer"
	"github.com/vidwadeseram/go-auth-template/internal/models"
	"github.com/vidwadeseram/go-auth-template/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo     *repository.UserRepository
	tokenRepo    *repository.TokenRepository
	tokenService *TokenService
	mailer       *mailer.Mailer
	db           *gorm.DB
	validator    *validator.Validate
}

func NewAuthService(userRepo *repository.UserRepository, tokenRepo *repository.TokenRepository, tokenService *TokenService, mailer *mailer.Mailer, db *gorm.DB, cfg *config.Config) *AuthService {
	_ = cfg
	return &AuthService{userRepo: userRepo, tokenRepo: tokenRepo, tokenService: tokenService, mailer: mailer, db: db, validator: validator.New()}
}

func (s *AuthService) Register(ctx context.Context, input dto.RegisterRequest) (*models.User, error) {
	if err := s.validator.Struct(input); err != nil {
		return nil, dto.ValidationError(err)
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))
	existing, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	if existing != nil {
		return nil, dto.NewAppError(400, "EMAIL_ALREADY_EXISTS", "A user with this email already exists.")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &models.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		IsActive:     true,
		IsVerified:   false,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	verificationToken, err := s.tokenService.CreateVerificationToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("create verification token: %w", err)
	}

	body := fmt.Sprintf("Welcome %s, your verification token is: %s", user.FullName(), verificationToken)
	if err := s.mailer.Send(user.Email, "Verify your account", body); err != nil {
		return nil, fmt.Errorf("send verification email: %w", err)
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, input dto.LoginRequest) (*dto.TokenData, error) {
	if err := s.validator.Struct(input); err != nil {
		return nil, dto.ValidationError(err)
	}

	user, err := s.userRepo.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(input.Email)))
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	if user == nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) != nil {
		return nil, dto.NewAppError(401, "INVALID_CREDENTIALS", "Invalid email or password.")
	}
	if !user.IsActive {
		return nil, dto.NewAppError(403, "USER_INACTIVE", "User account is inactive.")
	}

	var tokens *dto.TokenData
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		pair, refreshModel, issueErr := s.tokenService.IssueTokenPair(user.ID)
		if issueErr != nil {
			return issueErr
		}
		if createErr := s.tokenRepo.CreateWithTx(ctx, tx, refreshModel); createErr != nil {
			return fmt.Errorf("create refresh token: %w", createErr)
		}
		tokens = pair
		return nil
	})
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	input := dto.RefreshTokenRequest{RefreshToken: refreshToken}
	if err := s.validator.Struct(input); err != nil {
		return dto.ValidationError(err)
	}

	claims, err := s.tokenService.ParseToken(refreshToken, "refresh")
	if err != nil {
		return err
	}
	userID, err := s.tokenService.Subject(claims)
	if err != nil {
		return err
	}

	storedToken, err := s.tokenRepo.FindActiveByHash(ctx, s.tokenService.HashToken(refreshToken), userID)
	if err != nil {
		return fmt.Errorf("find refresh token: %w", err)
	}
	if storedToken == nil {
		return dto.NewAppError(401, "INVALID_REFRESH_TOKEN", "Refresh token is invalid.")
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return s.tokenRepo.RevokeWithTx(ctx, tx, storedToken.ID, time.Now().UTC())
	})
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*dto.TokenData, error) {
	input := dto.RefreshTokenRequest{RefreshToken: refreshToken}
	if err := s.validator.Struct(input); err != nil {
		return nil, dto.ValidationError(err)
	}

	claims, err := s.tokenService.ParseToken(refreshToken, "refresh")
	if err != nil {
		return nil, err
	}
	userID, err := s.tokenService.Subject(claims)
	if err != nil {
		return nil, err
	}

	storedToken, err := s.tokenRepo.FindActiveByHash(ctx, s.tokenService.HashToken(refreshToken), userID)
	if err != nil {
		return nil, fmt.Errorf("find refresh token: %w", err)
	}
	if storedToken == nil || storedToken.ExpiresAt.Before(time.Now().UTC()) {
		return nil, dto.NewAppError(401, "INVALID_REFRESH_TOKEN", "Refresh token is invalid or expired.")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	if user == nil {
		return nil, dto.NewAppError(401, "USER_NOT_FOUND", "Authenticated user was not found.")
	}

	var tokens *dto.TokenData
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if revokeErr := s.tokenRepo.RevokeWithTx(ctx, tx, storedToken.ID, time.Now().UTC()); revokeErr != nil {
			return fmt.Errorf("revoke refresh token: %w", revokeErr)
		}
		pair, refreshModel, issueErr := s.tokenService.IssueTokenPair(userID)
		if issueErr != nil {
			return issueErr
		}
		if createErr := s.tokenRepo.CreateWithTx(ctx, tx, refreshModel); createErr != nil {
			return fmt.Errorf("create refresh token: %w", createErr)
		}
		tokens = pair
		return nil
	})
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *AuthService) Me(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	if user == nil {
		return nil, dto.NewAppError(401, "USER_NOT_FOUND", "Authenticated user was not found.")
	}
	return user, nil
}
