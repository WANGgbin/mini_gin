package mini_gin

import (
	"fmt"
	"net/http"
	"path"
	"strings"
)

// RouteGroup 表示一个路由组
type RouteGroup struct {
	engine *Engine

	basePrefix   string
	baseHandlers []MiddleWare
}

func newRouteGroup(engine *Engine, basePrefix string, baseHandlers ...MiddleWare) *RouteGroup {
	return &RouteGroup{
		engine:       engine,
		basePrefix:   engine.rootRouteGroup.getAbsRoute(basePrefix),
		baseHandlers: engine.rootRouteGroup.getHandlers(baseHandlers...),
	}
}

func (rg *RouteGroup) Append(mws ...MiddleWare) {
	rg.baseHandlers = append(rg.baseHandlers, mws...)
}

func (rg *RouteGroup) GET(route string, handler MiddleWare) {
	rg.register(http.MethodGet, route, handler)
}

func (rg *RouteGroup) POST(route string, handler MiddleWare) {
	rg.register(http.MethodPost, route, handler)
}

func (rg *RouteGroup) PUT(route string, handler MiddleWare) {
	rg.register(http.MethodPut, route, handler)
}

func (rg *RouteGroup) DELETE(route string, handler MiddleWare) {
	rg.register(http.MethodDelete, route, handler)
}

func (rg *RouteGroup) register(method, route string, handler MiddleWare) {
	tree := rg.engine.method2routes[method]
	if tree == nil {
		panic(fmt.Sprintf("route tree of method %s is nil", method))
	}

	tree.insert(rg.getAbsRoute(route), rg.getHandlers(handler)...)
}

func (rg *RouteGroup) getAbsRoute(relativeRoute string) string {
	return path.Join(rg.basePrefix, relativeRoute)
}

func (rg *RouteGroup) getBaseHandlers() []MiddleWare {
	return rg.baseHandlers
}

func (rg *RouteGroup) getHandlers(deltas ...MiddleWare) []MiddleWare {
	length := len(rg.baseHandlers) + len(deltas)
	if length == 0 {
		return nil
	}

	handlers := make([]MiddleWare, length)
	copy(handlers, rg.baseHandlers)
	copy(handlers[len(rg.baseHandlers):], deltas)

	return handlers
}

// validateRoute 校验 route 是否合法
// 1. 必须以 '/' 开始
// 2. 如果包含动态参数，必须是 /[anything]:key[/] 格式
// 3. 两个 '/' 之间不能为空
func validateRoute(route string) bool {
	if route == "" || route[0] != '/' {
		return false
	}

	route = strings.TrimSuffix(route[1:], "/")
	segs := strings.Split(route, "/")
	for _, seg := range segs {
		if !validateSegment(seg) {
			return false
		}
	}
	return true
}

// validateSegment 校验每一个 segment 是否有效
// segment: xxx
// 1. 不能为空
// 2. 至多只能有一个 wildcard 且 wildcard 对应的 key 不能为空
func validateSegment(seg string) bool {
	if seg == "" {
		return false
	}

	firstIndex := strings.Index(seg, ":")
	if firstIndex != strings.LastIndex(seg, ":") {
		return false
	}

	if firstIndex == len(seg)-1 {
		return false
	}

	return true
}
