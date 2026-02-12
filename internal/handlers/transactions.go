package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"ewallet/internal/dto"
	"ewallet/internal/service"
	"ewallet/internal/utils"

	"github.com/labstack/echo/v4"
)

type TransactionHandler struct {
	transactionService service.TransactionService
}

func NewTransactionHandler(transactionService service.TransactionService) *TransactionHandler {
	return &TransactionHandler{transactionService: transactionService}
}

func (h *TransactionHandler) Transfer(c echo.Context) error {
	senderID := c.Get("user_id").(uint)

	var req dto.TransferRequest
	if err := c.Bind(&req); err != nil {
		return utils.RespErr(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	res, err := h.transactionService.Transfer(senderID, req)
	if err != nil {
		if errors.Is(err, service.ErrReceiverNotFound) {
			return utils.RespErr(c, http.StatusNotFound, "receiver not found", nil)
		}
		if errors.Is(err, service.ErrWalletNotFound) {
			return utils.RespErr(c, http.StatusNotFound, "wallet not found", nil)
		}
		if errors.Is(err, service.ErrSelfTransfer) {
			return utils.RespErr(c, http.StatusBadRequest, "cannot transfer to self", nil)
		}
		if errors.Is(err, service.ErrInsufficientBalance) {
			return utils.RespErr(c, http.StatusBadRequest, "insufficient balance", nil)
		}
		return utils.RespErr(c, http.StatusInternalServerError, "transfer failed", nil)
	}

	return utils.RespOK(c, res)
}

func (h *TransactionHandler) History(c echo.Context) error {
	userID := c.Get("user_id").(uint)

	page := 1
	if v := c.QueryParam("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}

	limit := 50
	if v := c.QueryParam("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			if n > 100 {
				n = 100
			}
			limit = n
		}
	}

	res, err := h.transactionService.History(userID, page, limit)
	if err != nil {
		return utils.RespErr(c, http.StatusInternalServerError, "failed to fetch history", nil)
	}

	return utils.RespOK(c, res)
}
