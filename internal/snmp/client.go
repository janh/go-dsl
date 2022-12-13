// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package snmp

import (
	"errors"
	"fmt"
	"time"

	"github.com/gosnmp/gosnmp"

	"3e8.eu/go/dsl"
	"3e8.eu/go/dsl/internal/helpers"
)

type Client struct {
	client *gosnmp.GoSNMP
}

func NewClient(host, transport, username string,
	authProto AuthProtocol, privacyProto PrivacyProtocol,
	password dsl.PasswordCallback,
	encryptionPassphrase dsl.EncryptionPassphraseCallback) (*Client, error) {

	c := Client{}

	err := c.setup(host, transport, username, authProto, privacyProto, password, encryptionPassphrase)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *Client) setup(host, transport, username string,
	authProtocol AuthProtocol, privacyProtocol PrivacyProtocol,
	passwordCallback dsl.PasswordCallback,
	encryptionPassphraseCallback dsl.EncryptionPassphraseCallback) error {

	host, port, err := helpers.SplitHostPort(host)
	if err != nil {
		return err
	}
	if port == 0 {
		port = 161
	}

	securityParams := &gosnmp.UsmSecurityParameters{
		UserName: username,
	}

	c.client = &gosnmp.GoSNMP{
		Target:             host,
		Port:               port,
		Transport:          transport,
		Version:            gosnmp.Version3,
		SecurityModel:      gosnmp.UserSecurityModel,
		SecurityParameters: securityParams,
		Timeout:            5 * time.Second,
		Retries:            2,
	}

	if authProtocol != AuthProtocolNone {
		authProto, err := authProtocol.proto()
		if err != nil {
			return err
		}

		securityParams.AuthenticationProtocol = authProto
		if passwordCallback != nil {
			securityParams.AuthenticationPassphrase, err = passwordCallback()
			if err != nil {
				return err
			}
		}

		if privacyProtocol != AuthProtocolNone {
			privacyProto, err := privacyProtocol.proto()
			if err != nil {
				return err
			}

			securityParams.PrivacyProtocol = privacyProto
			if encryptionPassphraseCallback != nil {
				securityParams.PrivacyPassphrase, err = encryptionPassphraseCallback()
				if err != nil {
					return err
				}
			}

			c.client.MsgFlags = gosnmp.AuthPriv
		} else {
			c.client.MsgFlags = gosnmp.AuthNoPriv
		}

	} else {
		if privacyProtocol != AuthProtocolNone {
			return errors.New("privacy without authentication is not supported")
		}
		c.client.MsgFlags = gosnmp.NoAuthNoPriv
	}

	return c.client.Connect()
}

func (c *Client) CheckResult(oid string, expectedType byte) error {
	result, err := c.client.Get([]string{oid})
	if err != nil {
		return err
	}

	variablesCount := len(result.Variables)
	if variablesCount != 1 {
		return fmt.Errorf("check failed: unexpected number of variables (%d)",
			variablesCount)
	}

	actualOID := result.Variables[0].Name
	if actualOID != oid {
		return fmt.Errorf("check failed: expected OID %s, but received %s",
			oid, actualOID)
	}

	actualTypeAsn1BER := result.Variables[0].Type
	actualType := byte(actualTypeAsn1BER)
	if actualType != expectedType {
		expectedTypeAsn1BER := gosnmp.Asn1BER(expectedType)
		return fmt.Errorf("check failed: expected type 0x%x (%s), but received 0x%x (%s)",
			expectedType, expectedTypeAsn1BER, actualType, actualTypeAsn1BER)
	}

	return nil
}

func (c *Client) Walk(oid string) (Values, error) {
	var v Values
	v.init()

	err := c.client.BulkWalk(oid, func(pdu gosnmp.SnmpPDU) error {
		v.add(Value{
			OID:  pdu.Name,
			Type: byte(pdu.Type),
			Val:  pdu.Value,
		})

		return nil
	})

	return v, err
}

func (c *Client) Close() error {
	return c.client.Conn.Close()
}
