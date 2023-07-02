package mini_gin

import "testing"

func TestNew(t *testing.T) {
	app := New()

	mw1 := func(ctx *Context) {
		return
	}
	app.Use(mw1)

	app.Run()
}

func TestRegisterRoute(t *testing.T) {
	app := New()
	mw1 := func(ctx *Context) {
		return
	}
	// two ways
	app.GET("route", mw1)
	app.POST("route", mw1)

	gp := app.NewGroup("prefix", mw1, mw1)
	gp.GET("route", mw1)
	gp.POST("route", mw1)
}