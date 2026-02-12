package handlers

import (
	"net/http"

	"ewallet/internal/utils"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

func HTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	if ve, ok := err.(validator.ValidationErrors); ok {
		details := map[string]string{}
		for _, fe := range ve {
			details[fe.Field()] = fe.Tag()
		}
		_ = utils.RespErr(c, http.StatusBadRequest, "validation error", details)
		return
	}

	if he, ok := err.(*echo.HTTPError); ok {
		message := "request error"
		if he.Message != nil {
			if s, ok := he.Message.(string); ok {
				message = s
			}
		}
		_ = utils.RespErr(c, he.Code, message, nil)
		return
	}

	_ = utils.RespErr(c, http.StatusInternalServerError, "internal server error", nil)
}
