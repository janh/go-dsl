// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lancom

import (
	"errors"
	"strings"

	"3e8.eu/go/dsl"
	"3e8.eu/go/dsl/internal/snmp"
	"3e8.eu/go/dsl/models"
)

type client struct {
	client  *snmp.Client
	oidBase string
	rawData []byte
	status  models.Status
	bins    models.Bins
}

func getBase(subtree string) ([]string, error) {
	subtree = strings.ToLower(subtree)
	if len(subtree) >= 2 && subtree[0] == '/' {
		subtree = subtree[1:]
	}

	switch subtree {
	case "status/vdsl":
		return []string{lcsStatusVdsl}, nil
	case "status/xdsl/vdsl1":
		return []string{lcsStatusXdslVdsl1}, nil
	case "status/xdsl/vdsl2":
		return []string{lcsStatusXdslVdsl2}, nil
	case "status/adsl":
		return []string{lcsStatusAdsl}, nil
	case "status/xdsl/adsl":
		return []string{lcsStatusXdslAdsl}, nil
	case "":
		return []string{lcsStatusVdsl, lcsStatusAdsl}, nil
	default:
		return nil, errors.New("unrecognized subtree value")
	}
}

func NewClient(config Config) (dsl.Client, error) {
	c := client{}

	var err error

	var baseList []string
	baseList, err = getBase(config.Subtree)
	if err != nil {
		return nil, err
	}

	c.client, err = snmp.NewClient(config.Host, "udp", config.User,
		snmp.AuthProtocol(config.AuthProtocol), snmp.PrivacyProtocol(config.PrivacyProtocol),
		config.Password, config.EncryptionPassphrase)
	if err != nil {
		return nil, err
	}

	for _, base := range baseList {
		err = c.client.CheckResult(base+oidLineState, 0x2)
		if err == nil {
			c.oidBase = base
			break
		}
	}
	if err != nil {
		c.client.Close()
		return nil, err
	}

	return &c, nil
}

func (c *client) RawData() []byte {
	return c.rawData
}

func (c *client) Status() models.Status {
	return c.status
}

func (c *client) Bins() models.Bins {
	return c.bins
}

func (c *client) UpdateData() (err error) {
	values, err := c.client.Walk(c.oidBase)
	if err != nil {
		return
	}

	c.status = parseStatus(values, c.oidBase)
	c.bins = parseBins(&c.status, values, c.oidBase)
	c.rawData = []byte(values.String())

	return
}

func (c *client) Close() {
	c.client.Close()
}
