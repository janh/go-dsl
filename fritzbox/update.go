// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package fritzbox

import (
	"encoding/json"
	"errors"
	"net/url"
)

func checkPageID(data, page string) bool {
	var decoded struct {
		PageID string `json:"pid"`
	}

	err := json.Unmarshal([]byte(data), &decoded)
	if err != nil {
		return false
	}

	return decoded.PageID == page
}

func (c *client) updateOverview(d *rawDataOverview) (err error) {
	// contains HTML for version < 7.19, JSON for version >= 7.19
	data := url.Values{}
	data.Add("lang", "de")
	data.Add("page", "dslOv")
	data.Add("xhr", "1")
	d.Data, err = c.session.loadPost("/data.lua", data)
	if err != nil {
		var httpErr *httpError
		if errors.As(err, &httpErr) && httpErr.StatusCode == 404 {
			// only versions < 6.50 (?)
			d.Ancient = true
			data = url.Values{}
			d.Data, err = c.session.loadGet("/internet/dsl_overview.lua", data)
			if err != nil {
				return
			}
		} else {
			return
		}
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
	// version >= 7.39
	data := url.Values{}
	data.Add("lang", "de")
	data.Add("page", "dslStat")
	data.Add("xhr", "1")
	d.Data, err = c.session.loadPost("/data.lua", data)
	if err != nil {
		var httpErr *httpError
		if errors.As(err, &httpErr) && httpErr.StatusCode == 404 {
			// only versions < 6.50 (?)
			d.Ancient = true
		} else {
			return
		}
	}

	if !checkPageID(d.Data, "dslStat") {
		// version < 7.39
		d.Legacy = true

		data = url.Values{}
		data.Add("update", "mainDiv")
		data.Add("useajax", "1")
		data.Add("xhr", "1")
		d.Data, err = c.session.loadGet("/internet/dsl_stats_tab.lua", data)
		if err != nil {
			return
		}
	}

	return
}

func (c *client) updateSpectrum(d *rawDataSpectrum) (err error) {
	// version >= 7.39
	data := url.Values{}
	data.Add("lang", "de")
	data.Add("page", "dslSpectrum")
	data.Add("xhr", "1")
	d.Data, err = c.session.loadPost("/data.lua", data)
	if err != nil {
		var httpErr *httpError
		if errors.As(err, &httpErr) && httpErr.StatusCode == 404 {
			// only versions < 6.50 (?)
			d.Ancient = true
		} else {
			return
		}
	}

	if !checkPageID(d.Data, "dslSpectrum") {
		// version < 7.39
		d.Legacy = true

		data = url.Values{}
		data.Add("myXhr", "1")
		data.Add("useajax", "1")
		data.Add("xhr", "1")
		d.Data, err = c.session.loadGet("/internet/dsl_spectrum.lua", data)
		if err != nil {
			return
		}
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
