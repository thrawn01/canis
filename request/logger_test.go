package request_test

import (
	"testing"

	"net/http"
	"net/http/httptest"

	"log"

	"bufio"
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thrawn01/canis"
	"github.com/thrawn01/canis/request"
	"golang.org/x/net/context"
)

func TestLogger(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Request Logger")
}

var _ = Describe("request", func() {
	var app canis.ContextHandler
	var resp *httptest.ResponseRecorder

	/*BeforeEach(func() {
	})*/

	Describe("Logger()", func() {
		It("should log normal 200 class requests", func() {
			app = canis.ContextHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				w.Write([]byte("payload"))
			})
			resp = httptest.NewRecorder()
			byteBuf := new(bytes.Buffer)
			writer := bufio.NewWriter(byteBuf)

			chain := canis.Chain(
				request.Logger(log.New(writer, "", 0)),
			)
			req, _ := http.NewRequest("GET", "/", nil)
			handler := chain.Then(app)
			handler.ServeHTTP(resp, req)
			writer.Flush()

			Expect(resp.Body.String()).To(Equal("payload"))
			Expect(byteBuf.String()).To(ContainSubstring("GET \"/\" HTTP/1.1 200 7"))
		})
	})

	Describe("ErrorLogger()", func() {
		It("should log the return payload for non 200 class requests", func() {
			app = canis.ContextHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(500)
				w.Write([]byte("some error"))
			})
			resp = httptest.NewRecorder()
			byteBuf := new(bytes.Buffer)
			writer := bufio.NewWriter(byteBuf)

			chain := canis.Chain(
				request.ErrorLogger(log.New(writer, "", 0)),
			)
			req, _ := http.NewRequest("GET", "/", nil)
			handler := chain.Then(app)
			handler.ServeHTTP(resp, req)
			writer.Flush()

			Expect(resp.Body.String()).To(Equal("some error"))
			Expect(byteBuf.String()).To(ContainSubstring("GET \"/\" HTTP/1.1 500 10 - some error"))
		})
	})
})
