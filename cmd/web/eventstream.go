// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"strings"
)

type eventStreamWriter struct {
	writer     http.ResponseWriter
	flusher    http.Flusher
	gzipWriter *gzip.Writer
}

func acceptsGzip(req *http.Request) bool {
	acceptEncoding := req.Header.Get("Accept-Encoding")

	encodings := strings.Split(acceptEncoding, ",")
	for _, encoding := range encodings {
		if index := strings.IndexRune(encoding, ';'); index != -1 {
			encoding = encoding[:index]
		}
		encoding = strings.TrimSpace(encoding)

		if strings.ToLower(encoding) == "gzip" {
			return true
		}
	}

	return false
}

func newEventStreamWriter(rw http.ResponseWriter, req *http.Request) (*eventStreamWriter, error) {
	e := &eventStreamWriter{
		writer: rw,
	}

	var ok bool
	e.flusher, ok = e.writer.(http.Flusher)
	if !ok {
		return nil, errors.New("given http.ResponseWriter is not a http.Flusher")
	}

	e.writer.Header().Set("Content-Type", "text/event-stream")
	e.writer.Header().Set("Cache-Control", "no-cache")
	e.writer.Header().Set("Connection", "keep-alive")

	if acceptsGzip(req) {
		e.writer.Header().Set("Content-Encoding", "gzip")
		e.gzipWriter = gzip.NewWriter(e.writer)
	}

	return e, nil
}

func (e *eventStreamWriter) WriteMessage(msg string) error {
	data := "data: " + msg + "\n\n"

	if e.gzipWriter != nil {
		_, err := io.WriteString(e.gzipWriter, data)
		if err != nil {
			return err
		}
		err = e.gzipWriter.Flush()
		if err != nil {
			return err
		}
	} else {
		_, err := io.WriteString(e.writer, data)
		if err != nil {
			return err
		}
	}

	e.flusher.Flush()

	return nil
}

func (e *eventStreamWriter) Close() error {
	if e.gzipWriter != nil {
		return e.gzipWriter.Close()
	}

	return nil
}
