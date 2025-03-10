// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package access

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/caeret/neo"
)

func TestCustomLogger(t *testing.T) {
	var buf bytes.Buffer
	customFunc := func(c *neo.Context, rw *LogResponseWriter, elapsed float64) {
		logWriter := getLogger(&buf)
		clientIP := c.RealIP()
		requestLine := fmt.Sprintf("%s %s %s", c.Request.Method, c.Request.URL.String(), c.Request.Proto)
		logWriter(`[%s] [%.3fms] %s %d %d`, clientIP, elapsed, requestLine, rw.Status, rw.BytesWritten)
	}
	h := CustomLogger(customFunc)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://127.0.0.1/users", nil)
	c := neo.NewContext(res, req, h, handler1)
	assert.NotNil(t, c.Next())
	assert.Contains(t, buf.String(), "GET http://127.0.0.1/users")
}

func TestLogger(t *testing.T) {
	var buf bytes.Buffer
	h := Logger(getLogger(&buf))

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://127.0.0.1/users", nil)
	c := neo.NewContext(res, req, h, handler1)
	assert.NotNil(t, c.Next())
	assert.Contains(t, buf.String(), "GET http://127.0.0.1/users")
}

func TestLogResponseWriter(t *testing.T) {
	res := httptest.NewRecorder()
	w := &LogResponseWriter{res, 0, 0}
	w.WriteHeader(http.StatusBadRequest)
	assert.Equal(t, http.StatusBadRequest, res.Code)
	assert.Equal(t, http.StatusBadRequest, w.Status)
	n, _ := w.Write([]byte("test"))
	assert.Equal(t, 4, n)
	assert.Equal(t, int64(4), w.BytesWritten)
	assert.Equal(t, "test", res.Body.String())
}

func getLogger(buf *bytes.Buffer) LogFunc {
	return func(format string, a ...interface{}) {
		fmt.Fprintf(buf, format, a...)
	}
}

func handler1(c *neo.Context) error {
	return errors.New("abc")
}
