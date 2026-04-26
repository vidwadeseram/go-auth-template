package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/vidwadeseram/go-auth-template/internal/config"
	"github.com/vidwadeseram/go-auth-template/internal/dto"
	"github.com/vidwadeseram/go-auth-template/internal/models"
)

type TokenClaims struct {
	TokenType string `json:"type,omitempty"`
	jwt.RegisteredClaims
}

type TokenService struct {
	cfg *config.Config
}

func NewTokenService(cfg *config.Config) *TokenService {
	return &TokenService{cfg: cfg}
}

func (s *TokenService) CreateAccessToken(userID uuid.UUID) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := time.Now().UTC().Add(time.Duration(s.cfg.JWTAccessExpireMinutes) * time.Minute)
	claims := TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ID:        uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token, err := s.sign(claims)
	return token, expiresAt, err
}

func (s *TokenService) CreateRefreshToken(userID uuid.UUID) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := time.Now().UTC().Add(time.Duration(s.cfg.JWTRefreshExpireDays) * 24 * time.Hour)
	claims := TokenClaims{
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ID:        uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token, err := s.sign(claims)
	return token, expiresAt, err
}

func (s *TokenService) CreateVerificationToken(userID uuid.UUID, email string) (string, error) {
	claims := TokenClaims{
		TokenType: "verify",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ID:        uuid.NewString(),
			Audience:  jwt.ClaimStrings{email},
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
	}
	return s.sign(claims)
}

func (s *TokenService) IssueTokenPair(userID uuid.UUID) (*dto.TokenData, *models.RefreshToken, error) {
	accessToken, _, err := s.CreateAccessToken(userID)
	if err != nil {
		return nil, nil, err
	}
	refreshToken, refreshExpiresAt, err := s.CreateRefreshToken(userID)
	if err != nil {
		return nil, nil, err
	}

	return &dto.TokenData{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    s.cfg.JWTAccessExpireMinutes * 60,
		}, &models.RefreshToken{
			UserID:    userID,
			TokenHash: s.HashToken(refreshToken),
			ExpiresAt: refreshExpiresAt,
		}, nil
}

func (s *TokenService) ParseToken(token string, expectedType string) (*TokenClaims, error) {
	claims := &TokenClaims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !parsed.Valid {
		return nil, dto.NewAppError(401, "INVALID_TOKEN", "Token is invalid.")
	}
	if expectedType != "" {
		tokenType := claims.TokenType
		if tokenType == "" {
			tokenType = "access"
		}
		if tokenType != expectedType {
			return nil, dto.NewAppError(401, "INVALID_TOKEN_TYPE", "Token type is invalid.")
		}
	}
	if claims.Subject == "" {
		return nil, dto.NewAppError(401, "INVALID_TOKEN", "Token subject is missing.")
	}
	return claims, nil
}

func (s *TokenService) Subject(claims *TokenClaims) (uuid.UUID, error) {
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, dto.NewAppError(401, "INVALID_TOKEN", "Token subject is invalid.")
	}
	return userID, nil
}

func (s *TokenService) HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (s *TokenService) sign(claims TokenClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}
