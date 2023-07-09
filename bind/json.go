package bind

import (
	"encoding/json"
	"net/http"
)

type jsonBinder struct {
}


func (j *jsonBinder) Bind(req *http.Request, target interface{}) error {
	return json.NewDecoder(req.Body).Decode(target)
}