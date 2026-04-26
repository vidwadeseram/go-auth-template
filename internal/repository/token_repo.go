package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/vidwadeseram/go-auth-template/internal/models"
	"gorm.io/gorm"
)

type TokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) CreateWithTx(ctx context.Context, tx *gorm.DB, token *models.RefreshToken) error {
	return tx.WithContext(ctx).Create(token).Error
}

func (r *TokenRepository) FindActiveByHash(ctx context.Context, tokenHash string, userID uuid.UUID) (*models.RefreshToken, error) {
	var token models.RefreshToken
	err := r.db.WithContext(ctx).Where("token_hash = ? AND user_id = ? AND revoked_at IS NULL", tokenHash, userID).First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *TokenRepository) RevokeWithTx(ctx context.Context, tx *gorm.DB, tokenID uuid.UUID, revokedAt time.Time) error {
	return tx.WithContext(ctx).Model(&models.RefreshToken{}).Where("id = ?", tokenID).Update("revoked_at", revokedAt).Error
}
