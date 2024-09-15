package errors

import (
	"github.com/labstack/echo/v4"
)

func NewErrorResponse(c echo.Context, status int, err error) error {
	return c.JSON(status, map[string]interface{}{
		"reason": err.Error(),
	})
}
