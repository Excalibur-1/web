package web_test

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/Excalibur-1/configuration"
	"github.com/Excalibur-1/web"
	"github.com/labstack/echo/v4"
	. "github.com/smartystreets/goconvey/convey"
)

func startServer() {
	web.App(func(eng *echo.Echo) {
		eng.GET("/none/api", func(ctx echo.Context) error {
			return ctx.String(http.StatusOK, "test app")
		})
	}, "myconf", "1000", configuration.MockEngine(map[string]string{
		"/myconf/base/server/1000": "{\"addr\":\":9999\"}",
	}))
}

func TestApp(t *testing.T) {
	go startServer()
	client := &http.Client{}
	Convey("test App\n", t, func() {
		req, err := http.NewRequest("GET", "http://localhost:9999/none/api", nil)
		So(err, ShouldBeNil)
		resp, err := client.Do(req)
		So(err, ShouldBeNil)
		rid := resp.Header.Get(echo.HeaderXRequestID)
		So(rid, ShouldNotBeNil)
		actual, err := ioutil.ReadAll(resp.Body)
		So(err, ShouldBeNil)
		So(string(actual), ShouldEqual, "test app")
		_ = resp.Body.Close()
	})
}
