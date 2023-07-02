package mini_gin

import (
	"fmt"
	"github.com/smartystreets/goconvey/convey"
	"strings"
	"testing"
)

func Test_trieTree_insert(t *testing.T) {
	convey.Convey("", t, func() {
		fakeHandler := func(ctx *Context) {
			return
		}

		testCases := []struct {
			name              string
			pathHandlersPairs []struct {
				path     string
				handlers []MiddleWare
			}
			shouldPanic bool
		}{
			{
				name: "multi key same segment",
				pathHandlersPairs: []struct {
					path     string
					handlers []MiddleWare
				}{
					{
						path:     "/a/prefix:key/b",
						handlers: []MiddleWare{fakeHandler},
					},
					{
						path:     "/a/prefix:key1/b",
						handlers: []MiddleWare{fakeHandler},
					},
				},
				shouldPanic: true,
			},
			{
				name: "/a/:key & /a/:key1/b",
				pathHandlersPairs: []struct {
					path     string
					handlers []MiddleWare
				}{
					{
						path:     "/a/:key",
						handlers: []MiddleWare{fakeHandler},
					},
					{
						path:     "/a/:key1/b",
						handlers: []MiddleWare{fakeHandler},
					},
				},
				shouldPanic: true,
			},
			{
				name: "insert same route again",
				pathHandlersPairs: []struct {
					path     string
					handlers []MiddleWare
				}{
					{
						path:     "/a/prefix:key/b",
						handlers: []MiddleWare{fakeHandler},
					},
					{
						path:     "/a/prefix:key/b",
						handlers: []MiddleWare{fakeHandler},
					},
				},
				shouldPanic: true,
			},
			{
				name: "regular",
				pathHandlersPairs: []struct {
					path     string
					handlers []MiddleWare
				}{
					{
						path:     "/a/b/c",
						handlers: []MiddleWare{fakeHandler},
					},
					{
						path:     "/a/b1/c",
						handlers: []MiddleWare{fakeHandler},
					},
					{
						path:     "/a/:key/c",
						handlers: []MiddleWare{fakeHandler},
					},
					{
						path:     "/a",
						handlers: []MiddleWare{fakeHandler},
					},
					{
						path:     "/a/prefix:usr/age",
						handlers: []MiddleWare{fakeHandler},
					},
					{
						path:     "/a/prefix1:usr/age",
						handlers: []MiddleWare{fakeHandler},
					},
					{
						path:     "/a/prefix1:usr/name",
						handlers: []MiddleWare{fakeHandler},
					},
				},
			},
		}

		for _, testCase := range testCases {
			tree := newTrieTree()
			if testCase.shouldPanic {
				convey.So(func() {
					for _, pathHandlerPair := range testCase.pathHandlersPairs {
						tree.insert(pathHandlerPair.path, pathHandlerPair.handlers...)
					}
				}, convey.ShouldPanic)
			} else {
				for _, pathHandlerPair := range testCase.pathHandlersPairs {
					tree.insert(pathHandlerPair.path, pathHandlerPair.handlers...)
				}
				printTrieTree(tree)
			}
		}
	})

}

// printTrieTree 打印 tree 信息
func printTrieTree(tree *trieTree) {
	printTraceback(tree.root, 0)
}

func printTraceback(n *node, level int) {
	printNode(n, level)
	for _, child := range n.children {
		printTraceback(child, level+1)
	}
}

func printNode(n *node, level int) {
	var builder strings.Builder
	builder.WriteString(strings.Repeat("\t", level))
	builder.WriteString(fmt.Sprintf("-> %s", n.content))
	if len(n.handlers) > 0 {
		builder.WriteString(fmt.Sprintf(" | handlers: %v\n", n.handlers))
	} else {
		builder.WriteByte('\n')
	}
	fmt.Printf(builder.String())
}

func Test_node_parseDynKeyFromRoute(t *testing.T) {
	convey.Convey("", t, func() {
		testCases := []struct {
			name        string
			route       string
			wantDynKeys [][2]int
		}{
			{
				name:  "no dyn key",
				route: "/a/b/c",
			},
			{
				name:  "one dyn key",
				route: "/a/:key/c",
				wantDynKeys: [][2]int{
					{3, 6},
				},
			},
			{
				name:  "multi dyn key",
				route: "/a/:key1/c/:key2",
				wantDynKeys: [][2]int{
					{3, 7},
					{11, 15},
				},
			},
			{
				route: ":key1/:key2",
				wantDynKeys: [][2]int{
					{0, 4},
					{6, 10},
				},
			},
			{
				route: ":key1/",
				wantDynKeys: [][2]int{
					{0, 4},
				},
			},
		}

		for _, testCase := range testCases {
			convey.Convey(testCase.route, func() {
				n := new(node)
				n.parseDynKeyFromRoute(testCase.route)
				convey.So(n.dynKeys, convey.ShouldResemble, testCase.wantDynKeys)
			})
		}
	})
}

func Test_node_getLenOfPrefix(t *testing.T) {
	convey.Convey("", t, func() {
		testCases := []struct {
			route       string
			node        *node
			length      int
			shouldPanic bool
		}{
			{
				route: "a",
				node: &node{
					content: "b",
				},
				length: 0,
			},
			{
				route: "a/b/c",
				node: &node{
					content: "a/b",
				},
				length: 3,
			},
			{
				route: "a/b",
				node: &node{
					content: "a/b/",
				},
				length: 3,
			},
			{
				route: "a/b/c/d",
				node: &node{
					content: "a/b/",
				},
				length: 4,
			},
			{
				route: "a/:b/c",
				node: &node{
					content: "a/:b/c",
				},
				length: 6,
			},
			{
				route: "a/:b1/c",
				node: &node{
					content: "a/:b/c",
				},
				shouldPanic: true,
			},
			{
				route: "a/:d/c",
				node: &node{
					content: "a/:b/c",
				},
				shouldPanic: true,
			},
		}

		for _, testCase := range testCases {
			convey.Convey(testCase.route, func() {
				// 注意 convey.ShouldPanic 的这种用法
				if testCase.shouldPanic {
					convey.So(func() {
						testCase.node.getLenOfPrefix(testCase.route, "")
					}, convey.ShouldPanic)
				} else {
					convey.So(testCase.node.getLenOfPrefix(testCase.route, ""), convey.ShouldEqual, testCase.length)
				}
			})
		}
	})
}

func Test_node_split(t *testing.T) {
	convey.Convey("", t, func() {
		testCases := []struct {
			name          string
			node          *node
			copyNode      *node
			curIndex      int
			parentDynKeys [][2]int
			childDynKeys  [][2]int
		}{
			{
				name: "without dyn key",
				node: &node{
					handlers: []MiddleWare{func(ctx *Context) {}},
					content:  "a/b/c",
					children: []*node{new(node), new(node)},
				},
				copyNode: &node{
					handlers: []MiddleWare{func(ctx *Context) {}},
					content:  "a/b/c",
					children: []*node{new(node), new(node)},
				},
				curIndex: 3,
			},
			{
				name: "dyn key",
				node: &node{
					handlers: []MiddleWare{func(ctx *Context) {}},
					content:  "a/:key1/c/:key2",
					children: []*node{new(node), new(node)},
					dynKeys: [][2]int{
						{2, 6},
						{10, 14},
					},
				},
				copyNode: &node{
					handlers: []MiddleWare{func(ctx *Context) {}},
					content:  "a/:key1/c/:key2",
					children: []*node{new(node), new(node)},
					dynKeys: [][2]int{
						{2, 6},
						{10, 14},
					},
				},
				curIndex: 8,
				parentDynKeys: [][2]int{
					{2, 6},
				},
				childDynKeys: [][2]int{
					{2, 6},
				},
			},
		}

		for _, testCase := range testCases {
			convey.Convey(testCase.name, func() {
				testCase.node.split(testCase.curIndex)
				convey.So(len(testCase.node.children), convey.ShouldEqual, 1)
				convey.So(testCase.node.handlers, convey.ShouldBeNil)
				convey.So(testCase.node.content, convey.ShouldEqual, testCase.copyNode.content[:testCase.curIndex])
				convey.So(testCase.node.dynKeys, convey.ShouldResemble, testCase.parentDynKeys)

				// 校验子节点
				convey.So(len(testCase.node.children[0].children), convey.ShouldEqual, len(testCase.copyNode.children))
				convey.So(len(testCase.node.children[0].handlers), convey.ShouldEqual, len(testCase.copyNode.handlers))
				convey.So(testCase.node.children[0].content, convey.ShouldEqual, testCase.copyNode.content[testCase.curIndex:])
				convey.So(testCase.node.children[0].parent, convey.ShouldEqual, testCase.node)
				convey.So(testCase.node.children[0].dynKeys, convey.ShouldResemble, testCase.childDynKeys)
			})
		}
	})
}

func Test_trieTree_getPathInfo(t *testing.T) {
	convey.Convey("", t, func(){
		fakeHandler := func(ctx *Context){}
		testCases := []struct{
			name string
			routeHandlerPairs []struct{
				route string
				handlers []MiddleWare
			}
			wantResult map[string]*pathInfo
		}{
			{
				name: "",
				routeHandlerPairs: []struct{
					route string
					handlers []MiddleWare
				}{
					{
						route: "/a/b/c",
						handlers: []MiddleWare{fakeHandler},
					},
					{
						route: "/a/prefix:key",
						handlers: []MiddleWare{fakeHandler},
					},
					{
						route: "/a/:key1/prefix:key2",
						handlers: []MiddleWare{fakeHandler},
					},
					{
						route: "/a/:key1",
						handlers: []MiddleWare{fakeHandler},
					},
				},
				wantResult: map[string]*pathInfo{
					"/a/b/c": {handlers: []MiddleWare{fakeHandler}},
					"/a/prefix": {handlers: []MiddleWare{fakeHandler}, params: map[string]string{"key": ""}},
					"/a/val1/prefixval2": {handlers: []MiddleWare{fakeHandler}, params: map[string]string{"key1": "val1", "key2": "val2"}},
					"/a/val1/prefix": {handlers: []MiddleWare{fakeHandler}, params: map[string]string{"key1": "val1", "key2": ""}},
					"/a/val1": {handlers: []MiddleWare{fakeHandler}, params: map[string]string{"key1": "val1"}},
					"/a/preifx": {handlers: []MiddleWare{fakeHandler}, params: map[string]string{"key1": "preifx"}},

					"/not/exist": nil,
				},
			},
		}

		for _, testCase := range testCases {
			convey.Convey(testCase.name, func(){
				tree := newTrieTree()
				for _, routeHandlerPair := range testCase.routeHandlerPairs {
					tree.insert(routeHandlerPair.route, routeHandlerPair.handlers...)
				}
				printTrieTree(tree)
				for route, info := range testCase.wantResult {
					gotPathInfo := tree.getPathInfo(route)
					if info == nil {
						convey.So(gotPathInfo, convey.ShouldBeNil)
					} else {
						convey.So(len(gotPathInfo.handlers), convey.ShouldEqual, len(info.handlers))
						convey.So(gotPathInfo.params, convey.ShouldResemble, info.params)
					}
				}
			})
		}
	})
}
