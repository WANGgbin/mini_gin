package mini_gin_test

import (
	"github.com/WANGgbin/mini_gin"
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	app := mini_gin.NewWithCfg(mini_gin.WithHandleMethodNotAllowed())
	//logFile, _ := os.OpenFile("log.txt", os.O_CREATE|os.O_RDWR, 0777)
	//pattern := func(pattern *mini_gin.LoggerParam) string {
	//	return fmt.Sprintf(
	//		"[%s] method: %s, route: %s, latency: %d ms, status: %d\n",
	//		pattern.TimeStamp.Format(time.RFC3339),
	//		pattern.Method,
	//		pattern.Route,
	//		pattern.Latency.Milliseconds(),
	//		pattern.StatusCode,
	//	)
	//}
	//app.Use(mini_gin.LoggerMWWithCfg(mini_gin.LoggerWithPattern(pattern), mini_gin.LoggerWithDest(logFile)))
	app.Use(mini_gin.LoggerMW)
	app.Use(mini_gin.RecoverMW)
	app.GET("/a/b/c", func(ctx *mini_gin.Context) {
		//ctx.w.Write([]byte("pong"))
		panic("info of panic")
	})
	app.Run()
}

func TestNoRoute(t *testing.T) {
	app := mini_gin.New()

	noRouteHandler := func(ctx *mini_gin.Context) {
		ctx.WriteHeaderAndStatus(http.StatusBadRequest)
	}

	app.NoRoute(noRouteHandler)

	app.Run()
}

func TestRegisterRoute(t *testing.T) {
	app := mini_gin.New()
	mw1 := func(ctx *mini_gin.Context) {
		return
	}
	// two ways
	app.GET("route", mw1)
	app.POST("route", mw1)

	gp := app.NewGroup("prefix", mw1, mw1)
	gp.GET("route", mw1)
	gp.POST("route", mw1)
}
