package utils

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type SuccessResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type ErrorResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func RespErr(c echo.Context, status int, message string, details interface{}) error {
	return c.JSON(status, ErrorResponse{
		Status:  "ERROR",
		Message: message,
		Data:    details,
	})
}

func RespOK(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, SuccessResponse{
		Status:  "OK",
		Message: "Response result successfully",
		Data:    data,
	})
}

func RespCreated(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusCreated, SuccessResponse{
		Status:  "OK",
		Message: "Response result successfully",
		Data:    data,
	})
}
