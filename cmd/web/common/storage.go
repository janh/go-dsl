// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package common

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"3e8.eu/go/dsl/history"
)

func (c *Client) readStateFile(filename string, readFunc func(r io.Reader) error) (err error) {
	file, err := os.Open(filepath.Join(c.stateDir, filename))
	if err != nil {
		return
	}
	defer func() {
		closeErr := file.Close()
		if err == nil {
			err = closeErr
		}
	}()

	reader, err := gzip.NewReader(file)
	if err != nil {
		return
	}
	defer func() {
		closeErr := reader.Close()
		if err == nil {
			err = closeErr
		}
	}()

	err = readFunc(reader)
	return
}

func (c *Client) writeStateFile(filename string, writeFunc func(w io.Writer) error) (err error) {
	err = os.MkdirAll(c.stateDir, os.ModePerm)
	if err != nil {
		return
	}

	file, err := os.Create(filepath.Join(c.stateDir, filename))
	if err != nil {
		return
	}
	defer func() {
		closeErr := file.Close()
		if err == nil {
			err = closeErr
		}
	}()

	writer := gzip.NewWriter(file)
	defer func() {
		closeErr := writer.Close()
		if err == nil {
			err = closeErr
		}
	}()

	err = writeFunc(writer)
	return
}

func (c *Client) loadHistory(bins *history.Bins, errors *history.Errors) {
	if c.stateDir == "" {
		return
	}

	err := c.readStateFile("bins.dat.gz", bins.Load)
	if err != nil && !os.IsNotExist(err) {
		fmt.Println("failed to load bins history:", err)
	}

	err = c.readStateFile("errors.dat.gz", errors.Load)
	if err != nil && !os.IsNotExist(err) {
		fmt.Println("failed to load errors history:", err)
	}
}

func (c *Client) saveHistory(bins *history.Bins, errors *history.Errors) {
	if c.stateDir == "" {
		return
	}

	err := c.writeStateFile("bins.dat.gz", bins.Save)
	if err != nil {
		fmt.Println("failed to save bins history:", err)
	}

	err = c.writeStateFile("errors.dat.gz", errors.Save)
	if err != nil {
		fmt.Println("failed to save errors history:", err)
	}
}
