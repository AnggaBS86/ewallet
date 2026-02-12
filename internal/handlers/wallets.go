package handlers

import (
	"errors"
	"net/http"

	"ewallet/internal/dto"
	"ewallet/internal/service"
	"ewallet/internal/utils"

	"github.com/labstack/echo/v4"
)

type WalletHandler struct {
	walletService service.WalletService
}

func NewWalletHandler(walletService service.WalletService) *WalletHandler {
	return &WalletHandler{walletService: walletService}
}

func (h *WalletHandler) TopUp(c echo.Context) error {
	userID := c.Get("user_id").(uint)

	var req dto.TopUpRequest
	if err := c.Bind(&req); err != nil {
		return utils.RespErr(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	res, err := h.walletService.TopUp(userID, req.Amount)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return utils.RespErr(c, http.StatusNotFound, "wallet not found", nil)
		}
		return utils.RespErr(c, http.StatusInternalServerError, "topup failed", nil)
	}

	return utils.RespOK(c, res)
}

func (h *WalletHandler) Balance(c echo.Context) error {
	userID := c.Get("user_id").(uint)

	res, err := h.walletService.Balance(userID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return utils.RespErr(c, http.StatusNotFound, "wallet not found", nil)
		}
		return utils.RespErr(c, http.StatusInternalServerError, "failed to fetch balance", nil)
	}

	return utils.RespOK(c, res)
}
