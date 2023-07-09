package bind_test

import (
	"github.com/WANGgbin/mini_gin"
	"net/http"
	"testing"
)

func TestBindJson(t *testing.T) {
	app := mini_gin.New()

	app.GET("/a/b/c", func(ctx *mini_gin.Context) {
		type person struct {
			name string
			age int
		}
		var p *person
		err := ctx.BindJSON(&p)
		if err != nil {
			ctx.WriteHeaderAndStatus(http.StatusBadRequest)
			_, _ = ctx.Write([]byte("body of request should be json pattern"))
		} else {
			ctx.Write([]byte("success"))
		}
	})
	app.Run()
}
