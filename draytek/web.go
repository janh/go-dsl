// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package draytek

import (
	"3e8.eu/go/dsl"
	"3e8.eu/go/dsl/models"
)

type webClient struct {
	session *webSession
	rawData []byte
	status  models.Status
	bins    models.Bins
}

func NewWebClient(config WebConfig) (dsl.Client, error) {
	c := webClient{}

	var err error

	user := config.User
	if user == "" {
		user = "admin"
	}

	c.session, err = newWebSession(config.Host, config.User, config.Password, config.TLSSkipVerify)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *webClient) RawData() []byte {
	return c.rawData
}

func (c *webClient) Status() models.Status {
	return c.status
}

func (c *webClient) Bins() models.Bins {
	return c.bins
}

func (c *webClient) UpdateData() (err error) {
	c.status, c.bins, c.rawData, err = updateData(c.session)
	return
}

func (c *webClient) Close() {
	c.session.close()
}
