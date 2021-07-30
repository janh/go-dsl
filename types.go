// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package dsl

type Tristate int

const (
	TristateMaybe Tristate = 0
	TristateNo    Tristate = -1
	TristateYes   Tristate = 1
)

type AuthTypes int

const (
	AuthTypePassword AuthTypes = 1 << iota
	AuthTypePrivateKey
)

type ClientDesc struct {
	RequiresUser       Tristate
	SupportedAuthTypes AuthTypes
	RequiresKnownHost  bool
}

type ClientType string

func (t ClientType) IsValid() bool {
	if _, ok := getClientDesc(t); ok {
		return true
	}
	return false
}

func (t ClientType) ClientDesc() ClientDesc {
	desc, _ := getClientDesc(t)
	return desc
}
