// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package draytek_v5

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"3e8.eu/go/dsl/internal/helpers"
	"3e8.eu/go/dsl/models"
)

var regexpFilterCharacters = regexp.MustCompile(`[^a-zA-Z0-9]+`)
var regexpVendorID = regexp.MustCompile(`([[:xdigit:]]{8})\s([[:xdigit:]]{8})`)

type statusInfo struct {
	Status       string            `json:"Status"`
	Mode         string            `json:"Mode"`
	Profile      string            `json:"Profile"`
	Annex        string            `json:"Annex"`
	DSLVersion   string            `json:"DSL_Version"`
	ATURVendorID string            `json:"ATU_R_Vendor_ID"`
	ATUCVendorID string            `json:"ATU_C_Vendor_ID"`
	StreamTable  []streamTableItem `json:"Stream_Table"`
	EndTable     []endTableItem    `json:"End_Table"`
}

type streamTableItem struct {
	Name       string `json:"Name"`
	Downstream string `json:"Downstream"`
	Upstream   string `json:"Upstream"`
}

type endTableItem struct {
	Name    string `json:"Name"`
	NearEnd string `json:"Near_End"`
	FarEnd  string `json:"Far_End"`
}

func parseStatus(data json.RawMessage) (models.Status, error) {
	var status models.Status

	var info statusInfo
	err := json.Unmarshal(data, &info)
	if err != nil {
		return status, err
	}

	status.State = parseState(info.Status)
	status.Mode = parseMode(info.Mode, info.Profile, info.Annex)

	status.NearEndInventory = parseInventory(info.ATURVendorID)
	status.NearEndInventory.Version = info.DSLVersion
	status.FarEndInventory = parseInventory(info.ATUCVendorID)

	parseStreamTable(&status, info.StreamTable)
	parseEndTable(&status, info.EndTable)

	return status, nil
}

func parseState(str string) models.State {
	switch strings.ToLower(str) {
	case "idle", "down":
		return models.StateDown
	case "training":
		return models.StateInit
	case "showtime":
		return models.StateShowtime
	}
	return models.StateUnknown
}

func parseMode(mode, profile, annex string) models.Mode {
	return helpers.ParseMode(mode + " " + profile + " " + annex)
}

func parseInventory(vendorid string) models.Inventory {
	out := models.Inventory{}

	match := regexpVendorID.FindStringSubmatch(vendorid)
	if match == nil {
		return out
	}

	v0 := helpers.ParseHexadecimal(match[1])
	v1 := helpers.ParseHexadecimal(match[2])

	vendor := []byte{v0[2], v0[3], v1[0], v1[1]}
	out.Vendor = helpers.FormatVendor(string(vendor))

	out.Version = helpers.FormatVersion(out.Vendor, v1[2:4])

	return out
}

func parseIntValue(val string) (out models.IntValue) {
	if valInt, err := strconv.ParseInt(val, 10, 64); err == nil {
		out.Int = valInt
		out.Valid = true
	}

	return
}

func parseIntValueSuffix(val string, suffix string) (out models.IntValue) {
	val = strings.ToLower(val)
	suffix = strings.ToLower(suffix)

	if strings.HasSuffix(val, suffix) {
		val := val[:len(val)-len(suffix)]
		if valInt, err := strconv.ParseInt(val, 10, 64); err == nil {
			out.Int = valInt
			out.Valid = true
		}
	}

	return
}

func parseFloatValueSuffix(val string, suffix string) (out models.FloatValue) {
	val = strings.ToLower(val)
	suffix = strings.ToLower(suffix)

	if strings.HasSuffix(val, suffix) {
		val := val[:len(val)-len(suffix)]
		if valFloat, err := strconv.ParseFloat(val, 64); err == nil {
			out.Float = valFloat
			out.Valid = true
		}
	}

	return
}

func parseBoolValue(val string) (out models.BoolValue) {
	val = strings.ToLower(val)

	if val == "on" {
		out.Bool = true
		out.Valid = true
	} else if val == "off" {
		out.Bool = false
		out.Valid = true
	}

	return
}

func parseInterleavingDelay(pathMode string) (out models.FloatValue) {
	pathMode = strings.ToLower(pathMode)

	if pathMode == "fast" {
		out.Float = 0
		out.Valid = true
	}

	return
}

func parseStreamTable(status *models.Status, table []streamTableItem) {
	for _, item := range table {
		key := strings.ToLower(regexpFilterCharacters.ReplaceAllString(item.Name, ""))

		switch key {

		case "actualrate":
			status.DownstreamActualRate.IntValue = parseIntValueSuffix(item.Downstream, " kbps")
			status.UpstreamActualRate.IntValue = parseIntValueSuffix(item.Upstream, " kbps")

		case "attainablerate":
			status.DownstreamAttainableRate.IntValue = parseIntValueSuffix(item.Downstream, " Kbps")
			status.UpstreamAttainableRate.IntValue = parseIntValueSuffix(item.Upstream, " Kbps")

		case "pathmode":
			status.DownstreamInterleavingDelay.FloatValue = parseInterleavingDelay(item.Downstream)
			status.UpstreamInterleavingDelay.FloatValue = parseInterleavingDelay(item.Upstream)

		case "actualpsd":
			status.DownstreamPower.FloatValue = parseFloatValueSuffix(item.Downstream, " dB")
			status.UpstreamPower.FloatValue = parseFloatValueSuffix(item.Upstream, " dB")

		case "snrmargin":
			status.DownstreamSNRMargin.FloatValue = parseFloatValueSuffix(item.Downstream, " dB")
			status.UpstreamSNRMargin.FloatValue = parseFloatValueSuffix(item.Upstream, " dB")

		}
	}
}

func parseEndTable(status *models.Status, table []endTableItem) {
	for _, item := range table {
		key := strings.ToLower(regexpFilterCharacters.ReplaceAllString(item.Name, ""))

		switch key {

		case "bitswap":
			status.DownstreamBitswap.Enabled = parseBoolValue(item.NearEnd)
			status.UpstreamBitswap.Enabled = parseBoolValue(item.FarEnd)

		case "retx":
			status.DownstreamRetransmissionEnabled = parseBoolValue(item.NearEnd)
			status.UpstreamRetransmissionEnabled = parseBoolValue(item.FarEnd)

		case "attenuation":
			status.DownstreamAttenuation.FloatValue = parseFloatValueSuffix(item.NearEnd, " dB")
			status.UpstreamAttenuation.FloatValue = parseFloatValueSuffix(item.FarEnd, " dB")

		case "crc":
			status.DownstreamCRCCount = parseIntValue(item.NearEnd)
			status.UpstreamCRCCount = parseIntValue(item.FarEnd)

		case "es":
			status.DownstreamESCount = parseIntValueSuffix(item.NearEnd, " s")
			status.UpstreamESCount = parseIntValueSuffix(item.FarEnd, " s")

		case "ses":
			status.DownstreamSESCount = parseIntValueSuffix(item.NearEnd, " s")
			status.UpstreamSESCount = parseIntValueSuffix(item.FarEnd, " s")

		}
	}
}
