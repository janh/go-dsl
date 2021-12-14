// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"errors"
	"io"
	"net/http"
)

type eventStreamWriter struct {
	writer  http.ResponseWriter
	flusher http.Flusher
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

	return e, nil
}

func (e *eventStreamWriter) WriteMessage(msg string) error {
	data := "data: " + msg + "\n\n"

	_, err := io.WriteString(e.writer, data)
	if err != nil {
		return err
	}

	e.flusher.Flush()

	return nil
}
