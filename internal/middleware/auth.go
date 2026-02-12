package middleware

import (
	"net/http"
	"strings"

	"ewallet/internal/service"
	"ewallet/internal/utils"

	"github.com/labstack/echo/v4"
)

const (
	HEADER_AUTH_NAME string = "Authorization"
	AUTH_NAME        string = "Bearer"
	USER_ID          string = "user_id"
)

func JWT(secret string, authService service.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get(HEADER_AUTH_NAME)
			if auth == "" {
				return utils.RespErr(c, http.StatusUnauthorized, "missing authorization header", nil)
			}

			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], AUTH_NAME) {
				return utils.RespErr(c, http.StatusUnauthorized, "invalid authorization header", nil)
			}

			userID, err := utils.ParseToken(parts[1], secret)
			if err != nil {
				return utils.RespErr(c, http.StatusUnauthorized, "invalid token", nil)
			}

			revoked, err := authService.IsTokenRevoked(parts[1])
			if err != nil {
				return utils.RespErr(c, http.StatusInternalServerError, "auth check failed", nil)
			}
			if revoked {
				return utils.RespErr(c, http.StatusUnauthorized, "token revoked", nil)
			}

			c.Set(USER_ID, userID)
			return next(c)
		}
	}
}
