// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package draytek_v5

import (
	"encoding/json"
	"fmt"

	"3e8.eu/go/dsl/internal/helpers"
	"3e8.eu/go/dsl/models"
)

type toneInfo struct {
	ToneTable []toneTable `json:"Tone_Table"`
}

type toneTable struct {
	Name string `json:"Name"`
	Data string `json:"data"`
}

type toneItem struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func parseBins(status models.Status, data json.RawMessage) (models.Bins, error) {
	var bins models.Bins

	bins.Mode = status.Mode

	var info toneInfo
	err := json.Unmarshal(data, &info)
	if err != nil {
		return bins, err
	}

	for _, table := range info.ToneTable {
		switch table.Name {
		case "down":
			err = parseBits(&bins.Bits.Downstream, table.Data, bins.Mode.BinCount())
			if err != nil {
				return bins, err
			}
		case "up":
			err = parseBits(&bins.Bits.Upstream, table.Data, bins.Mode.BinCount())
			if err != nil {
				return bins, err
			}
		default:
			err = fmt.Errorf("unexpected tone table name: %s", table.Name)
			return bins, err
		}
	}

	helpers.GenerateBandsData(&bins)

	return bins, nil
}

func parseBits(binsBits *models.BinsBits, data string, binCount int) error {
	var items []toneItem
	err := json.Unmarshal([]byte(data), &items)
	if err != nil {
		return err
	}

	binsBits.Data = make([]int8, binCount)

	for _, item := range items {
		if item.X < 0 || item.X >= binCount {
			continue
		}

		binsBits.Data[item.X] = int8(item.Y)
	}

	return nil
}
