package server

import (
	"github.com/labstack/echo/v4"
	"github.com/raian621/obsync-server/api"
)

func sendApiMessage(ctx echo.Context, code int32, message string) error {
	var res api.ApiResponse
	res.Code = &code
	res.Message = &message
	return ctx.JSON(
		int(*res.Code),
		res,
	)
}
