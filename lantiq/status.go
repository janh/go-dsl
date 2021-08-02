// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lantiq

import (
	"fmt"
	"strconv"
	"strings"

	"3e8.eu/go/dsl/internal/helpers"
	"3e8.eu/go/dsl/models"
)

func parseBasicStatus(data *data) models.Status {
	var status models.Status

	parseStatusState(&status, data.LineState)
	parseStatusMode(&status, data.G997_XTUSystemEnablingStatus, data.BandPlanSTatus)
	parseStatusInventory(&status, data.VersionInformation, data.G997_LineInventory_Far)

	return status
}

func parseExtendedStatus(status *models.Status, bins *models.Bins, data *data) {
	parseStatusChannelStatus(status, data.G997_ChannelStatus_US, data.G997_ChannelStatus_DS)
	parseStatusLineStatus(status, bins, data.G997_LineStatus_US, data.G997_LineStatus_DS)
	parseStatusLineFeatures(status, data.LineFeatureStatus_US, data.LineFeatureStatus_DS)

	parseStatusChannelCounters(status, data.PM_ChannelCountersShowtime_Near, data.PM_ChannelCountersShowtime_Far)
	parseStatusLineSecCounters(status, data.PM_LineSecCountersShowtime_Near, data.PM_LineSecCountersShowtime_Far)
	parseStatusReTxStatistics(status, data.ReTxStatistics_Near, data.ReTxStatistics_Far)
}

func interpretStatusBoolValue(values map[string]string, key string) (out models.BoolValue) {
	if val, ok := values[key]; ok {
		if valInt, err := strconv.Atoi(val); err == nil && (valInt == 0 || valInt == 1) {
			out.Bool = valInt == 1
			out.Valid = true
		}
	}
	return
}

func interpretStatusIntValue(values map[string]string, key string, factor int64) (out models.IntValue) {
	if val, ok := values[key]; ok {
		if valInt, err := strconv.ParseInt(val, 10, 64); err == nil {
			out.Int = valInt / factor
			out.Valid = true
		}
	}
	return
}

func interpretStatusFloatValue(values map[string]string, key string, factor float64) (out models.FloatValue) {
	if val, ok := values[key]; ok {
		if valInt, err := strconv.ParseFloat(val, 64); err == nil {
			out.Float = float64(valInt) / factor
			out.Valid = true
		}
	}
	return
}

func interpretStatusFloatValueWeightedPositiveAverage(values map[string]string, keys []string, weights []float64, factor float64) (out models.FloatValue) {
	val := make([]models.FloatValue, len(keys))
	for i, key := range keys {
		val[i] = interpretStatusFloatValue(values, key, factor)
	}

	maxCount := len(values)
	if len(weights) < maxCount {
		maxCount = len(weights)
	}

	var weightSum float64

	for i := 0; i < maxCount; i++ {
		if val[i].Valid && val[i].Float > 0 {
			out.Float += val[i].Float * weights[i]
			weightSum += weights[i]
		}
	}

	if weightSum > 0 {
		out.Float /= weightSum
		out.Valid = true
	}

	return
}

func interpretStatusBytes(values map[string]string, key string) []byte {
	if val, ok := values[key]; ok {
		if len(val) < 2 || val[0] != '(' || val[len(val)-1] != ')' {
			return nil
		}
		val = val[1 : len(val)-1]

		valSplit := strings.Split(val, ",")
		out := make([]byte, len(valSplit))

		for i, item := range valSplit {
			byteVal, err := strconv.ParseUint(item, 16, 8)
			if err != nil {
				return nil
			}
			out[i] = byte(byteVal)
		}

		return out
	}

	return nil
}

func interpretStatusByte(values map[string]string, key string) byte {
	valStr := values[key]
	valInt, _ := strconv.ParseUint(valStr, 10, 8)
	return byte(valInt)
}

func parseStatusState(status *models.Status, lsg string) {
	lsgValues := parseValues(lsg)

	lineStateStr := lsgValues["nLineState"]
	if strings.HasPrefix(lineStateStr, "0x") {
		lineStateStr = lineStateStr[2:]
	}
	lineState, err := strconv.ParseUint(lineStateStr, 16, 64)
	if err != nil {
		return
	}

	lineStateMap := map[uint64]string{
		0x0:   "not initialized",
		0x1:   "exception",
		0x10:  "not updated",
		0xff:  "idle request",
		0x100: "idle",
		0x1ff: "silent request",
		0x200: "silent",
		0x300: "handshake",
		0x310: "bonding clr",
		0x380: "full init",
		0x3c0: "short init",
		0x400: "discovery",
		0x500: "training",
		0x600: "analysis",
		0x700: "exchange",
		0x800: "showtime no sync",
		0x801: "showtime tc sync",
		0x900: "fastretrain",
		0xa00: "lowpower l2",
		0xb00: "loopdiagnostic active",
		0xb10: "loopdiagnostic data exchange",
		0xb20: "loopdiagnostic data request",
		0xc00: "loopdiagnostic complete",
		0xd00: "resync",
	}

	if lineStateString, ok := lineStateMap[lineState]; ok {
		status.State = models.ParseState(lineStateString)
	}
}

func parseStatusMode(status *models.Status, g997xtusesg, bpstg string) {
	g977xtusesgValues := parseValues(g997xtusesg)
	xtse1 := interpretStatusByte(g977xtusesgValues, "XTSE1")
	xtse2 := interpretStatusByte(g977xtusesgValues, "XTSE2")
	xtse3 := interpretStatusByte(g977xtusesgValues, "XTSE3")
	xtse4 := interpretStatusByte(g977xtusesgValues, "XTSE4")
	xtse5 := interpretStatusByte(g977xtusesgValues, "XTSE5")
	xtse6 := interpretStatusByte(g977xtusesgValues, "XTSE6")
	xtse7 := interpretStatusByte(g977xtusesgValues, "XTSE7")
	xtse8 := interpretStatusByte(g977xtusesgValues, "XTSE8")

	status.Mode.Type = getStatusModeType(xtse1, xtse2, xtse3, xtse4, xtse5, xtse6, xtse7, xtse8)

	if status.Mode.Type == models.ModeTypeVDSL2 || (status.Mode.Type == models.ModeTypeUnknown && status.Mode.Subtype == models.ModeSubtypeUnknown) {

		bpstgValues := parseValues(bpstg)
		nProfile, err := strconv.Atoi(bpstgValues["nProfile"])
		if err != nil {
			return
		}

		profiles := []models.ModeSubtype{
			models.ModeSubtypeProfile8a,
			models.ModeSubtypeProfile8b,
			models.ModeSubtypeProfile8c,
			models.ModeSubtypeProfile8d,
			models.ModeSubtypeProfile12a,
			models.ModeSubtypeProfile12b,
			models.ModeSubtypeProfile17a,
			models.ModeSubtypeProfile30a,
			models.ModeSubtypeProfile35b,
		}

		if nProfile >= 0 && nProfile < len(profiles) {
			status.Mode.Type = models.ModeTypeVDSL2
			status.Mode.Subtype = profiles[nProfile]
		}

	} else {

		status.Mode.Subtype = getStatusModeSubtype(xtse1, xtse2, xtse3, xtse4, xtse5, xtse6, xtse7, xtse8)

	}
}

func parseStatusInventory(status *models.Status, vig, g997ligFar string) {
	vigValues := parseValues(vig)
	status.NearEndInventory.Vendor = "Infineon"
	status.NearEndInventory.Version = vigValues["DSL_ChipSetFWVersion"]

	g997ligFarValues := parseValues(g997ligFar)
	vendorID := interpretStatusBytes(g997ligFarValues, "G994VendorID")
	if len(vendorID) == 8 {
		status.FarEndInventory.Vendor = helpers.FormatVendor(string(vendorID[2:6]))
		status.FarEndInventory.Version = fmt.Sprintf("%d.%d", vendorID[6], vendorID[7])
	}
}

func parseStatusChannelStatus(status *models.Status, g997csgUS, g997csgDS string) {
	g997csgUSValues := parseValues(g997csgUS)
	g997csgDSValues := parseValues(g997csgDS)

	status.UpstreamActualRate.IntValue = interpretStatusIntValue(g997csgUSValues, "ActualNetDataRate", 1000)
	if !status.UpstreamActualRate.Valid {
		status.UpstreamActualRate.IntValue = interpretStatusIntValue(g997csgUSValues, "ActualDataRate", 1000)
	}
	status.DownstreamActualRate.IntValue = interpretStatusIntValue(g997csgDSValues, "ActualNetDataRate", 1000)
	if !status.DownstreamActualRate.Valid {
		status.DownstreamActualRate.IntValue = interpretStatusIntValue(g997csgDSValues, "ActualDataRate", 1000)
	}

	status.UpstreamInterleavingDelay.FloatValue = interpretStatusFloatValue(g997csgUSValues, "ActualInterleaveDelay", 100)
	status.DownstreamInterleavingDelay.FloatValue = interpretStatusFloatValue(g997csgDSValues, "ActualInterleaveDelay", 100)
}

func getBandWeights(bands []models.Band) []float64 {
	if len(bands) == 0 || len(bands) > 5 {
		return []float64{1, 1, 1, 1, 1}
	}

	weights := make([]float64, 5, 5)

	for i, b := range bands {
		weights[i] = float64(b.End - b.Start)
	}

	return weights
}

func parseStatusLineStatus(status *models.Status, bins *models.Bins, g997lsgUS, g997lsgDS string) {
	g997lsgUSValues := parseValues(g997lsgUS)
	g997lsgDSValues := parseValues(g997lsgDS)

	status.UpstreamAttainableRate.IntValue = interpretStatusIntValue(g997lsgUSValues, "ATTNDR", 1000)
	status.DownstreamAttainableRate.IntValue = interpretStatusIntValue(g997lsgDSValues, "ATTNDR", 1000)

	powerUS := interpretStatusFloatValue(g997lsgUSValues, "ACTATP", 10)
	powerDS := interpretStatusFloatValue(g997lsgDSValues, "ACTATP", 10)
	// upstream and downstream power seem to be swapped
	if status.Mode.Type == models.ModeTypeVDSL2 && powerUS.Float > powerDS.Float {
		status.DownstreamPower.FloatValue = powerUS
		status.UpstreamPower.FloatValue = powerDS
	} else {
		status.DownstreamPower.FloatValue = powerDS
		status.UpstreamPower.FloatValue = powerUS
	}

	weightsUS := getBandWeights(bins.Bands.Upstream)
	weightsDS := getBandWeights(bins.Bands.Downstream)

	status.UpstreamSNRMargin.FloatValue = interpretStatusFloatValue(g997lsgUSValues, "SNR", 10)
	if !status.UpstreamSNRMargin.FloatValue.Valid {
		status.UpstreamSNRMargin.FloatValue = interpretStatusFloatValueWeightedPositiveAverage(g997lsgUSValues,
			[]string{"SNR[0]", "SNR[1]", "SNR[2]", "SNR[3]", "SNR[4]"}, weightsUS, 10)
	}
	status.DownstreamSNRMargin.FloatValue = interpretStatusFloatValue(g997lsgDSValues, "SNR", 10)
	if !status.DownstreamSNRMargin.FloatValue.Valid {
		status.DownstreamSNRMargin.FloatValue = interpretStatusFloatValueWeightedPositiveAverage(g997lsgDSValues,
			[]string{"SNR[0]", "SNR[1]", "SNR[2]", "SNR[3]", "SNR[4]"}, weightsDS, 10)
	}

	status.UpstreamAttenuation.FloatValue = interpretStatusFloatValue(g997lsgUSValues, "LATN", 10)
	if !status.UpstreamAttenuation.FloatValue.Valid {
		status.UpstreamAttenuation.FloatValue = interpretStatusFloatValueWeightedPositiveAverage(g997lsgUSValues,
			[]string{"LATN[0]", "LATN[1]", "LATN[2]", "LATN[3]", "LATN[4]"}, weightsUS, 10)
	}
	status.DownstreamAttenuation.FloatValue = interpretStatusFloatValue(g997lsgDSValues, "LATN", 10)
	if !status.DownstreamAttenuation.FloatValue.Valid {
		status.DownstreamAttenuation.FloatValue = interpretStatusFloatValueWeightedPositiveAverage(g997lsgDSValues,
			[]string{"LATN[0]", "LATN[1]", "LATN[2]", "LATN[3]", "LATN[4]"}, weightsDS, 10)
	}
}

func parseStatusLineFeatures(status *models.Status, lfsgUS, lfsgDS string) {
	lfsgUSValues := parseValues(lfsgUS)
	lfsgDSValues := parseValues(lfsgDS)

	status.UpstreamRetransmissionEnabled = interpretStatusBoolValue(lfsgUSValues, "bReTxEnable")
	status.DownstreamRetransmissionEnabled = interpretStatusBoolValue(lfsgDSValues, "bReTxEnable")
}

func parseStatusChannelCounters(status *models.Status, pmccsgNear, pmccsgFar string) {
	pmccsgNearValues := parseValues(pmccsgNear)
	pmccsgFarValues := parseValues(pmccsgFar)

	status.UpstreamFECCount = interpretStatusIntValue(pmccsgFarValues, "nFEC", 1)
	status.DownstreamFECCount = interpretStatusIntValue(pmccsgNearValues, "nFEC", 1)

	status.UpstreamCRCCount = interpretStatusIntValue(pmccsgFarValues, "nCodeViolations", 1)
	status.DownstreamCRCCount = interpretStatusIntValue(pmccsgNearValues, "nCodeViolations", 1)
}

func parseStatusLineSecCounters(status *models.Status, pmlscsgNear, pmlscsgFar string) {
	pmlscsgNearValues := parseValues(pmlscsgNear)
	pmlscsgFarValues := parseValues(pmlscsgFar)

	status.UpstreamESCount = interpretStatusIntValue(pmlscsgFarValues, "nES", 1)
	status.DownstreamESCount = interpretStatusIntValue(pmlscsgNearValues, "nES", 1)
}

func parseStatusReTxStatistics(status *models.Status, rtsgNear, rtsgFar string) {
	rtsgNearValues := parseValues(rtsgNear)
	rtsgFarValues := parseValues(rtsgFar)

	status.UpstreamRTXTXCount = interpretStatusIntValue(rtsgFarValues, "nTxRetransmitted", 1)
	status.DownstreamRTXTXCount = interpretStatusIntValue(rtsgNearValues, "nTxRetransmitted", 1)

	status.UpstreamRTXCCount = interpretStatusIntValue(rtsgFarValues, "nRxCorrected", 1)
	status.DownstreamRTXCCount = interpretStatusIntValue(rtsgNearValues, "nRxCorrected", 1)

	status.UpstreamRTXUCCount = interpretStatusIntValue(rtsgFarValues, "nRxUncorrectedProtected", 1)
	status.DownstreamRTXUCCount = interpretStatusIntValue(rtsgNearValues, "nRxUncorrectedProtected", 1)
}
