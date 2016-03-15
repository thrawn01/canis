package canis

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Handle func(http.ResponseWriter, *http.Request, Context)

func (self *Router) GET(path string, handle Handle) {
	self.router.Handle("GET", path, handle)
}

type Router struct {
	router httprouter.Router
}

func Router() *httprouter.Router {
	return &Router{
		&httprouter.Router{
			RedirectTrailingSlash:  true,
			RedirectFixedPath:      true,
			HandleMethodNotAllowed: true,
			HandleOPTIONS:          true,
		},
	}
}
