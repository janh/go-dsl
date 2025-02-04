// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package draytek_v5

import (
	"encoding/json"
	"fmt"
	"strings"

	"3e8.eu/go/dsl/internal/exec"
	"3e8.eu/go/dsl/models"
)

type response struct {
	RID   string        `json:"rid"`
	Total int           `json:"total"`
	Start int           `json:"start"`
	CT    []responseMap `json:"ct"`
}

type responseMap map[string]responseMapItem

type responseMapItem []json.RawMessage

func fetchData(e exec.Executor, config string, name string) (raw string, data json.RawMessage, err error) {
	res, err := e.Execute("config " + config)
	if err != nil {
		return
	}
	if strings.TrimSpace(res) != "" {
		err = fmt.Errorf("unexpected config result: %s", res)
		return
	}

	raw, err = e.Execute("show all")
	if err != nil {
		return
	}

	res, err = e.Execute("exit")
	if err != nil {
		return
	}
	if strings.TrimSpace(res) != "" {
		err = fmt.Errorf("unexpected exit result: %s", res)
		return
	}

	var resp response
	err = json.Unmarshal([]byte(raw), &resp)
	if err != nil {
		return
	}

	if len(resp.CT) != 1 {
		err = fmt.Errorf("unexpected length of \"ct\" field: %d", len(resp.CT))
		return
	}

	item, ok := resp.CT[0][name]
	if !ok {
		err = fmt.Errorf("expected item \"%s\" not found", name)
	}

	if len(item) != 1 {
		err = fmt.Errorf("unexpected item data length: %d", len(item))
	}

	data = item[0]

	return
}

func updateData(e exec.Executor) (status models.Status, bins models.Bins, rawData []byte, err error) {
	rawGeneral, dataGeneral, err := fetchData(e,
		"Monitoring DSL_Status Monitoring_DSL_General", "0MONITORING_DSL_GENERAL")
	if err != nil {
		err = fmt.Errorf("failed to fetch general data: %w", err)
		return
	}

	rawTone, dataTone, err := fetchData(e,
		"Monitoring DSL_Status Monitoring_DSL_Tone", "1MONITORING_DSL_TONE")
	if err != nil {
		err = fmt.Errorf("failed to fetch tone data: %w", err)
		return
	}

	status, err = parseStatus(dataGeneral)
	if err != nil {
		err = fmt.Errorf("failed to parse status data: %w", err)
		return
	}

	bins, err = parseBins(status, dataTone)
	if err != nil {
		err = fmt.Errorf("failed to parse bins data: %w", err)
		return
	}

	var b strings.Builder
	fmt.Fprintln(&b, "> config Monitoring DSL_Status Monitoring_DSL_General")
	fmt.Fprintln(&b, "> show all")
	fmt.Fprintln(&b, rawGeneral)
	fmt.Fprintln(&b, "> config Monitoring DSL_Status Monitoring_DSL_Tone")
	fmt.Fprintln(&b, "> show all")
	fmt.Fprintln(&b, rawTone)
	fmt.Fprintln(&b)
	rawData = []byte(b.String())

	return
}
