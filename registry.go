// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package dsl

import (
	"errors"
	"sync"
)

type registryItem struct {
	New  func(config Config) (Client, error)
	Desc ClientDesc
}

var (
	registryItems map[ClientType]registryItem
	registryMutex sync.Mutex
)

func init() {
	registryItems = make(map[ClientType]registryItem)
}

// RegisterClient registers a new device client. This function is not intended for use from external
// packages.
func RegisterClient(identifier ClientType, newFunc func(config Config) (Client, error), desc ClientDesc) {
	registryMutex.Lock()
	defer registryMutex.Unlock()

	if _, ok := registryItems[identifier]; ok {
		panic(errors.New("client type identifier already in use"))
	}

	registryItems[identifier] = registryItem{New: newFunc, Desc: desc}
}

func getClientDesc(identifier ClientType) (desc ClientDesc, ok bool) {
	registryMutex.Lock()
	defer registryMutex.Unlock()

	item, ok := registryItems[identifier]
	desc = item.Desc
	return
}

func getClientNewFunc(identifier ClientType) (newFunc func(config Config) (Client, error), ok bool) {
	registryMutex.Lock()
	defer registryMutex.Unlock()

	item, ok := registryItems[identifier]
	newFunc = item.New
	return
}

func getClientTypes() []ClientType {
	registryMutex.Lock()
	defer registryMutex.Unlock()

	clientTypes := make([]ClientType, 0, len(registryItems))

	for clientType := range registryItems {
		clientTypes = append(clientTypes, clientType)
	}

	return clientTypes
}
