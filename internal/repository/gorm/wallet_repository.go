package gormrepo

import (
	"errors"

	"ewallet/internal/models"
	"ewallet/internal/repository"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WalletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) FindByUserID(userID uint) (*models.Wallet, error) {
	var wallet models.Wallet
	if err := r.db.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &wallet, nil
}

func (r *WalletRepository) TopUp(userID uint, amount int64) (int64, error) {
	var balance int64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var wallet models.Wallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&wallet).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return repository.ErrNotFound
			}
			return err
		}
		wallet.Balance += amount
		if err := tx.Model(&wallet).Update("balance", wallet.Balance).Error; err != nil {
			return err
		}
		balance = wallet.Balance
		return nil
	})
	return balance, err
}
