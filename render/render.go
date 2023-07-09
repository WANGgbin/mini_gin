package render

type Render interface {
	Render(result interface{}) ([]byte, error)
	ContentType() string
}


var (
	JSON Render = (*jsonRender)(nil)
)