package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	. "github.com/smartystreets/goconvey/convey"
)

func newRequestLoggerTestContext(t *testing.T) (*context.Context, *httptest.ResponseRecorder) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/v1/api/todos/", nil)
	res := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(res, req)

	return ctx, res
}

func TestNewResponseWriter(t *testing.T) {
	Convey("newResponseWriter should default status code to 200 and track updates", t, func() {
		res := httptest.NewRecorder()
		rw := newResponseWriter(res)

		So(rw.statusCode, ShouldEqual, http.StatusOK)

		rw.WriteHeader(http.StatusCreated)
		So(rw.statusCode, ShouldEqual, http.StatusCreated)
		So(res.Code, ShouldEqual, http.StatusCreated)
	})
}

func TestRequestLoggerNonDevMode(t *testing.T) {
	Convey("RequestLogger should pass through without wrapping when not in dev mode", t, func() {
		originalRunMode := beego.BConfig.RunMode
		defer func() { beego.BConfig.RunMode = originalRunMode }()
		beego.BConfig.RunMode = "prod"

		ctx, res := newRequestLoggerTestContext(t)
		nextCalled := false
		sawWrappedWriter := false

		wrapped := RequestLogger(func(ctx *context.Context) {
			nextCalled = true
			_, sawWrappedWriter = ctx.ResponseWriter.ResponseWriter.(*responseWriter)
			ctx.ResponseWriter.WriteHeader(http.StatusAccepted)
		})

		wrapped(ctx)

		So(nextCalled, ShouldBeTrue)
		So(sawWrappedWriter, ShouldBeFalse)
		So(res.Code, ShouldEqual, http.StatusAccepted)
	})
}

func TestRequestLoggerDevMode(t *testing.T) {
	Convey("RequestLogger should wrap writer in dev mode and keep status code", t, func() {
		originalRunMode := beego.BConfig.RunMode
		defer func() { beego.BConfig.RunMode = originalRunMode }()
		beego.BConfig.RunMode = "dev"

		ctx, res := newRequestLoggerTestContext(t)
		nextCalled := false
		sawWrappedWriter := false
		capturedStatus := 0

		wrapped := RequestLogger(func(ctx *context.Context) {
			nextCalled = true
			if rw, ok := ctx.ResponseWriter.ResponseWriter.(*responseWriter); ok {
				sawWrappedWriter = true
				capturedStatus = rw.statusCode
			}

			ctx.ResponseWriter.WriteHeader(http.StatusCreated)

			if rw, ok := ctx.ResponseWriter.ResponseWriter.(*responseWriter); ok {
				capturedStatus = rw.statusCode
			}
		})

		wrapped(ctx)

		So(nextCalled, ShouldBeTrue)
		So(sawWrappedWriter, ShouldBeTrue)
		So(capturedStatus, ShouldEqual, http.StatusCreated)
		So(res.Code, ShouldEqual, http.StatusCreated)
	})
}
