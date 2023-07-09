package mini_gin

import (
	"fmt"
	"github.com/WANGgbin/mini_gin/bind"
	"github.com/WANGgbin/mini_gin/render"
	"github.com/WANGgbin/mini_gin/util"
	"io/ioutil"
	"net/http"
	"reflect"
)

type Context struct {
	indexOfHandlerChain int
	handlers            []MiddleWare
	params              map[string]string

	w   http.ResponseWriter
	req *http.Request

	status  int
	written bool
	e       *Engine
}

func newContext() interface{} {
	return &Context{}
}

// Next 经典的洋葱模型的实现
func (ctx *Context) Next() {
	for ; ctx.indexOfHandlerChain >= 0 && ctx.indexOfHandlerChain < len(ctx.handlers); {
		handler := ctx.handlers[ctx.indexOfHandlerChain]
		ctx.indexOfHandlerChain++
		handler(ctx)
	}
}

const abortIndex int = -1

func (ctx *Context) IsAborted() bool {
	return ctx.indexOfHandlerChain == abortIndex
}

func (ctx *Context) Abort() {
	ctx.indexOfHandlerChain = abortIndex
}

func (ctx *Context) reset() {
	ctx.indexOfHandlerChain = 0
	ctx.handlers = nil
	ctx.params = nil
	ctx.req = nil
	ctx.w = nil
	ctx.written = false
}

func (ctx *Context) setHandlers(handlers []MiddleWare) *Context {
	ctx.handlers = handlers
	return ctx
}

// setHandlersOnRouteNotHit 路由未命中时，使用该方法处理 req
func (ctx *Context) setHandlersOnRouteNotHit(status int) {
	util.Assert(status == http.StatusMethodNotAllowed || status == http.StatusNotFound, "status should only be notFound or methodNotAllowed, but got %d", status)

	if status == http.StatusMethodNotAllowed {
		ctx.setHandlers(append(ctx.e.noMethod, methodNotAllowedHandler))
	} else {
		ctx.setHandlers(append(ctx.e.noRoute, notFoundHandler))
	}
}

func (ctx *Context) setRespWriter(w http.ResponseWriter) *Context {
	ctx.w = w
	return ctx
}

func (ctx *Context) setRequest(req *http.Request) *Context {
	ctx.req = req
	return ctx
}

func (ctx *Context) setEngine(e *Engine) *Context {
	if ctx.e == nil {
		ctx.e = e
	}
	return ctx
}

// Header 获取 req 的 header
func (ctx *Context) Header(key string) string {
	return ctx.req.Header.Get(key)
}

// GetRawData 获取 req.body 全部内容
func (ctx *Context) GetRawData() ([]byte, error) {
	return ioutil.ReadAll(ctx.req.Body)
}

// SetHeader 设置 resp 的 header
func (ctx *Context) SetHeader(key, value string) {
	ctx.w.Header().Set(key, value)
}

func (ctx *Context) WriteHeaderAndStatus(status int) {
	if ctx.written {
		return
	}
	ctx.w.WriteHeader(status)
	ctx.written = true
	ctx.status = status
}

func (ctx *Context) Write(body []byte) (int, error) {
	ctx.WriteHeaderAndStatus(http.StatusOK)
	return ctx.w.Write(body)
}

func (ctx *Context) Written() bool {
	return ctx.written
}

/*
	USED FOR BINDING REQUEST
*/

func (ctx *Context) BindJSON(target interface{}) error {
	return ctx.bind(target, bind.JSON)
}

func (ctx *Context) BindFORM(target interface{}) error {
	return ctx.bind(target, bind.FORM)
}

func (ctx *Context) bind(target interface{}, binder bind.Binder) error {
	util.Assert(reflect.TypeOf(target).Kind() == reflect.Ptr, "target must be a pointer")

	return binder.Bind(ctx.req, target)
}

/*
	USED FOR RENDERING RESPONSE
*/

func (ctx *Context) JSON(status int, result interface{}) error {
	return ctx.render(status, result, render.JSON)
}

func (ctx *Context) render(status int, result interface{}, render render.Render) error {
	ctx.SetHeader("content-type", render.ContentType())
	ctx.WriteHeaderAndStatus(status)
	resultByte, err := render.Render(result)
	if err != nil {
		return err
	}

	n, err := ctx.Write(resultByte)
	if err != nil {
		return err
	}

	if n != len(resultByte) {
		return fmt.Errorf("write less")
	}

	return nil
}

// Param 获取路由的动态参数
func (ctx *Context) Param(key string) string {
	return ctx.params[key]
}

func (ctx *Context) setParams(params map[string]string) *Context {
	ctx.params = params
	return ctx
}
