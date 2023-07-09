package mini_gin

import (
	"fmt"
	"net/http"
	"testing"
)

func TestBindFORM(t *testing.T) {
	app := New()
	app.Use(LoggerMW)
	app.POST("/a/b/c", func(ctx *Context) {
		type Person struct {
			Name string	`form:"name"`
			Age  int `form:"age,default=10"`
		}
		var p *Person
		err := ctx.BindFORM(&p)
		if err != nil {
			ctx.WriteHeaderAndStatus(http.StatusBadRequest)
			_, _ = ctx.Write([]byte("body of request should be form pattern"))
		} else {
			ctx.Write([]byte(fmt.Sprintf("%+v", p)))
		}
	})
	app.Run()
}

func TestJSON(t *testing.T) {
	app := New()
	app.Use(LoggerMW)
	app.POST("/a/:name/c", func(ctx *Context) {
		type Person struct {
			Name string	`form:"name"`
			Age  int `form:"age,default=10"`
		}
		var p *Person
		err := ctx.BindFORM(&p)
		if err != nil {
			ctx.WriteHeaderAndStatus(http.StatusBadRequest)
			_, _ = ctx.Write([]byte("body of request should be form pattern"))
		} else {
			type resp struct {
				Code int
				Msg string
				Name string
			}
			ctx.JSON(http.StatusOK, &resp{Code: 0, Msg: "success", Name: ctx.Param("name")})
		}
	})
	app.Run()
}