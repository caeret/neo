// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package content provides content negotiation handlers for the ozzo routing package.
package content

import (
	"encoding/xml"
	"net/http"

	sonicencoder "github.com/bytedance/sonic/encoder"

	"github.com/caeret/neo"
)

// MIME types
const (
	JSON = neo.MIME_JSON
	XML  = neo.MIME_XML
	XML2 = neo.MIME_XML2
	HTML = neo.MIME_HTML
)

// DataWriters lists all supported content types and the corresponding data writers.
// By default, JSON, XML, and HTML are supported. You may modify this variable before calling TypeNegotiator
// to customize supported data writers.
var DataWriters = map[string]neo.DataWriter{
	JSON: &JSONDataWriter{},
	XML:  &XMLDataWriter{},
	XML2: &XMLDataWriter{},
	HTML: &HTMLDataWriter{},
}

// TypeNegotiator returns a content type negotiation handler.
//
// The method takes a list of response MIME types that are supported by the application.
// The negotiator will determine the best response MIME type to use by checking the "Accept" HTTP header.
// If no match is found, the first MIME type will be used.
//
// The negotiator will set the "Content-Type" response header as the chosen MIME type. It will call mat.Context.SetDataWriter()
// to set the appropriate data writer that can write data in the negotiated format.
//
// If you do not specify any supported MIME types, the negotiator will use "text/html" as the response MIME type.
func TypeNegotiator(formats ...string) neo.Handler {
	if len(formats) == 0 {
		formats = []string{HTML}
	}
	for _, format := range formats {
		if _, ok := DataWriters[format]; !ok {
			panic(format + " is not supported")
		}
	}

	return func(c *neo.Context) error {
		format := NegotiateContentType(c.Request, formats, formats[0])
		c.SetDataWriter(DataWriters[format])
		return nil
	}
}

// JSONDataWriter sets the "Content-Type" response header as "application/json" and writes the given data in JSON format to the response.
type JSONDataWriter struct{}

// SetHeader sets the Content-Type response header.
func (w *JSONDataWriter) SetHeader(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
}

func (w *JSONDataWriter) Write(res http.ResponseWriter, data interface{}) (err error) {
	enc := sonicencoder.NewStreamEncoder(res)
	enc.SetEscapeHTML(false)
	return enc.Encode(data)
}

// XMLDataWriter sets the "Content-Type" response header as "application/xml; charset=UTF-8" and writes the given data in XML format to the response.
type XMLDataWriter struct{}

// SetHeader sets the Content-Type response header.
func (w *XMLDataWriter) SetHeader(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/xml; charset=UTF-8")
}

func (w *XMLDataWriter) Write(res http.ResponseWriter, data interface{}) (err error) {
	var bytes []byte
	if bytes, err = xml.Marshal(data); err != nil {
		return
	}
	_, err = res.Write(bytes)
	return
}

// HTMLDataWriter sets the "Content-Type" response header as "text/html; charset=UTF-8" and calls mat.DefaultDataWriter to write the given data to the response.
type HTMLDataWriter struct{}

// SetHeader sets the Content-Type response header.
func (w *HTMLDataWriter) SetHeader(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "text/html; charset=UTF-8")
}

func (w *HTMLDataWriter) Write(res http.ResponseWriter, data interface{}) error {
	return neo.DefaultDataWriter.Write(res, data)
}
