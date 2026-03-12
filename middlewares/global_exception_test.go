package middlewares

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
	. "github.com/smartystreets/goconvey/convey"
)

func newGlobalExceptionTestContext(t *testing.T) (*context.Context, *httptest.ResponseRecorder) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/v1/api/todos/", nil)
	res := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(res, req)

	return ctx, res
}

func TestRecoveryFilterChainRecoversPanic(t *testing.T) {
	Convey("RecoveryFilterChain should recover panic and return standardized error", t, func() {
		ctx, res := newGlobalExceptionTestContext(t)

		wrapped := RecoveryFilterChain(func(ctx *context.Context) {
			panic("a panic occurred in the controller")
		})

		// wrapper should swallow panic and convert it to API error response
		wrapped(ctx)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)

		var payload map[string]interface{}
		err := json.Unmarshal(res.Body.Bytes(), &payload)
		So(err, ShouldBeNil)
		So(payload["error"], ShouldEqual, "internal server error.")
		So(payload["code"], ShouldEqual, float64(500))
	})
}

func TestRecoveryFilterChainPassThrough(t *testing.T) {
	Convey("RecoveryFilterChain should pass through when no panic occurs", t, func() {
		ctx, res := newGlobalExceptionTestContext(t)
		nextCalled := false

		wrapped := RecoveryFilterChain(func(ctx *context.Context) {
			nextCalled = true
			ctx.Output.SetStatus(http.StatusOK)
		})

		wrapped(ctx)

		So(nextCalled, ShouldBeTrue)
		So(res.Code, ShouldEqual, http.StatusOK)
	})
}


