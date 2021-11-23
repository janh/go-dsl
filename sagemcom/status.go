// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package sagemcom

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"3e8.eu/go/dsl/internal/helpers"
	"3e8.eu/go/dsl/models"
)

var regexpMode = regexp.MustCompile(`(?i)^G[_\.]([0-9]{3})[_\.]([0-9])_annex_(b)$`)

const regexpModeReplacement = `G.$1.$2 Annex $3`

func interpretStatus(data *dslObj) models.Status {
	var status models.Status

	interpretStatusState(&status, data)
	interpretStatusMode(&status, data)
	interpretStatusUptime(&status, data)
	interpretStatusInventory(&status, data)
	interpretStatusRates(&status, data)
	interpretStatusImpulseNoise(&status, data)
	interpretStatusVectoring(&status, data)
	interpretStatusSignal(&status, data)
	interpretStatusCounters(&status, data)

	return status
}

func interpretStatusState(status *models.Status, data *dslObj) {
	linkStatus := data.Lines[0].LinkStatus
	if linkStatus == "initializing" {
		status.State = models.StateIdle
	} else {
		status.State = models.ParseState(linkStatus)
	}
}

func interpretStatusMode(status *models.Status, data *dslObj) {
	mode := data.Lines[0].StandardUsed
	mode = regexpMode.ReplaceAllString(mode, regexpModeReplacement)
	status.Mode = models.ParseMode(mode)
}

func interpretStatusUptime(status *models.Status, data *dslObj) {
	status.Uptime.Duration = time.Duration(data.Lines[0].Stats.ShowtimeStart.Int) * time.Second
}

func interpretStatusInventory(status *models.Status, data *dslObj) {
	if len(data.Lines[0].IDDSLAM) == 11 {
		status.FarEndInventory.Vendor = helpers.FormatVendor(data.Lines[0].IDDSLAM[0:4])
		version := helpers.ParseHexadecimal(data.Lines[0].IDDSLAM[7:11])
		status.FarEndInventory.Version = helpers.FormatVersion(status.FarEndInventory.Vendor, version)
	} else if len(data.Lines[0].XTUCVendor) == 8 {
		vendor := helpers.ParseHexadecimal(data.Lines[0].XTUCVendor)
		status.FarEndInventory.Vendor = helpers.FormatVendor(string(vendor))
	}

	if len(data.Lines[0].XTURVendor) == 8 {
		vendor := helpers.ParseHexadecimal(data.Lines[0].XTURVendor)
		status.NearEndInventory.Vendor = helpers.FormatVendor(string(vendor))
	} else if data.Lines[0].ModemChip != "" {
		status.NearEndInventory.Vendor = data.Lines[0].ModemChip
	} else if strings.HasPrefix(data.Lines[0].FirmwareVersion, "A2pv") ||
		strings.HasPrefix(data.Lines[0].FirmwareVersion, "B2pv") {
		status.NearEndInventory.Vendor = "Broadcom"
	} else {
		status.NearEndInventory.Vendor = "Sagemcom"
	}

	status.NearEndInventory.Version = data.Lines[0].FirmwareVersion
}

func interpretStatusRates(status *models.Status, data *dslObj) {
	// this is not correct when G.INP is enabled, as these values report the "Actual Data Rate" instead of the
	// "Actual Net Data Rate", but the latter doesn't seem to be exposed at all, so this is the best we can have
	status.UpstreamActualRate.IntValue = data.Channels[0].UpstreamCurrRate
	status.DownstreamActualRate.IntValue = data.Channels[0].DownstreamCurrRate

	status.UpstreamAttainableRate.IntValue = data.Lines[0].UpstreamMaxBitRate
	status.DownstreamAttainableRate.IntValue = data.Lines[0].DownstreamMaxBitRate
}

func interpretStatusImpulseNoise(status *models.Status, data *dslObj) {
	status.UpstreamInterleavingDelay.FloatValue = convertFloatValue(data.Channels[0].ActualInterleavingDelayus, 0.01)
	status.DownstreamInterleavingDelay.FloatValue = convertFloatValue(data.Channels[0].ActualInterleavingDelay, 0.01)

	status.UpstreamImpulseNoiseProtection.FloatValue = convertFloatValue(data.Channels[0].ACTINPus, 0.1)
	status.DownstreamImpulseNoiseProtection.FloatValue = convertFloatValue(data.Channels[0].ACTINP, 0.1)
}

func interpretStatusVectoring(status *models.Status, data *dslObj) {
	vectoring := strings.ToLower(data.Lines[0].VectoringState)
	if vectoring == "running" {
		status.DownstreamVectoringState.State = models.VectoringStateFull
		status.DownstreamVectoringState.Valid = true
	} else if vectoring == "disabled" {
		status.DownstreamVectoringState.State = models.VectoringStateOff
		status.DownstreamVectoringState.Valid = true
	}
}

func interpretStatusSignal(status *models.Status, data *dslObj) {
	status.UpstreamAttenuation.FloatValue = convertFloatValue(data.Lines[0].UpstreamAttenuation, 0.1)
	status.DownstreamAttenuation.FloatValue = convertFloatValue(data.Lines[0].DownstreamAttenuation, 0.1)

	if data.Lines[0].ModemChip == "Lantiq" {
		// At least on Speedport Pro with firmware 4.5, the downstream and upstream
		// attenuation values are swapped. In addition, it seems that the Attenuation
		// and SignalAttenuation values are also swapped. However, the LATN values in
		// the TestParams object seem to be correct (those should actually report the
		// per-band attenuation, but the Lantiq FAPI reports total values instead).

		upstreamLATN := parseIntValue(data.Lines[0].TestParams.LATNus)
		downstreamLATN := parseIntValue(data.Lines[0].TestParams.LATNds)

		if upstreamLATN.Valid && downstreamLATN.Valid {
			status.UpstreamAttenuation.FloatValue = convertFloatValue(upstreamLATN, 0.1)
			status.DownstreamAttenuation.FloatValue = convertFloatValue(downstreamLATN, 0.1)
		}
	}

	status.UpstreamSNRMargin.FloatValue = convertFloatValue(data.Lines[0].UpstreamNoiseMargin, 0.1)
	status.DownstreamSNRMargin.FloatValue = convertFloatValue(data.Lines[0].DownstreamNoiseMargin, 0.1)

	status.UpstreamPower.FloatValue = convertFloatValue(data.Lines[0].UpstreamPower, 0.1)
	status.DownstreamPower.FloatValue = convertFloatValue(data.Lines[0].DownstreamPower, 0.1)
}

func interpretStatusCounters(status *models.Status, data *dslObj) {
	status.UpstreamFECCount = data.Channels[0].Stats.Showtime.XTUCFECErrors
	status.DownstreamFECCount = data.Channels[0].Stats.Showtime.XTURFECErrors

	status.UpstreamCRCCount = data.Channels[0].Stats.Showtime.XTUCCRCErrors
	status.DownstreamCRCCount = data.Channels[0].Stats.Showtime.XTURCRCErrors

	status.UpstreamESCount = data.Lines[0].Stats.Showtime.TxErroredSecs
	status.DownstreamESCount = data.Lines[0].Stats.Showtime.ErroredSecs

	status.UpstreamSESCount = data.Lines[0].Stats.Showtime.TxSeverelyErroredSecs
	status.DownstreamSESCount = data.Lines[0].Stats.Showtime.SeverelyErroredSecs
}

func parseIntValue(data string) (out models.IntValue) {
	if valInt, err := strconv.ParseInt(data, 10, 64); err == nil {
		out.Int = valInt
		out.Valid = true
	}
	return
}

func convertFloatValue(val models.IntValue, factor float64) (out models.FloatValue) {
	out.Float = float64(val.Int) * factor
	out.Valid = val.Valid
	return
}
