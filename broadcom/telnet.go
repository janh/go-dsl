// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package broadcom

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
		PromptAccount:  "login: ",
		PromptPassword: "Password: ",
		PromptCommand:  "# ",
	}
	c.client, err = telnet.NewClient(clientConfig, config.Host, "root", config.Password)
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
	stats, err := c.client.Execute("xdslctl info --stats")
	if err != nil {
		return err
	}

	vendor, err := c.client.Execute("xdslctl info --vendor")
	if err != nil {
		return err
	}

	version, err := c.client.Execute("xdslctl --version")
	if err != nil {
		return err
	}

	pbParams, err := c.client.Execute("xdslctl info --pbParams")
	if err != nil {
		return err
	}

	bits, err := c.client.Execute("xdslctl info --Bits")
	if err != nil {
		return err
	}

	snr, err := c.client.Execute("xdslctl info --SNR")
	if err != nil {
		return err
	}

	qln, err := c.client.Execute("xdslctl info --QLN")
	if err != nil {
		return err
	}

	hlog, err := c.client.Execute("xdslctl info --Hlog")
	if err != nil {
		return err
	}

	c.status = parseStatus(stats, vendor, version)
	c.bins = parseBins(c.status, pbParams, bits, snr, qln, hlog)

	var b strings.Builder
	fmt.Fprintln(&b, "# xdslctl info --stats")
	fmt.Fprintln(&b, stats)
	fmt.Fprintln(&b, "# xdslctl info --vendor")
	fmt.Fprintln(&b, vendor)
	fmt.Fprintln(&b, "# xdslctl info --version")
	fmt.Fprintln(&b, version)
	fmt.Fprintln(&b, "# xdslctl info --pbParams")
	fmt.Fprintln(&b, pbParams)
	fmt.Fprintln(&b, "# xdslctl info --Bits")
	fmt.Fprintln(&b, bits)
	fmt.Fprintln(&b, "# xdslctl info --SNR")
	fmt.Fprintln(&b, snr)
	fmt.Fprintln(&b, "# xdslctl info --QLN")
	fmt.Fprintln(&b, qln)
	fmt.Fprintln(&b, "# xdslctl info --Hlog")
	fmt.Fprintln(&b, hlog)
	fmt.Fprintln(&b)
	c.rawData = []byte(b.String())

	return nil
}

func (c *TelnetClient) Close() {
	c.client.Close()
}
