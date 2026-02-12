package service

import (
	"errors"
	"sort"
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

func (s *transactionService) History(userID uint, limit int) (*dto.TransactionHistoryResponse, error) {
	if s.cache != nil {
		if cached, ok := s.cache.Get(userID, limit); ok {
			return cached, nil
		}
	}

	transactions, err := s.transactions.FindByUser(userID, limit)
	if err != nil {
		return nil, err
	}

	userIDs := map[uint]struct{}{}
	for _, tx := range transactions {
		userIDs[tx.SenderID] = struct{}{}
		userIDs[tx.ReceiverID] = struct{}{}
	}

	ids := make([]uint, 0, len(userIDs))
	for id := range userIDs {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	users, err := s.users.FindByIDs(ids)
	if err != nil {
		return nil, err
	}

	byID := make(map[uint]dto.UserInfo, len(users))
	for _, u := range users {
		byID[u.ID] = dto.UserInfo{ID: u.ID, Name: u.Name, Email: u.Email}
	}

	items := make([]dto.TransactionHistoryItem, 0, len(transactions))
	for _, tx := range transactions {
		items = append(items, dto.TransactionHistoryItem{
			ID:        tx.ID,
			Sender:    byID[tx.SenderID],
			Receiver:  byID[tx.ReceiverID],
			Amount:    tx.Amount,
			Status:    tx.Status,
			CreatedAt: tx.CreatedAt.Format(time.RFC3339),
		})
	}

	res := &dto.TransactionHistoryResponse{Transactions: items}
	if s.cache != nil {
		s.cache.Set(userID, limit, res)
	}
	return res, nil
}
