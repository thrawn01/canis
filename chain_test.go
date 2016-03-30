package canis_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thrawn01/canis"
	"golang.org/x/net/context"
)

func TestArgs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Chain")
}

func newMiddleware(body string) canis.Middleware {
	body = body + "|"
	return func(h canis.ContextHandler) canis.ContextHandler {
		return canis.ContextHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(body))
			h.ServeHTTP(ctx, w, r)
		})
	}
}

var _ = Describe("MiddlewareChain", func() {
	var app canis.ContextHandler
	var resp *httptest.ResponseRecorder

	BeforeEach(func() {
		app = canis.ContextHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("app"))
		})
		resp = httptest.NewRecorder()
	})

	Describe("canis.Chain()", func() {
		It("should create a new MiddlewareChain", func() {
			chain := canis.Chain(newMiddleware("one"))
			req, _ := http.NewRequest("GET", "/", nil)

			handler := chain.Then(app)
			handler.ServeHTTP(resp, req)
			Expect(resp.Body.String()).To(Equal("one|app"))
		})
	})
	Describe("MiddlwareChain.Add()", func() {
		It("should add a new middleware to the chain", func() {
			chain := canis.Chain(newMiddleware("one"))
			chain.Add(newMiddleware("two"))
			req, _ := http.NewRequest("GET", "/", nil)

			handler := chain.Then(app)
			handler.ServeHTTP(resp, req)
			Expect(resp.Body.String()).To(Equal("one|two|app"))
		})
	})
	Describe("MiddlwareChain.Use()", func() {
		It("should add a new middleware to the chain", func() {
			chain := canis.Chain(newMiddleware("one"))
			chain.Add(newMiddleware("two"))
			req, _ := http.NewRequest("GET", "/", nil)

			handler := chain.Then(app)
			handler.ServeHTTP(resp, req)
			Expect(resp.Body.String()).To(Equal("one|two|app"))
		})
	})
	Describe("MiddlwareChain.Extend()", func() {
		It("should create a new chain while adding new middleware to new chain", func() {
			chainParent := canis.Chain(newMiddleware("one"), newMiddleware("two"))
			newChain := chainParent.Extend(newMiddleware("three"), newMiddleware("four"))
			req, _ := http.NewRequest("GET", "/", nil)

			// Chain Parent
			handler := chainParent.Then(app)
			handler.ServeHTTP(resp, req)
			Expect(resp.Body.String()).To(Equal("one|two|app"))

			// New Chain
			resp = httptest.NewRecorder()
			handler = newChain.Then(app)
			handler.ServeHTTP(resp, req)
			Expect(resp.Body.String()).To(Equal("one|two|three|four|app"))
		})
	})

})
