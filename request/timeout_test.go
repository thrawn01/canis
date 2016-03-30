package request_test

import (
	"testing"

	"net/http"
	"net/http/httptest"

	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thrawn01/canis"
	"github.com/thrawn01/canis/request"
	"golang.org/x/net/context"
)

func TestTimeout(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Request Timeout")
}

var _ = Describe("request", func() {
	var app canis.ContextHandler
	var resp *httptest.ResponseRecorder

	Describe("Timeout()", func() {
		It("should close the Done() channel once duration is reached", func() {
			app = canis.ContextHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
				sleep := make(chan struct{}, 1)
				go func() {
					time.Sleep(time.Second)
					w.Write([]byte("no timeout"))
					close(sleep)
				}()

				select {
				case <-ctx.Done():
					w.Write([]byte("timeout"))
					return

				case <-sleep:
					Fail("Timeout was not called", 0)
					return
				}
			})
			resp = httptest.NewRecorder()
			chain := canis.Chain(
				request.Timeout(time.Millisecond * 100),
			)
			req, _ := http.NewRequest("GET", "/", nil)
			handler := chain.Then(app)
			handler.ServeHTTP(resp, req)

			Expect(resp.Body.String()).To(Equal("timeout"))
		})
		It("should NOT close the Done() channel is duration is not reached", func() {
			app = canis.ContextHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
				sleep := make(chan struct{}, 1)
				go func() {
					time.Sleep(time.Millisecond * 100)
					w.Write([]byte("no timeout"))
					close(sleep)
				}()

				select {
				case <-ctx.Done():
					w.Write([]byte("timeout"))
					return

				case <-sleep:
					return
				}
			})
			resp = httptest.NewRecorder()
			chain := canis.Chain(
				request.Timeout(time.Second),
			)
			req, _ := http.NewRequest("GET", "/", nil)
			handler := chain.Then(app)
			handler.ServeHTTP(resp, req)
			Expect(resp.Body.String()).To(Equal("no timeout"))
		})
	})
	Describe("OnTimeout()", func() {
		It("Should call the passed handler when timeout is reached", func() {
			app = canis.ContextHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
				sleep := make(chan struct{}, 1)
				go func() {
					time.Sleep(time.Second)
					w.Write([]byte("no timeout"))
					close(sleep)
				}()

				select {
				case <-ctx.Done():
					return

				case <-sleep:
					Fail("Timeout was not called", 0)
					return
				}
			})
			resp = httptest.NewRecorder()
			chain := canis.Chain(
				request.OnTimeout(time.Millisecond*100, func(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
					resp.Write([]byte("timeout handler"))
				}),
			)
			req, _ := http.NewRequest("GET", "/", nil)
			handler := chain.Then(app)
			handler.ServeHTTP(resp, req)

			Expect(resp.Body.String()).To(Equal("timeout handler"))
		})
		It("should NOT call the passed handler if timeout is not reached", func() {
			app = canis.ContextHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
				sleep := make(chan struct{}, 1)
				go func() {
					time.Sleep(time.Millisecond * 100)
					w.Write([]byte("no timeout"))
					close(sleep)
				}()

				select {
				case <-ctx.Done():
					w.Write([]byte("timeout"))
					return

				case <-sleep:
					return
				}
			})
			resp = httptest.NewRecorder()
			chain := canis.Chain(
				request.OnTimeout(time.Second, func(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
					resp.Write([]byte("timeout handler"))
				}),
			)
			req, _ := http.NewRequest("GET", "/", nil)
			handler := chain.Then(app)
			handler.ServeHTTP(resp, req)
			Expect(resp.Body.String()).To(Equal("no timeout"))
		})
	})
})
