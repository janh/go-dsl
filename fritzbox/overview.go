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
		State string `json:"state"`
		Mode  string `json:"mode"`
		Time  string `json:"time"`
	} `json:"line"`
}

func parseOverview(status *models.Status, dslOverview, dslOverviewData string) {
	if len(dslOverview) > 0 && dslOverview[0] == '{' {
		parseOverviewJSON(status, dslOverview)
	} else {
		parseOverviewLegacy(status, dslOverview)
		parseOverviewDataLegacy(status, dslOverviewData)
	}
}

func interpretOverviewState(state string) models.State {
	switch {
	case strings.HasPrefix(state, "ready"):
		return models.StateShowtime
	case state == "training":
		return models.StateTraining
	case state == "off" || state == "error":
		return models.StateIdle
	}
	return models.StateUnknown
}

func interpretOverviewTime(timeStr string) time.Duration {
	var duration time.Duration

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
			return time.Duration(0)
		}

		duration += time.Duration(val) * factor
	}

	return duration
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
			status.Uptime.Duration = interpretOverviewTime(data.Lines[0].Time)
		}
	}

	status.NearEndInventory.Vendor = "AVM"
	status.NearEndInventory.Version = data.Version

	status.FarEndInventory.Vendor = data.ExternApText
	status.FarEndInventory.Version = strings.TrimPrefix(data.ExternApValue, "Version ")
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
	if len(farEndInventoryNodes) >= 2 {
		vendor := farEndInventoryNodes[len(farEndInventoryNodes)-2].Data
		vendorStr := strings.TrimSuffix(strings.TrimSpace(vendor), ":")
		version := farEndInventoryNodes[len(farEndInventoryNodes)-1].Data
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
		status.Mode = helpers.ParseMode(data.Lines[0].Mode)

		if status.State == models.StateShowtime {
			status.Uptime.Duration = interpretOverviewTime(data.Lines[0].Time)
		}
	}
}
