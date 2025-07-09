// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package fritzbox

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"

	"3e8.eu/go/dsl/internal/helpers"
	"3e8.eu/go/dsl/internal/htmlutil"
	"3e8.eu/go/dsl/models"
)

type overviewDataJSON struct {
	Data struct {
		ConnectionData struct {
			Version       string `json:"version"`
			ExternApText  string `json:"externApText"`
			ExternApValue string `json:"externApValue"`

			Lines []struct {
				State string `json:"state"`
				Mode  string `json:"mode"`
				Time  string `json:"time"`
			} `json:"line"`
		} `json:"connectionData"`
	} `json:"data"`
}

type overviewDataLegacy struct {
	DSLAM string `json:"dslam"`
	Lines []struct {
		State      string `json:"state"`
		TrainState string `json:"train_state"`
		Mode       string `json:"mode"`
		Time       string `json:"time"`
	} `json:"line"`
}

func parseOverview(status *models.Status, d *rawDataOverview) {
	if !d.Legacy {
		parseOverviewJSON(status, d.Data)
	} else {
		if d.Ancient {
			parseOverviewLegacyAncient(status, d.Data)
		} else {
			parseOverviewLegacy(status, d.Data)
		}
		parseOverviewDataLegacy(status, d.UpdateData)
	}
}

func interpretOverviewState(state string) models.State {
	switch {
	case strings.HasPrefix(state, "ready"):
		return models.StateShowtime
	case state == "training":
		return models.StateInit
	case state == "off":
		return models.StateDown
	case state == "error":
		return models.StateError
	}
	return models.StateUnknown
}

func interpretOverviewStateAncient(trainState string) models.State {
	trainState = strings.ToLower(trainState)
	switch {
	case strings.HasPrefix(trainState, "dsl aktiv"):
		return models.StateShowtime
	case strings.HasPrefix(trainState, "training"):
		return models.StateInit
	case strings.HasPrefix(trainState, "nicht verbunden"):
		return models.StateDown
	}
	return models.StateUnknown
}

func interpretOverviewTime(timeStr string) (out models.Duration) {
	split := strings.Fields(timeStr)

	for i := 1; i < len(split); i++ {
		part := strings.ToLower(split[i])

		var factor time.Duration
		switch {
		case strings.HasPrefix(part, "minute"):
			factor = time.Minute
		case strings.HasPrefix(part, "stunde"):
			factor = time.Hour
		case strings.HasPrefix(part, "tag"):
			factor = 24 * time.Hour
		default:
			continue
		}

		val, err := strconv.ParseInt(split[i-1], 10, 64)
		if err != nil {
			out.Valid = false
			return
		}

		out.Valid = true
		out.Duration += time.Duration(val) * factor
	}

	return
}

func parseOverviewJSON(status *models.Status, dslOverview string) {
	var overviewData overviewDataJSON
	err := json.Unmarshal([]byte(dslOverview), &overviewData)
	if err != nil {
		return
	}
	data := &overviewData.Data.ConnectionData

	if len(data.Lines) > 0 {
		status.State = interpretOverviewState(data.Lines[0].State)
		status.Mode = helpers.ParseMode(data.Lines[0].Mode)

		if status.State == models.StateShowtime {
			status.Uptime = interpretOverviewTime(data.Lines[0].Time)
		}
	}

	status.NearEndInventory.Vendor = "AVM"
	status.NearEndInventory.Version = data.Version

	status.FarEndInventory.Vendor = data.ExternApText
	status.FarEndInventory.Version = strings.TrimPrefix(data.ExternApValue, "Version ")
}

func parseOverviewLegacyAncient(status *models.Status, dslOverview string) {
	doc, err := html.Parse(strings.NewReader(dslOverview))
	if err != nil {
		return
	}

	dslVersionNode := htmlutil.FindFirstNode(doc, htmlutil.MatcherTagNameAndClass("td", "dsl_txt_info"))
	if dslVersionNode != nil {
		nearEndVersionNode := htmlutil.FindLastNode(dslVersionNode, func(n *html.Node) bool {
			return n.Type == html.TextNode && strings.TrimSpace(n.Data) != ""
		})
		if nearEndVersionNode != nil {
			status.NearEndInventory.Vendor = "AVM"
			status.NearEndInventory.Version = strings.TrimSpace(nearEndVersionNode.Data)
		}
	}

	dslamVersionNode := htmlutil.FindFirstNode(doc, htmlutil.MatcherTagNameAndClass("td", "dsl_txt_info_dslam"))
	if dslamVersionNode != nil {
		farEndInventoryNodes := htmlutil.FindAllNodes(dslamVersionNode, func(n *html.Node) bool {
			return n.Type == html.TextNode && strings.TrimSpace(n.Data) != ""
		})
		if len(farEndInventoryNodes) >= 2 {
			vendor := farEndInventoryNodes[0].Data
			vendorStr := strings.TrimSpace(vendor)
			version := farEndInventoryNodes[1].Data
			versionStr := strings.TrimSpace(version)

			if !strings.Contains(vendorStr, "---") && versionStr != "0" {
				status.FarEndInventory.Vendor = vendorStr
				status.FarEndInventory.Version = versionStr
			}
		}
	}
}

func parseOverviewLegacy(status *models.Status, dslOverview string) {
	doc, err := html.Parse(strings.NewReader(dslOverview))
	if err != nil {
		return
	}

	tableRow := htmlutil.FindFirstNode(doc, htmlutil.MatcherTagNameAndClass("tr", "tInfo"))
	if tableRow == nil {
		return
	}

	columns := htmlutil.FindAllNodes(tableRow, htmlutil.MatcherTagName("td"))
	if len(columns) != 3 {
		return
	}

	nearEndVersionNode := htmlutil.FindLastNode(columns[0], func(n *html.Node) bool {
		return n.Type == html.TextNode && strings.TrimSpace(n.Data) != ""
	})
	if nearEndVersionNode != nil {
		status.NearEndInventory.Vendor = "AVM"
		status.NearEndInventory.Version = strings.TrimSpace(nearEndVersionNode.Data)
	}

	farEndInventoryNodes := htmlutil.FindAllNodes(columns[2], func(n *html.Node) bool {
		return n.Type == html.TextNode && strings.TrimSpace(n.Data) != ""
	})
	if len(farEndInventoryNodes) >= 3 {
		vendor := farEndInventoryNodes[1].Data
		vendorStr := strings.TrimSuffix(strings.TrimSpace(vendor), ":")
		version := farEndInventoryNodes[2].Data
		versionStr := strings.TrimSpace(version)

		if !strings.Contains(vendorStr, "---") && !strings.Contains(versionStr, "---") {
			status.FarEndInventory.Vendor = vendorStr
			status.FarEndInventory.Version = versionStr
		}
	}
}

func parseOverviewDataLegacy(status *models.Status, dslOverviewData string) {
	var data overviewDataLegacy
	err := json.Unmarshal([]byte(dslOverviewData), &data)
	if err != nil {
		return
	}

	if len(data.Lines) > 0 {
		status.State = interpretOverviewState(data.Lines[0].State)
		if status.State == models.StateUnknown {
			status.State = interpretOverviewStateAncient(data.Lines[0].TrainState)
		}

		status.Mode = helpers.ParseMode(data.Lines[0].Mode)

		if status.State == models.StateShowtime {
			status.Uptime = interpretOverviewTime(data.Lines[0].Time)
		}
	}
}
