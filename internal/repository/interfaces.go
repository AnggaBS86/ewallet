package repository

import "ewallet/internal/models"

type UserRepository interface {
	FindByEmail(email string) (*models.User, error)
	FindByID(id uint) (*models.User, error)
	FindByIDs(ids []uint) ([]models.User, error)
	CreateWithWallet(user *models.User) error
}

type WalletRepository interface {
	FindByUserID(userID uint) (*models.Wallet, error)
	TopUp(userID uint, amount int64) (int64, error)
}

type TransactionRepository interface {
	Transfer(senderID, receiverID uint, amount int64) (*models.Transaction, error)
	FindByUser(userID uint, limit int) ([]models.Transaction, error)
}

type RevokedTokenRepository interface {
	IsRevoked(token string) (bool, error)
	Revoke(token string) error
}
