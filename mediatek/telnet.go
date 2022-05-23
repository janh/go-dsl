// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package mediatek

import (
	"3e8.eu/go/dsl"
	"3e8.eu/go/dsl/internal/telnet"
	"3e8.eu/go/dsl/models"
)

type telnetClient struct {
	client  *telnet.Client
	rawData []byte
	status  models.Status
	bins    models.Bins
}

func NewTelnetClient(config TelnetConfig) (dsl.Client, error) {
	c := telnetClient{}

	var err error

	user := config.User
	if user == "" {
		user = "admin"
	}

	clientConfig := telnet.ClientConfig{
		Prompts: []telnet.Prompts{
			telnet.Prompts{
				Account:  "login:",
				Password: "Password:",
				Command:  "# ",
			},
			// ZTE (from ZXHN H168N V3)
			telnet.Prompts{
				Account:  "Login: ",
				Password: "Password: ",
				Command:  "# ",
			},
		},
	}
	c.client, err = telnet.NewClient(clientConfig, config.Host, user, config.Password)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *telnetClient) RawData() []byte {
	return c.rawData
}

func (c *telnetClient) Status() models.Status {
	return c.status
}

func (c *telnetClient) Bins() models.Bins {
	return c.bins
}

func (c *telnetClient) UpdateData() (err error) {
	c.status, c.bins, c.rawData, err = updateData(c.client)
	return
}

func (c *telnetClient) Close() {
	c.client.Close()
}
