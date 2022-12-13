// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lancom

import (
	"3e8.eu/go/dsl"
	"3e8.eu/go/dsl/internal/snmp"
	"3e8.eu/go/dsl/models"
)

type client struct {
	client  *snmp.Client
	rawData []byte
	status  models.Status
	bins    models.Bins
}

func NewClient(config Config) (dsl.Client, error) {
	c := client{}

	var err error

	c.client, err = snmp.NewClient(config.Host, "udp", config.User,
		snmp.AuthProtocol(config.AuthProtocol), snmp.PrivacyProtocol(config.PrivacyProtocol),
		config.Password, config.EncryptionPassphrase)
	if err != nil {
		return nil, err
	}

	err = c.client.CheckResult(lcsStatusVdslLineState, 0x2)
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
	values, err := c.client.Walk(lcsStatusVdsl)
	if err != nil {
		return
	}

	c.status = parseStatus(values)
	c.bins = parseBins(&c.status, values)
	c.rawData = []byte(values.String())

	return
}

func (c *client) Close() {
	c.client.Close()
}
