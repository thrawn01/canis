package canis

import (
	"time"

	"github.com/julienschmidt/httprouter"
)

type Context interface {
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}
	Err() error
	ByName(string)
	Value(key interface{}) interface{}
}

type paramContext struct {
	Context
	params httprouter.Params
}

func main() {
	router := httprouter.New()
	router.GET()
}
