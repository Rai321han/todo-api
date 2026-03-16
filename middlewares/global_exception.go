package middlewares

import (
	"todo-api/utils"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
)

// RecoveryFilterChain is a Beego filter function that recovers from panics in the application and returns a standardized JSON error response.
// It captures any panic that occurs during the processing of a request, logs the error, and responds with a 500 Internal Server Error status and a generic error message.
// This filter should be inserted into the filter chain to ensure that all panics are handled gracefully without crashing the server.
func RecoveryFilterChain(next beego.FilterFunc) beego.FilterFunc {
	return func(ctx *context.Context) {
		defer func() {
			if err := recover(); err != nil {
				logs.Error("PANIC Recovery: ", err)
				utils.RespondWithError(ctx, 500, "internal server error.")
			}
		}()
		next(ctx) // controller runs here, inside the defer's scope
	}
}
