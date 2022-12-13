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
	AuthTypePrivateKeys
)

type OptionType int

const (
	OptionTypeString = iota
	OptionTypeBool
	OptionTypeEnum
)

type OptionValue struct {
	Value string
	Title string
}

type Option struct {
	Description string
	Type        OptionType
	Values      []OptionValue
}

type ClientDesc struct {
	Title                        string
	RequiresUser                 Tristate
	SupportedAuthTypes           AuthTypes
	RequiresKnownHosts           bool
	SupportsEncryptionPassphrase bool
	Options                      map[string]Option
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
