package mini_gin

type Context struct {
	indexOfHandlerChain int
	handlers []MiddleWare
}

func newContext() interface{} {
	return &Context{}
}

// Next 经典的洋葱模型的实现
func (ctx *Context) Next() {
	for ; ctx.indexOfHandlerChain >= 0 && ctx.indexOfHandlerChain  < len(ctx.handlers); ctx.indexOfHandlerChain++ {
		ctx.handlers[ctx.indexOfHandlerChain](ctx)
	}
}

func (ctx *Context) Abort() {
	ctx.indexOfHandlerChain = -1
}

