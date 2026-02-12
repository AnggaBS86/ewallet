package service

import (
	"errors"
	"time"

	"ewallet/internal/dto"
	"ewallet/internal/repository"
)

type transactionService struct {
	users        repository.UserRepository
	transactions repository.TransactionRepository
	cache        HistoryCache
}

func NewTransactionService(users repository.UserRepository, transactions repository.TransactionRepository, cache HistoryCache) TransactionService {
	return &transactionService{users: users, transactions: transactions, cache: cache}
}

func (s *transactionService) Transfer(senderID uint, req dto.TransferRequest) (*dto.TransferResponse, error) {
	receiver, err := s.users.FindByEmail(req.ReceiverEmail)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrReceiverNotFound
		}
		return nil, err
	}

	// cannot transfer to itself
	if receiver.ID == senderID {
		return nil, ErrSelfTransfer
	}

	tx, err := s.transactions.Transfer(senderID, receiver.ID, req.Amount)
	if err != nil {
		if errors.Is(err, repository.ErrInsufficientBalance) {
			return nil, ErrInsufficientBalance
		}
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrWalletNotFound
		}
		return nil, err
	}

	// invalidate the cache --> make the cache is empty again
	// if theere are new db updates
	if s.cache != nil {
		s.cache.InvalidateUser(senderID)
		s.cache.InvalidateUser(receiver.ID)
	}

	return &dto.TransferResponse{TransactionID: tx.ID, Status: tx.Status}, nil
}

func (s *transactionService) History(userID uint, page, limit int) (*dto.TransactionHistoryResponse, error) {
	if page < 1 {
		page = 1
	}

	if s.cache != nil {
		if cached, ok := s.cache.Get(userID, page, limit); ok {
			return cached, nil
		}
	}

	offset := (page - 1) * limit
	transactions, err := s.transactions.FindByUser(userID, limit, offset)
	if err != nil {
		return nil, err
	}
	total, err := s.transactions.CountByUser(userID)
	if err != nil {
		return nil, err
	}

	items := make([]dto.TransactionHistoryItem, 0, len(transactions))
	for _, tx := range transactions {
		items = append(items, dto.TransactionHistoryItem{
			ID:        tx.ID,
			Sender:    dto.UserInfo{ID: tx.Sender.ID, Name: tx.Sender.Name, Email: tx.Sender.Email},
			Receiver:  dto.UserInfo{ID: tx.Receiver.ID, Name: tx.Receiver.Name, Email: tx.Receiver.Email},
			Amount:    tx.Amount,
			Status:    tx.Status,
			CreatedAt: tx.CreatedAt.Format(time.RFC3339),
		})
	}

	totalPages := 0
	if limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}

	res := &dto.TransactionHistoryResponse{
		Transactions: items,
		Pagination: dto.PaginationMeta{
			Page:       page,
			Limit:      limit,
			Total:      int(total),
			TotalPages: totalPages,
		},
	}
	if s.cache != nil {
		s.cache.Set(userID, page, limit, res)
	}
	return res, nil
}
