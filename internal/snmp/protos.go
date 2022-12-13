// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package snmp

import (
	"errors"

	"github.com/gosnmp/gosnmp"
)

type AuthProtocol string

const (
	AuthProtocolNone   = ""
	AuthProtocolMD5    = "md5"
	AuthProtocolSHA    = "sha"
	AuthProtocolSHA224 = "sha224"
	AuthProtocolSHA256 = "sha256"
	AuthProtocolSHA384 = "sha384"
	AuthProtocolSHA512 = "sha512"
)

func (p AuthProtocol) proto() (gosnmp.SnmpV3AuthProtocol, error) {
	switch p {
	case AuthProtocolNone:
		return gosnmp.NoAuth, nil
	case AuthProtocolMD5:
		return gosnmp.MD5, nil
	case AuthProtocolSHA:
		return gosnmp.SHA, nil
	case AuthProtocolSHA224:
		return gosnmp.SHA224, nil
	case AuthProtocolSHA256:
		return gosnmp.SHA256, nil
	case AuthProtocolSHA384:
		return gosnmp.SHA384, nil
	case AuthProtocolSHA512:
		return gosnmp.SHA512, nil
	default:
		return gosnmp.NoAuth, errors.New("invalid authentication protocol")
	}
}

func (p AuthProtocol) desc() (string, error) {
	switch p {
	case AuthProtocolNone:
		return "none", nil
	case AuthProtocolMD5:
		return "MD5", nil
	case AuthProtocolSHA:
		return "SHA", nil
	case AuthProtocolSHA224:
		return "SHA-224", nil
	case AuthProtocolSHA256:
		return "SHA-256", nil
	case AuthProtocolSHA384:
		return "SHA-384", nil
	case AuthProtocolSHA512:
		return "SHA-512", nil
	default:
		return "", errors.New("invalid authentication protocol")
	}
}

type PrivacyProtocol string

const (
	PrivacyProtocolNone    = ""
	PrivacyProtocolDES     = "des"
	PrivacyProtocolAES128  = "aes128"
	PrivacyProtocolAES192  = "aes192"
	PrivacyProtocolAES192C = "aes192c"
	PrivacyProtocolAES256  = "aes256"
	PrivacyProtocolAES256C = "aes256c"
)

func (p PrivacyProtocol) proto() (gosnmp.SnmpV3PrivProtocol, error) {
	switch p {
	case PrivacyProtocolNone:
		return gosnmp.NoPriv, nil
	case PrivacyProtocolDES:
		return gosnmp.DES, nil
	case PrivacyProtocolAES128:
		return gosnmp.AES, nil
	case PrivacyProtocolAES192:
		return gosnmp.AES192, nil
	case PrivacyProtocolAES192C:
		return gosnmp.AES192C, nil
	case PrivacyProtocolAES256:
		return gosnmp.AES256, nil
	case PrivacyProtocolAES256C:
		return gosnmp.AES256C, nil
	default:
		return gosnmp.NoPriv, errors.New("invalid privacy protocol")
	}
}

func (p PrivacyProtocol) desc() (string, error) {
	switch p {
	case PrivacyProtocolNone:
		return "none", nil
	case PrivacyProtocolDES:
		return "DES", nil
	case PrivacyProtocolAES128:
		return "AES-128", nil
	case PrivacyProtocolAES192:
		return "AES-192", nil
	case PrivacyProtocolAES192C:
		return "AES-192-C", nil
	case PrivacyProtocolAES256:
		return "AES-256", nil
	case PrivacyProtocolAES256C:
		return "AES-256-C", nil
	default:
		return "", errors.New("invalid privacy protocol")
	}
}
