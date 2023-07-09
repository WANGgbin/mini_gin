package mini_gin

import (
	"context"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func New() *Engine {
	return NewWithCfg()
}

func NewWithCfg(opts ...EngineOption) *Engine {
	options := NewEngineOptions(opts...)
	engine := &Engine{
		server: &http.Server{
			Addr:              options.Addr,
			ReadTimeout:       options.ReadTimeout,
			ReadHeaderTimeout: options.ReadHeaderTimeout,
			WriteTimeout:      options.WriteTimeout,
			IdleTimeout:       options.IdlTimeout,
		},
		rootRouteGroup: &RouteGroup{
			basePrefix: "/",
		},
		method2routes: map[string]*trieTree{
			http.MethodGet:    newTrieTree(),
			http.MethodPost:   newTrieTree(),
			http.MethodPut:    newTrieTree(),
			http.MethodDelete: newTrieTree(),
		},
		ctxPool: sync.Pool{
			New: newContext,
		},
		HandleMethodNotAllowed: options.HandleMethodNotAllowed,
	}

	engine.rootRouteGroup.engine = engine

	return engine
}

type Engine struct {
	server         *http.Server
	rootRouteGroup *RouteGroup
	method2routes  map[string]*trieTree

	// 对于每条链接都会使用到的结构体类型，使用池化技术减少内存的分配次数，进而提高系统性能
	ctxPool sync.Pool

	noRoute  []MiddleWare
	noMethod []MiddleWare

	// 设置为 true，当某个未匹配的路由的另一种方法存在时，返回 Method not allowed
	HandleMethodNotAllowed bool
}

func (e *Engine) Use(mws ...MiddleWare) {
	e.rootRouteGroup.Append(mws...)
	e.combineNoRouteHandlers()
	e.combineNoMethodHandlers()
}

// NoRoute 用户自定义 路由未命中时的 处理逻辑
func (e *Engine) NoRoute(mws ...MiddleWare) {
	e.noRoute = mws
	e.combineNoRouteHandlers()
}

// NoMethod 用户自定义 方法不支持时的 处理逻辑
func (e *Engine) NoMethod(mws ...MiddleWare) {
	e.noMethod = mws
	e.combineNoMethodHandlers()
}

func (e *Engine) combineNoRouteHandlers() {
	e.noRoute = e.rootRouteGroup.getHandlers(e.noRoute...)
}

func (e *Engine) combineNoMethodHandlers() {
	e.noMethod = e.rootRouteGroup.getHandlers(e.noMethod...)
}

// ServeHTTP 实现 http.Handler
func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	routeInfo := e.getRouteInfo(req.Method, req.URL.Path)
	ctx := e.ctxPool.Get().(*Context)
	ctx.setEngine(e).setRespWriter(w).setRequest(req)

	if routeInfo == nil {
		var hasSetHandlers bool
		if e.HandleMethodNotAllowed {
			for method := range e.method2routes {
				if method == req.Method {
					continue
				}
				routeInfo = e.getRouteInfo(method, req.URL.Path)
				if routeInfo != nil {
					hasSetHandlers = true
					ctx.setHandlersOnRouteNotHit(http.StatusMethodNotAllowed)
				}
			}
		}
		if !hasSetHandlers {
			ctx.setHandlersOnRouteNotHit(http.StatusNotFound)
		}
	} else {
		ctx.setHandlers(routeInfo.handlers)
		ctx.setParams(routeInfo.params)
	}
	ctx.Next()
	ctx.reset()
	e.ctxPool.Put(ctx)
}

// Run 基于 net/http 实现
func (e *Engine) Run() {
	e.server.Handler = e

	// 服务器异常退出
	errCh := make(chan error, 1)
	go func() {
		if err := e.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()
	sigCh := make(chan os.Signal)
	// SIGQUIT: 优雅退出
	// SIGTERM: 强制退出
	signal.Notify(sigCh, syscall.SIGQUIT, syscall.SIGTERM)
	select {
	case err := <-errCh:
		log.Errorf("Server exit unexpectedly, error: %v", err)
	case sig := <-sigCh:
		if sig == syscall.SIGTERM {
			if err := e.server.Close(); err != nil {
				log.Errorf("Close server error: %v", err)
			} else {
				log.Infof("Close server")
			}
		} else {
			// TODO(@wangguobin): 设置优雅退出最大等待时间
			if err := e.server.Shutdown(context.Background()); err != nil {
				log.Errorf("Shutdown server error: %v", err)
			} else {
				log.Infof("Shutdown server")
			}
		}
	}
}

/*
	Register Routes
*/

func (e *Engine) GET(route string, handler MiddleWare) {
	e.rootRouteGroup.GET(route, handler)
}

func (e *Engine) POST(route string, handler MiddleWare) {
	e.rootRouteGroup.POST(route, handler)
}

func (e *Engine) PUT(route string, handler MiddleWare) {
	e.rootRouteGroup.PUT(route, handler)
}

func (e *Engine) DELETE(route string, handler MiddleWare) {
	e.rootRouteGroup.DELETE(route, handler)
}

func (e *Engine) NewGroup(baseRoute string, handlers ...MiddleWare) *RouteGroup {
	return newRouteGroup(e, baseRoute, handlers...)
}

func (e *Engine) getRouteInfo(method, route string) *pathInfo {
	tree := e.method2routes[method]
	if tree == nil {
		return nil
	}

	return tree.getRouteInfo(route)
}
