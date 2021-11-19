// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package sagemcom

import (
	"encoding/json"
	"fmt"

	"3e8.eu/go/dsl/models"
)

type Client struct {
	session *session
	rawData []byte
	status  models.Status
	bins    models.Bins
}

func NewClient(config Config) (*Client, error) {
	c := Client{}

	var err error

	user := config.User
	if user == "" {
		user = "admin"
	}

	c.session, err = newSession(config.Host, user, config.Password, config.TLSSkipVerify)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *Client) RawData() []byte {
	return c.rawData
}

func (c *Client) Status() models.Status {
	return c.status
}

func (c Client) Bins() models.Bins {
	return c.bins
}

func (c *Client) UpdateData() (err error) {
	c.rawData, err = c.session.loadValue("Device/DSL")
	if err != nil {
		return
	}

	var data dslWrapper
	err = json.Unmarshal(c.rawData, &data)
	if err != nil {
		return
	}

	if len(data.DSL.Lines) != 1 || len(data.DSL.Channels) != 1 {
		return fmt.Errorf("unexpected number of lines (%d) or channels (%d)",
			len(data.DSL.Lines), len(data.DSL.Channels))
	}

	c.status = interpretStatus(&data.DSL)
	c.bins = interpretBins(&c.status, &data.DSL)

	return
}

func (c *Client) Close() {
	c.session.close()
}
