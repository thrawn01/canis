package request

import (
	"bytes"
	"net"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/oxtoacart/bpool"
	"github.com/thrawn01/canis"
	"github.com/thrawn01/httprouter"
)

var PoolSize = 20000

type WrappedResponseWriter interface {
	http.ResponseWriter
	WriteLogPostfix(*bytes.Buffer)
}

/*
 ResponseLogger - Wrap the ResponseWriter so we can intercept the status code and the size of
		  the response body set by upstream middleware
*/
type ResponseLogger struct {
	resp   http.ResponseWriter
	status int
	size   int
}

func (self *ResponseLogger) Header() http.Header {
	return self.resp.Header()
}

func (self *ResponseLogger) Write(buf []byte) (int, error) {
	self.size += len(buf)
	return self.resp.Write(buf)
}

func (self *ResponseLogger) WriteHeader(status int) {
	self.status = status
	self.resp.WriteHeader(status)
}

func (self *ResponseLogger) WriteLogPostfix(buf *bytes.Buffer) {
	// Status Code
	buf.WriteString(strconv.Itoa(self.status))
	buf.WriteString(" ")
	// Result Size
	buf.WriteString(strconv.Itoa(self.size))
}

/*
 ErrorResponseLogger - Same as ResponseLogger but also captures the response buffer for non HTTP 200 return codes
*/
type ErrorResponseLogger struct {
	*ResponseLogger
	errorMsg []byte
}

func (self *ErrorResponseLogger) Write(buf []byte) (int, error) {
	// If the status NOT a 2XX error
	if !((self.status - 200) < 100) {
		self.errorMsg = buf
	}
	self.size += len(buf)
	return self.resp.Write(buf)
}

func (self *ErrorResponseLogger) WriteLogPostfix(buf *bytes.Buffer) {
	// Status Code
	buf.WriteString(strconv.Itoa(self.status))
	buf.WriteString(" ")
	// Result Size
	buf.WriteString(strconv.Itoa(self.size))

	// Write the error message returned
	if self.errorMsg != nil {
		buf.WriteString(" - ")
		buf.Write(bytes.TrimSuffix(self.errorMsg, []byte("\n")))
		self.errorMsg = nil
	}
}

/*
 Create a new instance of the request logger
*/
func Logger(log logrus.StdLogger) canis.Middleware {
	req := &RequestLogger{bpool.NewBufferPool(PoolSize), log, false}
	return req.Handler
}

/*
 Just like Logger() but also logs the payload if the HTTP status code is non 200
*/
func ErrorLogger(log logrus.StdLogger) canis.Middleware {
	req := &RequestLogger{bpool.NewBufferPool(PoolSize), log, true}
	return req.Handler
}

// Writes apache style access logs using an efficient buffer pool to avoid
// garbage collection
type RequestLogger struct {
	bufferPool *bpool.BufferPool
	log        logrus.StdLogger
	capture    bool
}

func (self *RequestLogger) Handler(handler httprouter.ContextHandler) httprouter.ContextHandler {
	return httprouter.ContextHandlerFunc(func(ctx context.Context, originalResp http.ResponseWriter, req *http.Request) {
		buf := self.bufferPool.Get()

		// TODO: Parse X-Forward-Host if present

		// Add the client remote address
		remoteAddress, _, err := net.SplitHostPort(req.RemoteAddr)
		if err != nil {
			remoteAddress = req.RemoteAddr
		}

		buf.WriteString(remoteAddress)
		buf.WriteString(" - ")

		// Add the authenticated user (if none then '-')
		if req.URL.User != nil {
			if name := req.URL.User.Username(); name != "" {
				buf.WriteString(name)
			} else {
				buf.WriteString("-")
			}
		} else {
			buf.WriteString("-")
		}

		// Time
		buf.WriteString(" [")
		buf.WriteString(time.Now().Format("01/Jan/2015:01:01:01 -0600"))
		buf.WriteString("] ")
		// Http Verb
		buf.WriteString(" ")
		buf.WriteString(req.Method)
		// Uri
		buf.WriteString(" \"")
		buf.WriteString(req.URL.RequestURI())
		buf.WriteString("\" ")
		// Proto
		buf.WriteString(req.Proto)
		buf.WriteString(" ")

		var resp WrappedResponseWriter
		if self.capture {
			resp = &ErrorResponseLogger{&ResponseLogger{originalResp, 200, 0}, nil}
		} else {
			resp = &ResponseLogger{originalResp, 200, 0}
		}
		// Call up the middleware chain
		handler.ServeHTTP(ctx, resp, req)

		resp.WriteLogPostfix(buf)
		// Write out the log entry
		self.log.Println(buf)
		// Put the buffer back into the pool
		self.bufferPool.Put(buf)
	})
}
