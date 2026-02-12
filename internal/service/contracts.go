package service

import "ewallet/internal/dto"

type AuthService interface {
	Register(req dto.RegisterRequest) (*dto.UserResponse, error)
	Login(req dto.LoginRequest) (*dto.AuthTokenResponse, error)
	Logout(token string) error
	IsTokenRevoked(token string) (bool, error)
}

type UserService interface {
	Profile(userID uint) (*dto.UserResponse, error)
}

type WalletService interface {
	TopUp(userID uint, amount int64) (*dto.WalletBalanceResponse, error)
	Balance(userID uint) (*dto.WalletBalanceResponse, error)
}

type TransactionService interface {
	Transfer(senderID uint, req dto.TransferRequest) (*dto.TransferResponse, error)
	History(userID uint, page, limit int) (*dto.TransactionHistoryResponse, error)
}
