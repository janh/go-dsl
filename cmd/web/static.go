// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"
	"time"
)

type staticItem interface {
	ReadSeeker() (io.ReadSeeker, error)
}

type staticItemFile struct {
	FS   fs.FS
	Path string
}

func (i staticItemFile) ReadSeeker() (io.ReadSeeker, error) {
	file, err := i.FS.Open(i.Path)
	if err != nil {
		return nil, err
	}

	reader, ok := file.(io.ReadSeeker)
	if !ok {
		return nil, errors.New("file does not implement io.ReadSeeker")
	}

	return reader, nil
}

type staticItemBytes struct {
	Data []byte
}

func (i staticItemBytes) ReadSeeker() (io.ReadSeeker, error) {
	return bytes.NewReader(i.Data), nil
}

type staticHandler struct {
	items map[string]staticItemWithETag
}

type staticItemWithETag struct {
	staticItem
	etag string
}

func (s *staticHandler) Add(path string, item staticItem) error {
	reader, err := item.ReadSeeker()
	if err != nil {
		return err
	}

	hash := sha256.New()
	_, err = io.Copy(hash, reader)
	if err != nil {
		return err
	}

	etag := fmt.Sprintf(`"%x"`, hash.Sum(nil))

	if s.items == nil {
		s.items = make(map[string]staticItemWithETag)
	}

	s.items[path] = staticItemWithETag{
		staticItem: item,
		etag:       etag,
	}

	return nil
}

func (s *staticHandler) MustAdd(path string, item staticItem) {
	err := s.Add(path, item)
	if err != nil {
		panic(err)
	}
}

func (s *staticHandler) AddFS(targetPath string, fileSystem fs.ReadDirFS, fsPath string) error {
	if fsPath != "" {
		subFS, err := fs.Sub(fileSystem, fsPath)
		if err != nil {
			return err
		}
		fileSystem = subFS.(fs.ReadDirFS)
	}

	return fs.WalkDir(fileSystem, ".", func(itemPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		return s.Add(path.Join(targetPath, itemPath), staticItemFile{fileSystem, itemPath})
	})
}

func (s *staticHandler) MustAddFS(targetPath string, fileSystem fs.ReadDirFS, fsPath string) {
	err := s.AddFS(targetPath, fileSystem, fsPath)
	if err != nil {
		panic(err)
	}
}

func (s *staticHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	item, ok := s.items[req.URL.Path]
	if !ok {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}

	name := path.Base(req.URL.Path)
	reader, err := item.ReadSeeker()
	if err != nil {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}

	w.Header().Set("ETag", item.etag)
	http.ServeContent(w, req, name, time.Time{}, reader)
}
