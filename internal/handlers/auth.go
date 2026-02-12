package handlers

import (
	"errors"
	"net/http"
	"strings"

	"ewallet/internal/config"
	"ewallet/internal/dto"
	"ewallet/internal/service"
	"ewallet/internal/utils"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authService service.AuthService
	cfg         config.Config
}

func NewAuthHandler(authService service.AuthService, cfg config.Config) *AuthHandler {
	return &AuthHandler{authService: authService, cfg: cfg}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return utils.RespErr(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	res, err := h.authService.Register(req)
	if err != nil {
		if errors.Is(err, service.ErrConflict) {
			return utils.RespErr(c, http.StatusConflict, "email already registered", nil)
		}
		return utils.RespErr(c, http.StatusInternalServerError, "failed to register user", nil)
	}

	return utils.RespCreated(c, res)
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return utils.RespErr(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	res, err := h.authService.Login(req)
	if err != nil {
		if errors.Is(err, service.ErrUnauthorized) {
			return utils.RespErr(c, http.StatusUnauthorized, "invalid email or password", nil)
		}
		return utils.RespErr(c, http.StatusInternalServerError, "login failed", nil)
	}

	return utils.RespOK(c, res)
}

func (h *AuthHandler) Logout(c echo.Context) error {
	auth := c.Request().Header.Get("Authorization")
	if auth == "" {
		return utils.RespErr(c, http.StatusUnauthorized, "missing authorization header", nil)
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return utils.RespErr(c, http.StatusUnauthorized, "invalid authorization header", nil)
	}
	token := parts[1]

	if _, err := utils.ParseToken(token, h.cfg.JWTSecret); err != nil {
		return utils.RespErr(c, http.StatusUnauthorized, "invalid token", nil)
	}

	if err := h.authService.Logout(token); err != nil {
		return utils.RespErr(c, http.StatusInternalServerError, "logout failed", nil)
	}

	return utils.RespOK(c, dto.MessageResponse{Message: "logged out"})
}
