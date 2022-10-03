// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package fritzbox

import (
	"3e8.eu/go/dsl"
	"3e8.eu/go/dsl/models"
)

type client struct {
	loadSupportData bool
	session         *session
	rawData         []byte
	status          models.Status
	bins            models.Bins
}

func NewClient(config Config) (dsl.Client, error) {
	c := client{}
	c.loadSupportData = config.LoadSupportData

	var err error

	c.session, err = newSession(config.Host, config.User, config.Password, config.TLSSkipVerify)
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
	var d rawData

	err = c.updateOverview(&d.Overview)
	if err != nil {
		return err
	}

	err = c.updateStats(&d.Stats)
	if err != nil {
		return err
	}

	err = c.updateSpectrum(&d.Spectrum)
	if err != nil {
		return err
	}

	err = c.updateTR064(&d.TR064)
	if err != nil {
		return err
	}

	if c.loadSupportData {
		err = c.updateSupportData(&d.SupportData)
		if err != nil {
			return err
		}
	}

	c.status = models.Status{}
	c.bins = models.Bins{}

	parseOverview(&c.status, &d.Overview)
	parseStats(&c.status, &d.Stats)
	parseSpectrum(&c.bins, &c.status, &d.Spectrum)
	parseTR064Data(&c.status, &d.TR064)
	parseSupportData(&c.status, &c.bins, &d.SupportData)

	c.rawData = []byte(d.String())

	return
}

func (c *client) Close() {
	c.session.close()
}
