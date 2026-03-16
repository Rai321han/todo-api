package utils

import (
	"github.com/beego/beego/v2/server/web/context"
)

// APIError represents a standardized error response structure for the API, containing an error message and an HTTP status code.
type APIError struct {
	Error      string `json:"error"`
	StatusCode int    `json:"code"`
}

// RespondWithError is a helper function to send standardized JSON error responses.
// It sets the HTTP status code and returns a JSON object containing the error message and status code.
func RespondWithError(ctx *context.Context, statusCode int, message string) {
	ctx.Output.SetStatus(statusCode)
	resp := APIError{
		Error:      message,
		StatusCode: statusCode,
	}
	ctx.Output.JSON(resp, false, false)
}
