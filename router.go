package canis

import "github.com/thrawn01/httprouter"

func Router() *httprouter.Router {
	return httprouter.New()
}

type ParamContext interface {
	httprouter.ParamContext
}
