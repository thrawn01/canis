// Package alice provides a convenient way to chain http handlers.
package canis

import (
	"net/http"
)

type Middleware func(http.Handler) http.Handler

type MiddlewareChain struct {
	middleware []Middleware
}

// Create a new chain
func Chain(middleware ...Middleware) *MiddlewareChain {
	return &MiddlewareChain{append(([]Middleware)(nil), middleware...)}
}

// End the chain and return the http.Handler
func (self *MiddlewareChain) Then(handler http.Handler) http.Handler {
	for i := len(self.middleware) - 1; i >= 0; i-- {
		handler = self.middleware[i](handler)
	}

	return handler
}

// Same as Then(), but accepts a Middleware
func (self *MiddlewareChain) ThenFunc(handlerFunc http.HandlerFunc) http.Handler {
	if handlerFunc == nil {
		return self.Then(nil)
	}
	return self.Then(http.HandlerFunc(handlerFunc))
}

// Add middleware to the chain
func (self *MiddlewareChain) Add(middleware ...Middleware) *MiddlewareChain {
	self.middleware = append(self.middleware, middleware...)
	return self
}

// Add middleware to the chain
func (self *MiddlewareChain) Use(middleware ...Middleware) *MiddlewareChain {
	self.middleware = append(self.middleware, middleware...)
	return self
}

// Creates a new chain from the existing chain new middleware
func (self *MiddlewareChain) Extend(middleware ...Middleware) *MiddlewareChain {
	new := make([]Middleware, len(self.middleware)+len(middleware))
	// Copy all the existing middleware to the new chain list
	copy(new, self.middleware)
	// Append the middleware passed in to the end of the middleware list
	copy(new[len(self.middleware):], middleware)
	// Return the new chain
	return Chain(new...)
}

