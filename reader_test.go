package neo

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type FA struct {
	A1 string
	A2 int
}

type FB struct {
	B1 string
	B2 bool
	B3 float64
}

func TestReadForm(t *testing.T) {
	var a struct {
		X1 string `form:"x1"`
		FA
		X2 int
		B  *FB
		FB `form:"c"`
		E  FB `form:"e"`
		c  int
		D  []int
	}
	values := map[string][]string{
		"x1":   {"abc", "123"},
		"A1":   {"a1"},
		"x2":   {"1", "2"},
		"B.B1": {"b1", "b2"},
		"B.B2": {"true"},
		"B.B3": {"1.23"},
		"c.B1": {"fb1", "fb2"},
		"e.B1": {"fe1", "fe2"},
		"c":    {"100"},
		"D":    {"100", "200", "300"},
	}
	err := ReadFormData(values, &a)
	assert.Nil(t, err)
	assert.Equal(t, "abc", a.X1)
	assert.Equal(t, "a1", a.A1)
	assert.Equal(t, 0, a.X2)
	assert.Equal(t, "b1", a.B.B1)
	assert.True(t, a.B.B2)
	assert.Equal(t, 1.23, a.B.B3)
	assert.Equal(t, "fb1", a.B1)
	assert.Equal(t, "fe1", a.E.B1)
	assert.Equal(t, 0, a.c)
	assert.Equal(t, []int{100, 200, 300}, a.D)
}

func TestDefaultDataReader(t *testing.T) {
	tests := []struct {
		tag         string
		header      string
		method, URL string
		body        string
	}{
		{"t1", "", "GET", "/test?A1=abc&A2=100", ""},
		{"t2", "", "POST", "/test?A1=abc&A2=100", ""},
		{"t3", "application/x-www-form-urlencoded", "POST", "/test", "A1=abc&A2=100"},
		{"t4", "application/json", "POST", "/test", `{"A1":"abc","A2":100}`},
		{"t5", "application/xml", "POST", "/test", `<data><A1>abc</A1><A2>100</A2></data>`},
	}

	expected := FA{
		A1: "abc",
		A2: 100,
	}
	for _, test := range tests {
		var data FA
		req, _ := http.NewRequest(test.method, test.URL, bytes.NewBufferString(test.body))
		req.Header.Set("Content-Type", test.header)
		c := NewContext(nil, req)
		err := c.Read(&data)
		assert.Nil(t, err, test.tag)
		assert.Equal(t, expected, data, test.tag)
	}
}

type TU struct {
	UValue string
}

func (tu *TU) UnmarshalText(text []byte) error {
	tu.UValue = "TU_" + string(text[:])
	return nil
}

func TestTextUnmarshaler(t *testing.T) {
	var a struct {
		ATU TU     `form:"atu"`
		NTU string `form:"ntu"`
	}
	values := map[string][]string{
		"atu": {"ORIGINAL"},
		"ntu": {"ORIGINAL"},
	}
	err := ReadFormData(values, &a)
	assert.Nil(t, err)
	assert.Equal(t, "TU_ORIGINAL", a.ATU.UValue)
	assert.Equal(t, "ORIGINAL", a.NTU)
}

func TestRead(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(`{"foo":"bar"}`))
	req.Header.Set("Content-Type", "application/json")
	c := NewContext(nil, req)
	type a struct {
		Foo string `json:"foo"`
	}
	v, err := Read[a](c)
	if err != nil {
		t.Errorf("read fail: %s", err)
		return
	}
	if v.Foo != "bar" {
		t.Errorf("read fail: %s", v.Foo)
	}
}
