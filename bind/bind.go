package bind

import "net/http"

type Binder interface {
	Bind(req *http.Request, target interface{}) error
}


var JSON Binder = (*jsonBinder)(nil)
var FORM Binder = (*formBinder)(nil)