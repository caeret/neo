// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package access provides an access logging handler for the ozzo routing package.
package access

import (
	"fmt"
	"net/http"
	"time"

	"github.com/caeret/neo"
)

// LogFunc logs a message using the given format and optional arguments.
// The usage of format and arguments is similar to that for fmt.Printf().
// LogFunc should be thread safe.
type LogFunc func(format string, a ...interface{})

// LogWriterFunc takes in the request and responseWriter objects as well
// as a float64 containing the elapsed time since the request first passed
// through this middleware and does whatever log writing it wants with that
// information.
// LogWriterFunc should be thread safe.
type LogWriterFunc func(c *neo.Context, res *LogResponseWriter, elapsed float64)

// CustomLogger returns a handler that calls the LogWriterFunc passed to it for every request.
// The LogWriterFunc is provided with the http.Request and LogResponseWriter objects for the
// request, as well as the elapsed time since the request first came through the middleware.
// LogWriterFunc can then do whatever logging it needs to do.
//
//	import (
//	    "log"
//	    "github.com/caeret/neo"
//	    "github.com/caeret/neo/access"
//	    "net/http"
//	)
//
//	func myCustomLogger(req http.Context, res access.LogResponseWriter, elapsed int64) {
//	    // Do something with the request, response, and elapsed time data here
//	}
//	r := mat.New()
//	r.Use(access.CustomLogger(myCustomLogger))
func CustomLogger(loggerFunc LogWriterFunc) neo.Handler {
	return func(c *neo.Context) error {
		startTime := time.Now()

		rw := &LogResponseWriter{c.Response, http.StatusOK, 0}
		c.Response = rw

		err := c.Next()

		elapsed := float64(time.Now().Sub(startTime).Nanoseconds()) / 1e6
		loggerFunc(c, rw, elapsed)

		return err
	}
}

// Logger returns a handler that logs a message for every request.
// The access log messages contain information including client IPs, time used to serve each request, request line,
// response status and size.
//
//	import (
//	    "log"
//	    "github.com/caeret/neo"
//	    "github.com/caeret/neo/access"
//	)
//
//	r := mat.New()
//	r.Use(access.Logger(log.Printf))
func Logger(log LogFunc) neo.Handler {
	logger := func(c *neo.Context, rw *LogResponseWriter, elapsed float64) {
		clientIP := c.RealIP()
		requestLine := fmt.Sprintf("%s %s %s", c.Request.Method, c.Request.URL.String(), c.Request.Proto)
		log(`[%s] [%.3fms] %s %d %d`, clientIP, elapsed, requestLine, rw.Status, rw.BytesWritten)
	}
	return CustomLogger(logger)
}

// LogResponseWriter wraps http.ResponseWriter in order to capture HTTP status and response length information.
type LogResponseWriter struct {
	http.ResponseWriter
	Status       int
	BytesWritten int64
}

func (r *LogResponseWriter) Write(p []byte) (int, error) {
	written, err := r.ResponseWriter.Write(p)
	r.BytesWritten += int64(written)
	return written, err
}

// WriteHeader records the response status and then writes HTTP headers.
func (r *LogResponseWriter) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}
