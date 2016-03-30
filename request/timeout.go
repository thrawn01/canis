package request

import (
	"net/http"
	"time"

	"github.com/thrawn01/canis"
	"golang.org/x/net/context"
)

func Timeout(timeout time.Duration) canis.Middleware {
	return func(next canis.ContextHandler) canis.ContextHandler {
		return canis.ContextHandlerFunc(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
			ctx, _ = context.WithTimeout(ctx, timeout)
			next.ServeHTTP(ctx, resp, req)
		})
	}
}

func OnTimeout(timeout time.Duration, handler canis.ContextHandlerFunc) canis.Middleware {
	return func(next canis.ContextHandler) canis.ContextHandler {
		return canis.ContextHandlerFunc(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			next.ServeHTTP(ctx, resp, req)
			cancel()
			if ctx.Err() == context.DeadlineExceeded {
				handler(ctx, resp, req)
			}
		})
	}
}
