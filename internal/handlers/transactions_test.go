package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"ewallet/internal/dto"
	"ewallet/internal/service"
	"ewallet/internal/validator"

	"github.com/labstack/echo/v4"
)

type responseAPI struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type mockTransactionService struct {
	transferResp *dto.TransferResponse
	transferErr  error
	historyResp  *dto.TransactionHistoryResponse
	historyErr   error
}

func (m *mockTransactionService) Transfer(_ uint, _ dto.TransferRequest) (*dto.TransferResponse, error) {
	return m.transferResp, m.transferErr
}

func (m *mockTransactionService) History(_ uint, _, _ int) (*dto.TransactionHistoryResponse, error) {
	return m.historyResp, m.historyErr
}

func newCtx(t *testing.T, method, path string, body interface{}) echo.Context {
	t.Helper()
	var buf *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal failed: %v", err)
		}
		buf = bytes.NewBuffer(b)
	} else {
		buf = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, path, buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	e := echo.New()
	e.Validator = validator.New()
	return e.NewContext(req, rec)
}

func parseResp(t *testing.T, c echo.Context) responseAPI {
	t.Helper()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)
	var resp responseAPI
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	return resp
}

func TestTransferSuccess(t *testing.T) {
	h := NewTransactionHandler(&mockTransactionService{transferResp: &dto.TransferResponse{TransactionID: 1, Status: "completed"}})
	c := newCtx(t, http.MethodPost, "/api/transactions/transfer", dto.TransferRequest{ReceiverEmail: "bob@example.com", Amount: 1000})
	c.Set("user_id", uint(1))

	if err := h.Transfer(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	resp := parseResp(t, c)
	if resp.Status != "OK" {
		t.Fatalf("expected OK, got %s", resp.Status)
	}
}

func TestTransferSelf(t *testing.T) {
	h := NewTransactionHandler(&mockTransactionService{transferErr: service.ErrSelfTransfer})
	c := newCtx(t, http.MethodPost, "/api/transactions/transfer", dto.TransferRequest{ReceiverEmail: "alice@example.com", Amount: 1000})
	c.Set("user_id", uint(1))

	if err := h.Transfer(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	resp := parseResp(t, c)
	if resp.Message != "cannot transfer to self" {
		t.Fatalf("unexpected message: %s", resp.Message)
	}
}

func TestTransferInsufficientBalance(t *testing.T) {
	h := NewTransactionHandler(&mockTransactionService{transferErr: service.ErrInsufficientBalance})
	c := newCtx(t, http.MethodPost, "/api/transactions/transfer", dto.TransferRequest{ReceiverEmail: "bob@example.com", Amount: 1000})
	c.Set("user_id", uint(1))

	if err := h.Transfer(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	resp := parseResp(t, c)
	if resp.Message != "insufficient balance" {
		t.Fatalf("unexpected message: %s", resp.Message)
	}
}

func TestTransferNotFound(t *testing.T) {
	h := NewTransactionHandler(&mockTransactionService{transferErr: service.ErrReceiverNotFound})
	c := newCtx(t, http.MethodPost, "/api/transactions/transfer", dto.TransferRequest{ReceiverEmail: "unknown@example.com", Amount: 1000})
	c.Set("user_id", uint(1))

	if err := h.Transfer(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	resp := parseResp(t, c)
	if resp.Message != "receiver not found" {
		t.Fatalf("unexpected message: %s", resp.Message)
	}
}

func TestHistorySuccess(t *testing.T) {
	history := &dto.TransactionHistoryResponse{Transactions: []dto.TransactionHistoryItem{{ID: 1, Status: "completed"}}}
	h := NewTransactionHandler(&mockTransactionService{historyResp: history})
	c := newCtx(t, http.MethodGet, "/api/transactions/history?limit=50", nil)
	c.Set("user_id", uint(1))

	if err := h.History(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	resp := parseResp(t, c)
	if resp.Status != "OK" {
		t.Fatalf("expected OK, got %s", resp.Status)
	}
}

func TestHistoryError(t *testing.T) {
	h := NewTransactionHandler(&mockTransactionService{historyErr: errors.New("db error")})
	c := newCtx(t, http.MethodGet, "/api/transactions/history?limit=50", nil)
	c.Set("user_id", uint(1))

	if err := h.History(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	resp := parseResp(t, c)
	if resp.Message != "failed to fetch history" {
		t.Fatalf("unexpected message: %s", resp.Message)
	}
}
