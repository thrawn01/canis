// This code is based on the excellent 'alice' module  https://github.com/justinas/alice
package canis

import (
	"net/http"

	"fmt"

	"golang.org/x/net/context"
)

type Middleware func(ContextHandler) ContextHandler

type MiddlewareChain struct {
	middleware []Middleware
}

// Create a new chain
func Chain(middleware ...interface{}) *MiddlewareChain {
	chain := &MiddlewareChain{}
	chain.Add(middleware...)
	return chain
}

// End the chain and return the http.Handler
func (self *MiddlewareChain) Then(handler ContextHandler) http.Handler {
	for i := len(self.middleware) - 1; i >= 0; i-- {
		handler = self.middleware[i](handler)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		handler.ServeHTTP(context.Background(), w, req)
	})
}

// Same as Then(), but accepts a ContextHandlerFunc
func (self *MiddlewareChain) ThenFunc(handlerFunc ContextHandlerFunc) http.Handler {
	if handlerFunc == nil {
		return self.Then(nil)
	}
	return self.Then(ContextHandlerFunc(handlerFunc))
}

// Add middleware to the chain
func (self *MiddlewareChain) Add(middleware ...interface{}) *MiddlewareChain {
	self.middleware = appendMiddleware(self.middleware, middleware...)
	return self
}

// Add middleware to the chain
func (self *MiddlewareChain) Use(middleware ...interface{}) *MiddlewareChain {
	self.middleware = appendMiddleware(self.middleware, middleware...)
	return self
}

// Creates a new chain from the existing chain new middleware
func (self *MiddlewareChain) Extend(middleware ...interface{}) *MiddlewareChain {
	new := make([]Middleware, len(self.middleware))
	// Copy all the existing middleware to the new chain list
	copy(new, self.middleware)
	// Append the middleware passed in to the end of the middleware list
	new = appendMiddleware(new, middleware...)
	// Return the new chain
	return &MiddlewareChain{new}
}

func appendMiddleware(dest []Middleware, middleware ...interface{}) []Middleware {
	for _, ware := range middleware {
		switch t := ware.(type) {

		// Normal Middleware
		case Middleware:
			dest = append(dest, t)
		// Handle http.Handler middleware
		case http.Handler:
			// NOTE: http.Handler middleware can not catch panic's as they are not in the chain,
			// they also can not modify http.ResponseWriter and expect modifications to propigated up the chain
			wrapper := func(next ContextHandler) ContextHandler {
				return ContextHandlerFunc(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
					// Call our normal http.Handler style middleware
					t.ServeHTTP(resp, req)
					// Continue to call the next middleware in the stack
					next.ServeHTTP(ctx, resp, req)
				})
			}
			dest = append(dest, wrapper)
		default:
			panic(fmt.Sprintf("unsupported middleware handler signature %T", t))
		}

	}
	return dest
}
