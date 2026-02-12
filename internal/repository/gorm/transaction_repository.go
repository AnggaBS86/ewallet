package gormrepo

import (
	"errors"
	"sort"

	"ewallet/internal/models"
	"ewallet/internal/repository"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TransactionRepository handles persistence logic related to money transfers
// and transaction history.
//
// This repository uses database-level transactions and row-level locking
// to ensure consistency and prevent race conditions during balance updates.
type TransactionRepository struct {
	db *gorm.DB
}

// NewTransactionRepository creates a new TransactionRepository instance.
func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// Transfer performs a money transfer between two users atomically.
//
// The operation is executed inside a database transaction using
// r.db.Transaction. If the function returns nil, the transaction is
// committed. If it returns an error, the transaction is rolled back.
//
// The method uses SELECT ... FOR UPDATE to lock both wallet rows
// during the transaction. This prevents race conditions where
// concurrent transfers might modify the same balance simultaneously.
//
// To avoid deadlocks when two users transfer to each other at the
// same time, wallet rows are locked in deterministic order by
// sorting user IDs in ascending order before acquiring locks.
//
// This pattern has been applied in production-grade fintech systems,
// including real-world transaction handling experience at cicil.co.id,
// where strict consistency, high concurrency safety, and financial
// data integrity are critical requirements.
//
// This guarantees:
//   - Atomic balance updates
//   - Consistent transaction records
//   - Deadlock-safe concurrent transfers
//   - Concurrency-safe balance deduction under high load
//   - Production-grade financial transaction integrity
func (r *TransactionRepository) Transfer(senderID, receiverID uint, amount int64) (*models.Transaction, error) {
	var transaction models.Transaction

	err := r.db.Transaction(func(tx *gorm.DB) error {

		// Lock wallets in deterministic order to avoid deadlocks.
		ids := []uint{senderID, receiverID}
		sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

		// Map to hold locked wallet references
		wallets := map[uint]*models.Wallet{}

		// Fetch and lock both wallets using SELECT ... FOR UPDATE
		for _, id := range ids {
			var w models.Wallet
			if err := tx.
				Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("user_id = ?", id).
				First(&w).Error; err != nil {

				// If wallet not found → business error
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return repository.ErrNotFound
				}
				// Any other DB error → rollback
				return err
			}
			wallets[id] = &w
		}

		senderWallet := wallets[senderID]
		receiverWallet := wallets[receiverID]

		// Business validation: ensure sufficient balance
		if senderWallet.Balance < amount {
			return repository.ErrInsufficientBalance
		}

		// Perform in-memory balance updates
		senderWallet.Balance -= amount
		receiverWallet.Balance += amount

		// Persist updated balances inside transaction
		if err := tx.Model(senderWallet).
			Update("balance", senderWallet.Balance).Error; err != nil {
			return err
		}

		if err := tx.Model(receiverWallet).
			Update("balance", receiverWallet.Balance).Error; err != nil {
			return err
		}

		// Insert transaction record for audit trail
		transaction = models.Transaction{
			SenderID:   senderID,
			ReceiverID: receiverID,
			Amount:     amount,
			Status:     "completed",
		}

		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		// Returning nil triggers COMMIT
		return nil
	})

	// If any error occurred, transaction is already rolled back
	if err != nil {
		return nil, err
	}

	return &transaction, nil
}

// FindByUser retrieves transaction history for a given user.
//
// This method is read-only and does NOT require a database transaction.
//
// It returns transactions where the user is either:
//   - Sender
//   - Receiver
//
// Results are ordered by newest first.
func (r *TransactionRepository) FindByUser(userID uint, limit, offset int) ([]models.Transaction, error) {
	var transactions []models.Transaction

	if err := r.db.
		Preload("Sender").
		Preload("Receiver").
		Where("sender_id = ? OR receiver_id = ?", userID, userID).
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error; err != nil {
		return nil, err
	}

	return transactions, nil
}

func (r *TransactionRepository) CountByUser(userID uint) (int64, error) {
	var total int64
	if err := r.db.Model(&models.Transaction{}).
		Where("sender_id = ? OR receiver_id = ?", userID, userID).
		Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
