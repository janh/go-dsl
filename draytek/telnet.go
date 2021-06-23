// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package draytek

import (
	"fmt"
	"strings"

	"3e8.eu/go/dsl/models"
	"3e8.eu/go/dsl/telnet"
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

	clientConfig := telnet.ClientConfig{
		PromptAccount:  "Account:",
		PromptPassword: "Password:",
		PromptCommand:  "> ",
	}
	c.client, err = telnet.NewClient(clientConfig, config.Host, "admin", config.Password)
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

func (c *TelnetClient) UpdateData() error {
	status, err := c.client.Execute("adsl status")
	if err != nil {
		return err
	}

	counts, err := c.client.Execute("adsl status counts")
	if err != nil {
		return err
	}

	downstream, err := c.client.Execute("adsl showbins")
	if err != nil {
		return err
	}

	upstream, err := c.client.Execute("adsl showbins up")
	if err != nil {
		return err
	}

	c.status = parseStatus(status, counts)
	c.bins = parseBins(c.status, downstream, upstream)

	var b strings.Builder
	fmt.Fprintln(&b, "# adsl status")
	fmt.Fprintln(&b, status)
	fmt.Fprintln(&b, "# adsl status counts")
	fmt.Fprintln(&b, counts)
	fmt.Fprintln(&b, "# adsl showbins")
	fmt.Fprintln(&b, downstream)
	fmt.Fprintln(&b, "# adsl showbins up")
	fmt.Fprintln(&b, upstream)
	fmt.Fprintln(&b)
	c.rawData = []byte(b.String())

	return nil
}

func (c *TelnetClient) Close() {
	c.client.Close()
}
