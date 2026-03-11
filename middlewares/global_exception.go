package middlewares

import (
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web/context"
)

func RecoveryFilter(ctx *context.Context) {

    defer func() {
        if err := recover(); err != nil {

            logs.Error("PANIC:", err)

            ctx.Output.SetStatus(500)
            ctx.Output.JSON(map[string]interface{}{
                "error": "internal server error",
            }, false, false)
        }
    }()
}