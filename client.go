// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package dsl // import "3e8.eu/go/dsl"

import (
	"errors"
	"sort"
	"time"

	"3e8.eu/go/dsl/models"
)

type AuthenticationError struct {
	Err      error
	WaitTime time.Duration
}

func (e *AuthenticationError) Error() string {
	return e.Err.Error()
}

func (e *AuthenticationError) Unwrap() error {
	return e.Err
}

type ConnectionError struct {
	Err error
}

func (e *ConnectionError) Error() string {
	return e.Err.Error()
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}

type Client interface {
	RawData() []byte
	Status() models.Status
	Bins() models.Bins
	UpdateData() error
	Close()
}

func NewClient(config Config) (Client, error) {
	newFunc, ok := getClientNewFunc(config.Type)
	if !ok {
		return nil, errors.New("invalid client type")
	}

	return newFunc(config)
}

func GetClientTypes() []ClientType {
	clientTypes := getClientTypes()

	sort.Slice(clientTypes, func(i, j int) bool { return clientTypes[i] < clientTypes[j] })

	return clientTypes
}
