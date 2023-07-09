package mini_gin

import (
	"fmt"
	"github.com/WANGgbin/mini_gin/util"
	"strings"
)

/*
路由规则：
1. 每个 segment 中，相同前缀的通配符只能有一个
	eg:
	/a/prefix:key1 /a/prefix:key2 conflict
	/a/prefix:key1 /a/prefix1:key2 right
	/a/prefix:key1 /a/:key	right

2. 。。。
*/

func newTrieTree() *trieTree {
	return &trieTree{
		root: &node{
			content: "/",
		},
	}
}

type trieTree struct {
	root *node
}

func (tree *trieTree) insert(route string, handlers ...MiddleWare) {
	util.Assert(len(handlers) > 0, "handlers should not be empty")

	curNode := tree.root
	curIndex := 0

	for {
		lenOfPrefix := curNode.getLenOfPrefix(route[curIndex:], route)
		// 当前节点是 route 前缀，继续判断子孩子
		curIndex += lenOfPrefix
		if lenOfPrefix == len(curNode.content) {
			// route 匹配完毕，如果当前节点是一个有效路由，则路由冲突，直接 panic
			if curIndex == len(route) {
				if curNode.isRoute() {
					// 同一条路由重复插入
					panic(fmt.Sprintf("route %s has been registered", route))
				}
				// 否则，标记当前节点为有效路由
				curNode.setRoute(handlers)
				return
			}
			// route 未匹配完毕，寻找子孩子节点
			childNode := curNode.findNextNode(route[curIndex:])
			if childNode != nil {
				// 找到孩子节点，继续判断
				curNode = childNode
				continue
			}
			// 未找到，插入新节点
			curNode.addChild(newNode(route[curIndex:], handlers, route))
			return
		}

		// 分裂当前节点
		curNode.split(lenOfPrefix)
		if curIndex < len(route) {
			curNode.addChild(newNode(route[curIndex:], handlers, route))
		} else {
			curNode.setRoute(handlers)
		}
		return
	}
}

type pathInfo struct {
	handlers []MiddleWare      // url handlers
	params   map[string]string // url 参数
}

// getRouteInfo 获取与 route 对应的 handlers & params
func (tree *trieTree) getRouteInfo(route string) *pathInfo {
	return tree.root.getRouteInfo(route)
}

type node struct {
	handlers []MiddleWare

	// 动态参数的索引，用于记录当前节点是否有动态参数，目前仅支持通配符 ':'
	// 例子：
	// /:key1/a:key2
	// dynKeys = [][2]int{{1, 5}, {8, 12}}
	// 无论在插入还是查找的时候，通过该参数可以很方便的进行判断是否冲突以及是否匹配
	dynKeys [][2]int
	content string
	// 用于打印报错信息
	fullPath string

	parent   *node
	children []*node
}

func newNode(route string, handlers []MiddleWare, fullPath string) *node {
	n := &node{
		handlers: handlers,
		content:  route,
		fullPath: fullPath,
	}
	n.parseDynKeyFromRoute(route)
	return n
}

// parseDynKeyFromRoute 从 route 解析出所有的动态参数的起始/结束索引
func (n *node) parseDynKeyFromRoute(route string) {
	isInDynKey := false
	start := 0
	n.dynKeys = nil

	for index, char := range route {
		if char == ':' {
			isInDynKey = true
			start = index
		} else if char == '/' && isInDynKey {
			n.dynKeys = append(n.dynKeys, [2]int{start, index - 1})
			start = 0
			isInDynKey = false
		}
	}

	if isInDynKey {
		n.dynKeys = append(n.dynKeys, [2]int{start, len(route) - 1})
	}
}

// getLenOfPrefix 获取当前节点 content 跟 route 的最大前缀长度
// 因为路由的注册通常在服务启动阶段，所以，如果 path 冲突，则直接 panic
func (n *node) getLenOfPrefix(route string, fullPath string) int {
	curIndex := 0

	for ; curIndex < len(n.content) && curIndex < len(route); curIndex++ {
		if n.content[curIndex] != route[curIndex] {
			break
		}

		// 每个节点中如果存在通配符，则一定存储 key 的完整格式: ':key'
		// 如果存储不完整的格式，意味着出现了多个相同前缀的 key，这显然是不对的。
		if n.content[curIndex] == ':' {
			oldRouteKey := getSubStrBeforeFirstSlash(n.content[curIndex+1:])
			newRouteKey := getSubStrBeforeFirstSlash(route[curIndex+1:])

			// key 不相同直接 panic
			if oldRouteKey != newRouteKey {
				panic(
					fmt.Sprintf("key: %s in new path: %s is conflict with existing key %s in existing path: %s",
						route[curIndex:curIndex+len(newRouteKey)+1],
						fullPath,
						n.content[curIndex:curIndex+len(oldRouteKey)+1],
						n.fullPath,
					),
				)
			}
			curIndex += len(oldRouteKey)
		}
	}

	return curIndex
}

func getSubStrBeforeFirstSlash(str string) string {
	index := strings.Index(str, "/")
	if index == -1 {
		return str
	}
	return str[:index]
}

func (n *node) isRoute() bool {
	// 如果当前节点有 handlers，则为一个有效的路由
	return len(n.handlers) > 0
}

func (n *node) setRoute(handlers []MiddleWare) {
	n.handlers = handlers
}

// findNextNode 注册路由的时候，寻找下一个匹配的节点
func (n *node) findNextNode(route string) *node {
	for _, child := range n.children {
		if child.content[0] == route[0] {
			return child
		}
	}
	return nil
}

func (n *node) addChild(child *node) {
	n.children = append(n.children, child)
	child.parent = n
}

// split 从 n 拆分出子节点, 子节点路径为 n.content[curIndex:]
func (n *node) split(curIndex int) {
	child := &node{
		handlers: n.handlers,
		content:  n.content[curIndex:],
		parent:   n,
		children: n.children,
	}

	n.handlers = nil
	n.children = []*node{child}
	n.content = n.content[:curIndex]

	if len(n.dynKeys) > 0 {
		child.parseDynKeyFromRoute(child.content)
		n.parseDynKeyFromRoute(n.content)
	}

	return
}

// getRouteInfo 获取与 route 匹配的路由信息，未找到返回 nil
func (n *node) getRouteInfo(route string) *pathInfo {
	nextIndex, params := n.match(route)
	if nextIndex == -1 {
		return nil
	}

	if nextIndex == len(route) {
		if n.isRoute() {
			return &pathInfo{params: params, handlers: n.handlers}
		}
		return nil
	}

	candidateNodes := n.findCandidateNodes(route[nextIndex:])

	for _, candidate := range candidateNodes {
		info := candidate.getRouteInfo(route[nextIndex:])
		if info != nil {
			util.MergeParam(&params, info.params)
			return &pathInfo{handlers: info.handlers, params: params}
		}
	}
	return nil
}

// match 判断 route 是否跟节点 n 匹配，如果匹配，则返回 route 新的索引，用于后续判断.
// 同时 返回动态参数。如果不匹配，则返回索引为 -1.
func (n *node) match(route string) (int, map[string]string) {
	// 如果没有动态参数，直接匹配
	if len(n.dynKeys) == 0 {
		if strings.HasPrefix(route, n.content) {
			return len(n.content), nil
		}
		return -1, nil
	}

	var params map[string]string
	curIndex := 0

	if n.dynKeys[0][0] > 0 {
		if route[curIndex:n.dynKeys[0][0]] != n.content[:n.dynKeys[0][0]] {
			return -1, nil
		}
		curIndex += n.dynKeys[0][0]
	}

	for idx, dynKey := range n.dynKeys {
		var subStr string
		if idx < len(n.dynKeys)-1 {
			subStr = n.content[dynKey[1]+1 : n.dynKeys[idx+1][0]]
		} else {
			subStr = n.content[dynKey[1]+1:]
		}

		var value string
		if subStr == "" {
			nextSlashIndex := strings.Index(route[curIndex:], "/")
			if nextSlashIndex == -1 {
				value = route[curIndex:]
				curIndex = len(route)
			} else {
				value = route[curIndex : curIndex+nextSlashIndex]
				curIndex += nextSlashIndex
			}
		} else {
			subIndex := strings.Index(route[curIndex:], subStr)
			if subIndex == -1 {
				return -1, nil
			}
			value = route[curIndex : curIndex+subIndex]
			curIndex += subIndex + len(subStr)
		}

		if params == nil {
			// 延迟创建
			params = make(map[string]string)
		}
		key := n.content[dynKey[0]+1 : dynKey[1]+1]
		params[key] = value
	}

	return curIndex, params
}

// findCandidateNodes 匹配路由的时候，寻找符合要求的孩子节点。可能存在两个匹配的孩子节点。
// 需要特别注意优先级：通配符的孩子节点优先级最低。
func (n *node) findCandidateNodes(route string) []*node {
	var nodeBeginWithWildCard *node
	var candidateNodes []*node
	for _, child := range n.children {
		if child.content[0] == route[0] {
			candidateNodes = append(candidateNodes, child)
		} else {
			if child.content[0] == ':' {
				nodeBeginWithWildCard = child
			}
		}
	}

	if nodeBeginWithWildCard != nil {
		candidateNodes = append(candidateNodes, nodeBeginWithWildCard)
	}

	// 最多存在两个节点
	util.Assert(len(candidateNodes) <= 2, fmt.Sprintf("candidateNodes must be less than 2"))

	return candidateNodes
}
