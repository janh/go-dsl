// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package fritzbox

import (
	"fmt"
	"net/url"
	"strings"

	"3e8.eu/go/dsl/models"
)

type Client struct {
	loadSupportData bool
	session         *session
	rawData         []byte
	status          models.Status
	bins            models.Bins
}

func NewClient(config Config) (*Client, error) {
	c := Client{}
	c.loadSupportData = config.LoadSupportData

	var err error

	c.session, err = newSession(config.Host, config.User, config.Password, config.TLSSkipVerify)
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
	// contains HTML for version < 7.19, JSON for version >= 7.19
	data := url.Values{}
	data.Add("lang", "de")
	data.Add("page", "dslOv")
	data.Add("xhr", "1")
	dslOverview, err := c.session.loadPost("/data.lua", data)
	if err != nil {
		return err
	}

	// only for version < 7.19: contains JSON
	var dslOverviewData string
	if len(dslOverview) == 0 || dslOverview[0] != '{' {
		data = url.Values{}
		data.Add("action", "get_data")
		data.Add("myXhr", "1")
		data.Add("useajax", "1")
		data.Add("xhr", "1")
		dslOverviewData, err = c.session.loadGet("/internet/dsl_overview.lua", data)
		if err != nil {
			return err
		}
	}

	data = url.Values{}
	data.Add("update", "mainDiv")
	data.Add("useajax", "1")
	data.Add("xhr", "1")
	dslStats, err := c.session.loadGet("/internet/dsl_stats_tab.lua", data)
	if err != nil {
		return err
	}

	data = url.Values{}
	data.Add("myXhr", "1")
	data.Add("useajax", "1")
	data.Add("xhr", "1")
	dslSpectrum, err := c.session.loadGet("/internet/dsl_spectrum.lua", data)
	if err != nil {
		return err
	}

	interfaceConfigInfo, err := c.session.loadTR064(
		"/upnp/control/wandslifconfig1",
		"urn:dslforum-org:service:WANDSLInterfaceConfig:1",
		"GetInfo")
	if err != nil {
		return err
	}

	interfaceConfigStatisticsTotal, err := c.session.loadTR064(
		"/upnp/control/wandslifconfig1",
		"urn:dslforum-org:service:WANDSLInterfaceConfig:1",
		"GetStatisticsTotal")
	if err != nil {
		return err
	}

	var supportData string
	if c.loadSupportData {
		supportData, err = c.session.loadSupportData()
		if err != nil {
			return err
		}
	}

	c.status = models.Status{}
	parseOverview(&c.status, dslOverview, dslOverviewData)
	parseStats(&c.status, dslStats)
	parseTR064Data(&c.status, interfaceConfigInfo, interfaceConfigStatisticsTotal)

	c.bins = models.Bins{}
	parseSpectrum(&c.bins, &c.status, dslSpectrum)
	parseSupportData(&c.status, &c.bins, supportData)

	var b strings.Builder
	fmt.Fprintln(&b, "////// DSL Overview\n")
	fmt.Fprintln(&b, dslOverview+"\n")
	fmt.Fprintln(&b, "////// DSL Overview data\n")
	fmt.Fprintln(&b, dslOverviewData+"\n")
	fmt.Fprintln(&b, "////// DSL Stats\n")
	fmt.Fprintln(&b, dslStats+"\n")
	fmt.Fprintln(&b, "////// DSL Spectrum\n")
	fmt.Fprintln(&b, dslSpectrum+"\n")
	fmt.Fprintln(&b, "////// Interface Config Info\n")
	fmt.Fprintln(&b, interfaceConfigInfo+"\n")
	fmt.Fprintln(&b, "////// Interface Config Statistics Total\n")
	fmt.Fprintln(&b, interfaceConfigStatisticsTotal+"\n")
	fmt.Fprintln(&b, "////// Support Data\n")
	fmt.Fprintln(&b, supportData+"\n")
	fmt.Fprintln(&b)
	c.rawData = []byte(b.String())

	return
}

func (c *Client) Close() {
	c.session.close()
}
