// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package speedport

import (
	"fmt"
	"strings"

	"3e8.eu/go/dsl"
	"3e8.eu/go/dsl/models"
)

type client struct {
	session *session
	rawData []byte
	status  models.Status
	bins    models.Bins
}

func NewClient(config Config) (dsl.Client, error) {
	c := client{}

	var err error

	c.session, err = newSession(config.Host, config.Password, config.TLSSkipVerify)
	if err != nil {
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
	rawVersion, valuesVersion, err := c.session.loadData("/engineer/data/Version.json")
	if err != nil {
		return
	}

	rawDSL, valuesDSL, err := c.session.loadData("/engineer/data/DSL.json")
	if err != nil {
		return
	}

	var b strings.Builder
	fmt.Fprintln(&b, "/engineer/data/Version.json")
	b.Write(rawVersion)
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "/engineer/data/DSL.json")
	b.Write(rawDSL)
	fmt.Fprintln(&b)
	c.rawData = []byte(b.String())

	c.status = interpretStatus(valuesVersion, valuesDSL)
	c.bins = interpretBins(&c.status, valuesDSL)

	return
}

func (c *client) Close() {
	c.session.close()
}
