package utils

import (
	"github.com/beego/beego/v2/server/web/context"
) 

type APIError struct {
	Error    string `json:"error"`
}

func RespondWithError(ctx *context.Context, statusCode int, message string) {
	ctx.Output.SetStatus(statusCode)
	resp := APIError{
		Error:    message,
	}
	ctx.Output.JSON(resp, false, false)
}