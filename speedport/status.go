// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package speedport

import (
	"strconv"
	"strings"

	"3e8.eu/go/dsl/internal/helpers"
	"3e8.eu/go/dsl/models"
)

func interpretStatus(valuesVersion, valuesDSL map[string]responseVar) models.Status {
	var status models.Status

	status.State = interpretState(valuesDSL, "State")
	status.Mode = interpretMode(valuesDSL, "DslOperMode")

	status.NearEndInventory = interpretNearEndInventory(valuesVersion)

	status.DownstreamActualRate.IntValue = interpretIntValue(valuesDSL, "ActualDataDown")
	status.UpstreamActualRate.IntValue = interpretIntValue(valuesDSL, "ActualDataUp")

	status.DownstreamAttainableRate.IntValue = interpretIntValue(valuesDSL, "AttainDataDown")
	status.UpstreamAttainableRate.IntValue = interpretIntValue(valuesDSL, "AttainDataUp")

	status.DownstreamInterleavingDelay.FloatValue = interpretFloatValue(valuesDSL, "InterDelayDown", 1)
	status.UpstreamInterleavingDelay.FloatValue = interpretFloatValue(valuesDSL, "InterDelayUp", 1)

	status.DownstreamAttenuation.FloatValue = interpretFloatValue(valuesDSL, "LineAttenDown", 0.1)
	status.UpstreamAttenuation.FloatValue = interpretFloatValue(valuesDSL, "LineAttenUp", 0.1)

	status.DownstreamSNRMargin.FloatValue = interpretFloatValue(valuesDSL, "SnrMarginDown", 0.1)
	status.UpstreamSNRMargin.FloatValue = interpretFloatValue(valuesDSL, "SnrMarginUp", 0.1)

	status.DownstreamPower.FloatValue = interpretFloatValue(valuesDSL, "SignalLevDown", 0.1)
	status.UpstreamPower.FloatValue = interpretFloatValue(valuesDSL, "SignalLevUp", 0.1)

	status.DownstreamFECCount = interpretIntValue(valuesDSL, "FecErrCDown")
	status.UpstreamFECCount = interpretIntValue(valuesDSL, "FecErrCUp")

	status.DownstreamCRCCount = interpretIntValue(valuesDSL, "CrcErrCDown")
	status.UpstreamCRCCount = interpretIntValue(valuesDSL, "CrcErrCUp")

	return status
}

func interpretState(values map[string]responseVar, key string) (out models.State) {
	if val, ok := values[key]; ok {
		out = helpers.ParseStateTR06X(val.Value)
	}
	return
}

func interpretMode(values map[string]responseVar, key string) (out models.Mode) {
	if val, ok := values[key]; ok {
		if val.Value == "VDSL" {
			out.Type = models.ModeTypeVDSL2
			return
		}
		out = helpers.ParseMode(val.Value)
	}
	return
}

func interpretNearEndInventory(values map[string]responseVar) (out models.Inventory) {
	out.Vendor = "Speedport"

	if val, ok := values["Xdsl"]; ok {
		out.Version = val.Value
	}

	if strings.HasPrefix(out.Version, "B2pv") {
		out.Vendor = "Broadcom"
	}

	return
}

func interpretIntValue(values map[string]responseVar, key string) (out models.IntValue) {
	if val, ok := values[key]; ok {
		if valInt, err := strconv.ParseInt(val.Value, 10, 64); err == nil {
			out.Int = valInt
			out.Valid = true
		}
	}
	return
}

func interpretFloatValue(values map[string]responseVar, key string, factor float64) (out models.FloatValue) {
	if val, ok := values[key]; ok {
		if valFloat, err := strconv.ParseFloat(val.Value, 64); err == nil {
			out.Float = valFloat * factor
			out.Valid = true
		}
	}
	return
}
