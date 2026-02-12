package gormrepo

import (
	"errors"

	"ewallet/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RevokedTokenRepository struct {
	db *gorm.DB
}

func NewRevokedTokenRepository(db *gorm.DB) *RevokedTokenRepository {
	return &RevokedTokenRepository{db: db}
}

func (r *RevokedTokenRepository) IsRevoked(token string) (bool, error) {
	var revoked models.RevokedToken
	err := r.db.Where("token = ?", token).First(&revoked).Error
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

func (r *RevokedTokenRepository) Revoke(token string) error {
	return r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.RevokedToken{Token: token}).Error
}
