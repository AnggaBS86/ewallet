package handlers

import (
	"errors"
	"net/http"

	"ewallet/internal/service"
	"ewallet/internal/utils"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Profile(c echo.Context) error {
	userID := c.Get("user_id").(uint)

	res, err := h.userService.Profile(userID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return utils.RespErr(c, http.StatusNotFound, "user not found", nil)
		}
		return utils.RespErr(c, http.StatusInternalServerError, "failed to fetch user", nil)
	}

	return utils.RespOK(c, res)
}
