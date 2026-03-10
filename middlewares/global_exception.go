package middlewares

import (
	"todo-api/utils"

	"github.com/beego/beego/v2/server/web/context"
)

func RecoveryMiddleware(ctx *context.Context) {
	defer func() {
		if r := recover(); r != nil {
			utils.RespondWithError(ctx, 500, "Internal server error")
		}
	}()
}