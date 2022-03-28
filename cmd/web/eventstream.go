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

	"github.com/andybalholm/brotli"
)

type writeFlushCloser interface {
	io.Writer
	io.Closer
	Flush() error
}

type eventStreamWriter struct {
	writer  http.ResponseWriter
	flusher http.Flusher
	zWriter writeFlushCloser
}

func getCompressedWriter(rw http.ResponseWriter, req *http.Request) writeFlushCloser {
	acceptEncoding := req.Header.Get("Accept-Encoding")

	encodingsMap := make(map[string]bool)

	encodings := strings.Split(acceptEncoding, ",")
	for _, encoding := range encodings {
		if index := strings.IndexRune(encoding, ';'); index != -1 {
			encoding = encoding[:index]
		}
		encoding = strings.TrimSpace(encoding)

		encodingsMap[strings.ToLower(encoding)] = true
	}

	if encodingsMap["br"] {
		rw.Header().Set("Content-Encoding", "br")
		options := brotli.WriterOptions{
			Quality: brotli.DefaultCompression,
			LGWin:   18,
		}
		return brotli.NewWriterOptions(rw, options)
	}

	if encodingsMap["gzip"] {
		rw.Header().Set("Content-Encoding", "gzip")
		return gzip.NewWriter(rw)
	}

	return nil
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
	e.writer.Header().Set("X-Accel-Buffering", "no")

	e.zWriter = getCompressedWriter(rw, req)

	return e, nil
}

func (e *eventStreamWriter) WriteMessage(msg string) error {
	data := "data: " + msg + "\n\n"

	if e.zWriter != nil {
		_, err := io.WriteString(e.zWriter, data)
		if err != nil {
			return err
		}
		err = e.zWriter.Flush()
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
	if e.zWriter != nil {
		return e.zWriter.Close()
	}

	return nil
}
