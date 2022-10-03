// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package fritzbox

import (
	"net/url"
)

func (c *client) updateOverview(d *rawDataOverview) (err error) {
	// contains HTML for version < 7.19, JSON for version >= 7.19
	data := url.Values{}
	data.Add("lang", "de")
	data.Add("page", "dslOv")
	data.Add("xhr", "1")
	d.Data, err = c.session.loadPost("/data.lua", data)
	if err != nil {
		return
	}

	if len(d.Data) == 0 || d.Data[0] != '{' {
		// only for version < 7.19: contains JSON
		d.Legacy = true

		data = url.Values{}
		data.Add("action", "get_data")
		data.Add("myXhr", "1")
		data.Add("useajax", "1")
		data.Add("xhr", "1")
		d.UpdateData, err = c.session.loadGet("/internet/dsl_overview.lua", data)
		if err != nil {
			return
		}
	}

	return
}

func (c *client) updateStats(d *rawDataStats) (err error) {
	data := url.Values{}
	data.Add("update", "mainDiv")
	data.Add("useajax", "1")
	data.Add("xhr", "1")
	d.Data, err = c.session.loadGet("/internet/dsl_stats_tab.lua", data)
	if err != nil {
		return
	}

	return
}

func (c *client) updateSpectrum(d *rawDataSpectrum) (err error) {
	data := url.Values{}
	data.Add("myXhr", "1")
	data.Add("useajax", "1")
	data.Add("xhr", "1")
	d.Data, err = c.session.loadGet("/internet/dsl_spectrum.lua", data)
	if err != nil {
		return
	}

	return
}

func (c *client) updateTR064(d *rawDataTR064) (err error) {
	d.InterfaceConfigInfo, err = c.session.loadTR064(
		"/upnp/control/wandslifconfig1",
		"urn:dslforum-org:service:WANDSLInterfaceConfig:1",
		"GetInfo")
	if err != nil {
		return
	}

	d.InterfaceConfigStatisticsTotal, err = c.session.loadTR064(
		"/upnp/control/wandslifconfig1",
		"urn:dslforum-org:service:WANDSLInterfaceConfig:1",
		"GetStatisticsTotal")
	if err != nil {
		return
	}

	return
}

func (c *client) updateSupportData(d *rawDataSupport) (err error) {
	d.Data, err = c.session.loadSupportData()
	if err != nil {
		return
	}

	return
}
