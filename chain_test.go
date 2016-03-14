package canis_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thrawn01/canis"
	"testing"
	"net/http"
	"net/http/httptest"
)

func TestArgs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Args Parser")
}

func newMiddleware(body string) canis.Middleware {
	body = body + "|"
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(body))
			h.ServeHTTP(w, r)
		})
	}
}


var _ = Describe("canis.Chain()", func() {
	var app http.Handler
	var resp *httptest.ResponseRecorder

	BeforeEach(func() {
		app = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("app"))
		})
		resp = httptest.NewRecorder()
	})

	It("should create a new MiddlewareChain object", func() {
		chain := canis.Chain(newMiddleware("one"))
		req, _ := http.NewRequest("GET", "/", nil)

		handler := chain.Then(app)
		handler.ServeHTTP(resp, req)
		Expect(resp.Body.String()).To(Equal("one|app"))
	})
})
