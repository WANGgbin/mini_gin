package mini_gin

import (
	"fmt"
	"github.com/WANGgbin/mini_gin/util"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"time"
)

type MiddleWare func(ctx *Context)

// notFoundHandler 返回 404
func notFoundHandler(ctx *Context) {
	handleOnRouteNotHit(ctx, http.StatusNotFound)
}

func methodNotAllowedHandler(ctx *Context) {
	handleOnRouteNotHit(ctx, http.StatusMethodNotAllowed)
}

var (
	defaultBodyOnNotFound         = []byte("Not Found")
	defaultBodyOnMethodNotAllowed = []byte("Method Not Allowed")
)

func handleOnRouteNotHit(ctx *Context, status int) {
	if ctx.Written() {
		return
	}
	// 返回默认响应
	ctx.SetHeader("content-type", "text/plain")
	var body []byte
	if status == http.StatusMethodNotAllowed {
		body = defaultBodyOnMethodNotAllowed
	} else {
		body = defaultBodyOnNotFound
	}

	ctx.WriteHeaderAndStatus(status)
	_, err := ctx.Write(body)
	if err != nil {
		log.Errorf("write body error: %v", err)
	}
}

// RecoverMW used to recover
func RecoverMW(ctx *Context) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			// TODO(@wangguobin): 打印错误栈帧
			log.Errorf("%s %s, panic happend: %v", ctx.req.Method, ctx.req.URL.Path, panicInfo)
			ctx.Abort()
			ctx.WriteHeaderAndStatus(http.StatusInternalServerError)
			_, _ = ctx.Write([]byte("Internal Server error"))
		}
	}()
	ctx.Next()
}

type LoggerCfg struct {
	Dest    io.Writer
	Pattern func(ctx *LoggerParam) string
}

type LoggerParam struct {
	Method     string
	Route      string
	StatusCode int
	TimeStamp  time.Time
	Latency    time.Duration
}

type Option func(cfg *LoggerCfg)

func LoggerWithPattern(pattern func(ctx *LoggerParam) string) Option {
	return func(cfg *LoggerCfg) {
		cfg.Pattern = pattern
	}
}

func LoggerWithDest(dest io.Writer) Option {
	return func(cfg *LoggerCfg) {
		cfg.Dest = dest
	}
}

var loggerCfg = &LoggerCfg{
	Dest: os.Stdout,
	Pattern: func(param *LoggerParam) string {
		return fmt.Sprintf(
			"%s method: %s, route: %s latency: %d ms\n",
			time.Now().Format(time.RFC3339),
			param.Method,
			param.Route,
			param.Latency.Milliseconds(),
		)

	},
}

func LoggerMWWithCfg(options ...Option) func(ctx *Context) {
	for _, op := range options {
		op(loggerCfg)
	}

	return LoggerMW
}

// LoggerMW 日志中间件
func LoggerMW(ctx *Context) {
	start := time.Now()
	defer func() {
		param := LoggerParam{
			Method:     ctx.req.Method,
			Route:      ctx.req.URL.String(),
			StatusCode: ctx.status,
			Latency:    time.Now().Sub(start),
			TimeStamp:  time.Now(),
		}
		_, err := loggerCfg.Dest.Write(util.String2Byte(loggerCfg.Pattern(&param)))
		if err != nil {
			fmt.Printf("Logger error: %v\n", err)
		}
	}()
	ctx.Next()
}
