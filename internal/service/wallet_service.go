package service

import (
	"errors"

	"ewallet/internal/dto"
	"ewallet/internal/repository"
)

type walletService struct {
	wallets repository.WalletRepository
}

func NewWalletService(wallets repository.WalletRepository) WalletService {
	return &walletService{wallets: wallets}
}

func (s *walletService) TopUp(userID uint, amount int64) (*dto.WalletBalanceResponse, error) {
	balance, err := s.wallets.TopUp(userID, amount)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &dto.WalletBalanceResponse{Balance: balance}, nil
}

func (s *walletService) Balance(userID uint) (*dto.WalletBalanceResponse, error) {
	wallet, err := s.wallets.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &dto.WalletBalanceResponse{Balance: wallet.Balance}, nil
}
