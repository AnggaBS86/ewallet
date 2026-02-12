package gormrepo

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"ewallet/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newTxRepo(t *testing.T) (*TransactionRepository, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock init failed: %v", err)
	}
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("gorm open failed: %v", err)
	}
	return NewTransactionRepository(db), mock
}

func TestTransferInsufficientBalance(t *testing.T) {
	repo, mock := newTxRepo(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM \"wallets\" WHERE user_id = $1 ORDER BY \"wallets\".\"id\" LIMIT $2 FOR UPDATE")).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "balance", "created_at"}).AddRow(10, 1, 500, time.Now()))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM \"wallets\" WHERE user_id = $1 ORDER BY \"wallets\".\"id\" LIMIT $2 FOR UPDATE")).
		WithArgs(2, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "balance", "created_at"}).AddRow(11, 2, 100, time.Now()))
	mock.ExpectRollback()

	_, err := repo.Transfer(1, 2, 1000)
	if !errors.Is(err, repository.ErrInsufficientBalance) {
		t.Fatalf("expected ErrInsufficientBalance, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations not met: %v", err)
	}
}

func TestTransferSuccess(t *testing.T) {
	repo, mock := newTxRepo(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM \"wallets\" WHERE user_id = $1 ORDER BY \"wallets\".\"id\" LIMIT $2 FOR UPDATE")).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "balance", "created_at"}).AddRow(10, 1, 5000, time.Now()))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM \"wallets\" WHERE user_id = $1 ORDER BY \"wallets\".\"id\" LIMIT $2 FOR UPDATE")).
		WithArgs(2, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "balance", "created_at"}).AddRow(11, 2, 100, time.Now()))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE \"wallets\" SET \"balance\"=$1 WHERE \"id\" = $2")).WithArgs(4000, 10).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE \"wallets\" SET \"balance\"=$1 WHERE \"id\" = $2")).WithArgs(1100, 11).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO \"transactions\" (\"sender_id\",\"receiver_id\",\"amount\",\"status\",\"created_at\") VALUES ($1,$2,$3,$4,$5) RETURNING \"id\"")).
		WithArgs(1, 2, 1000, "completed", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	tx, err := repo.Transfer(1, 2, 1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.ID != 1 {
		t.Fatalf("unexpected transaction ID: %d", tx.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations not met: %v", err)
	}
}

func TestTransferLocksAscendingOrder(t *testing.T) {
	repo, mock := newTxRepo(t)

	mock.ExpectBegin()
	// sender=2 receiver=1 should still lock user_id=1 first then user_id=2.
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM \"wallets\" WHERE user_id = $1 ORDER BY \"wallets\".\"id\" LIMIT $2 FOR UPDATE")).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "balance", "created_at"}).AddRow(11, 1, 100, time.Now()))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM \"wallets\" WHERE user_id = $1 ORDER BY \"wallets\".\"id\" LIMIT $2 FOR UPDATE")).
		WithArgs(2, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "balance", "created_at"}).AddRow(10, 2, 5000, time.Now()))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE \"wallets\" SET \"balance\"=$1 WHERE \"id\" = $2")).WithArgs(4000, 10).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE \"wallets\" SET \"balance\"=$1 WHERE \"id\" = $2")).WithArgs(1100, 11).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO \"transactions\" (\"sender_id\",\"receiver_id\",\"amount\",\"status\",\"created_at\") VALUES ($1,$2,$3,$4,$5) RETURNING \"id\"")).
		WithArgs(2, 1, 1000, "completed", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(9))
	mock.ExpectCommit()

	_, err := repo.Transfer(2, 1, 1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations not met: %v", err)
	}
}
