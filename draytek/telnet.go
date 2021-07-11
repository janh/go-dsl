// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package draytek

import (
	"3e8.eu/go/dsl/internal/telnet"
	"3e8.eu/go/dsl/models"
)

type TelnetClient struct {
	client  *telnet.Client
	rawData []byte
	status  models.Status
	bins    models.Bins
}

func NewTelnetClient(config TelnetConfig) (*TelnetClient, error) {
	c := TelnetClient{}

	var err error

	user := config.User
	if user == "" {
		user = "admin"
	}

	clientConfig := telnet.ClientConfig{
		PromptAccount:  "Account:",
		PromptPassword: "Password:",
		PromptCommand:  "> ",
	}
	c.client, err = telnet.NewClient(clientConfig, config.Host, user, config.Password)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *TelnetClient) RawData() []byte {
	return c.rawData
}

func (c *TelnetClient) Status() models.Status {
	return c.status
}

func (c *TelnetClient) Bins() models.Bins {
	return c.bins
}

func (c *TelnetClient) UpdateData() (err error) {
	c.status, c.bins, c.rawData, err = updateData(c.client)
	return
}

func (c *TelnetClient) Close() {
	c.client.Close()
}
