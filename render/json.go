package render

import "encoding/json"

type jsonRender struct {}

func (j *jsonRender) Render(result interface{}) ([]byte, error) {
	return json.Marshal(result)
}

func (j *jsonRender) ContentType() string {
	return "application/json"
}
