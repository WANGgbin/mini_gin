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
	return &Engine{
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
	}
}

type Engine struct {
	server         *http.Server
	rootRouteGroup *RouteGroup
	method2routes  map[string]*trieTree

	// 对于每条链接都会使用到的结构体类型，使用池化技术减少内存的分配次数，进而提高系统性能
	ctxPool	sync.Pool
}

func (e *Engine) Use(mws ...MiddleWare) {
	e.rootRouteGroup.Append(mws...)
}

// ServeHTTP 实现 http.Handler
func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	pathInfo := e.getRouteInfo(req.Method, req.URL.Path)
	// TODO(@wangguobin): 没有相关 handler，返回 404
	if pathInfo == nil {
		return
	}

	ctx := e.ctxPool.Get().(*Context)


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