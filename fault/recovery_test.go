// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fault

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

func TestRecovery(t *testing.T) {
	var buf bytes.Buffer
	h := Recovery(getLogger(&buf))

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/", nil)
	c := neo.NewContext(res, req, h, handler1, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "abc", res.Body.String())
	assert.Equal(t, "abc", buf.String())

	buf.Reset()
	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users/", nil)
	c = neo.NewContext(res, req, h, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "test", res.Body.String())
	assert.Equal(t, "", buf.String())

	buf.Reset()
	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users/", nil)
	c = neo.NewContext(res, req, h, handler3, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "xyz", res.Body.String())
	assert.Contains(t, buf.String(), "recovery_test.go")
	assert.Contains(t, buf.String(), "xyz")

	buf.Reset()
	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users/", nil)
	c = neo.NewContext(res, req, h, handler4, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, http.StatusBadRequest, res.Code)
	assert.Equal(t, "123", res.Body.String())
	assert.Contains(t, buf.String(), "recovery_test.go")
	assert.Contains(t, buf.String(), "123")

	buf.Reset()
	h = Recovery(getLogger(&buf), convertError)
	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users/", nil)
	c = neo.NewContext(res, req, h, handler3, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "123", res.Body.String())
	assert.Contains(t, buf.String(), "recovery_test.go")
	assert.Contains(t, buf.String(), "xyz")

	buf.Reset()
	h = Recovery(getLogger(&buf), convertError)
	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users/", nil)
	c = neo.NewContext(res, req, h, handler1, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "123", res.Body.String())
	assert.Equal(t, "abc", buf.String())
}

func getLogger(buf *bytes.Buffer) LogFunc {
	return func(format string, a ...interface{}) {
		fmt.Fprintf(buf, format, a...)
	}
}

func handler1(c *neo.Context) error {
	return errors.New("abc")
}

func handler2(c *neo.Context) error {
	c.Write("test")
	return nil
}

func handler3(c *neo.Context) error {
	panic("xyz")
}

func handler4(c *neo.Context) error {
	panic(neo.NewHTTPError(http.StatusBadRequest, "123"))
}
