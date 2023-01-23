// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lantiq

import (
	"strconv"
	"strings"
	"time"

	"3e8.eu/go/dsl/internal/helpers"
	"3e8.eu/go/dsl/models"
)

func isCarrierOffice(inventory models.Inventory) bool {
	switch {

	case strings.HasPrefix(inventory.Version, "10."), strings.HasPrefix(inventory.Version, "11."):
		// Platforms 10/11 are known to be used in DSLAMs (probably VINAX-CO)
		return true

	case strings.HasPrefix(inventory.Version, "9.7."):
		// Platform 9 is VINAX which can run in both CPE and CO mode
		// Feature set 7 has been seen only in CO firmware of NV-600L/ALL126AM2
		return true

	case strings.HasPrefix(inventory.Version, "14."):
		// Platform 14 seems to be VRX618 in CO mode
		return true

	}

	return false
}

func parseBasicStatus(data *data) models.Status {
	var status models.Status

	parseStatusState(&status, data.LineState)
	parseStatusMode(&status, data.G997_XTUSystemEnablingStatus, data.BandPlanSTatus)
	parseStatusInventory(&status, data.VersionInformation, data.G997_LineInventory_Far)

	return status
}

func parseExtendedStatus(status *models.Status, bins *models.Bins, data *data) {
	parseStatusChannelStatus(status, data.G997_ChannelStatus_US, data.G997_ChannelStatus_DS, data.APIVersion)
	parseStatusLineStatus(status, bins, data.G997_LineStatus_US, data.G997_LineStatus_DS)
	parseStatusLineFeatures(status, data.LineFeatureStatus_US, data.LineFeatureStatus_DS)
	parseStatusRateAdaptationStatus(status, data.G997_RateAdaptationStatus_US, data.G997_RateAdaptationStatus_DS)
	parseStatusOlrStatistics(status, data.OlrStatistics_US, data.OlrStatistics_DS)
	parseStatusDSMStatus(status, data.DSM_Status)

	normalizeOLRValues(status)

	if isCarrierOffice(status.NearEndInventory) {
		parseStatusChannelCounters(status, data.PM_ChannelCountersShowtime_Far, data.PM_ChannelCountersShowtime_Near)
		parseStatusLineSecCounters(status, data.PM_LineSecCountersShowtime_Far, data.PM_LineSecCountersShowtime_Near)
		parseStatusReTxCounters(status, data.PM_ReTxCountersShowtimeGet_Far, data.PM_ReTxCountersShowtimeGet_Near)
		parseStatusReTxStatistics(status, data.ReTxStatistics_Far, data.ReTxStatistics_Near)
	} else {
		parseStatusChannelCounters(status, data.PM_ChannelCountersShowtime_Near, data.PM_ChannelCountersShowtime_Far)
		parseStatusLineSecCounters(status, data.PM_LineSecCountersShowtime_Near, data.PM_LineSecCountersShowtime_Far)
		parseStatusReTxCounters(status, data.PM_ReTxCountersShowtimeGet_Near, data.PM_ReTxCountersShowtimeGet_Far)
		parseStatusReTxStatistics(status, data.ReTxStatistics_Near, data.ReTxStatistics_Far)
	}
}

func normalizeOLRValues(status *models.Status) {
	status.UpstreamBitswap.Normalize()
	status.DownstreamBitswap.Normalize()

	status.UpstreamSeamlessRateAdaptation.Normalize()
	status.DownstreamSeamlessRateAdaptation.Normalize()
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

func interpretStatusDuration(values map[string]string, key string) (out models.Duration) {
	if val, ok := values[key]; ok {
		if valInt, err := strconv.ParseInt(val, 10, 64); err == nil {
			out.Duration = time.Duration(valInt) * time.Second
		}
	}
	return
}

func interpretStatusEFTRMin(values map[string]string, key string) (out models.IntValue) {
	if val, ok := values[key]; ok {
		if valInt, err := strconv.ParseInt(val, 10, 64); err == nil && valInt > 0 && valInt != 4294967295 {
			out.Int = valInt / 1000
			out.Valid = true
		}
	}
	return
}

func interpretStatusRAMode(values map[string]string, key string) (out models.BoolValue) {
	if val, ok := values[key]; ok {
		if valInt, err := strconv.ParseInt(val, 10, 64); err == nil {
			out.Bool = (valInt == raModeDynamic || valInt == raModeDynamicSOS)
			out.Valid = true
		}
	}
	return
}

func interpretVectoringState(values map[string]string) (downstream, upstream models.VectoringValue) {
	vectorStatus := interpretStatusIntValue(values, "eVectorStatus", 1)
	vectorFriendlyStatus := interpretStatusIntValue(values, "eVectorFriendlyStatus", 1)

	if vectorStatus.Valid && vectorFriendlyStatus.Valid {
		downstream.Valid = true
		upstream.Valid = true

		switch {
		case vectorStatus.Int == 2:
			downstream.State = models.VectoringStateFull
			upstream.State = models.VectoringStateFull
		case vectorStatus.Int == 1:
			downstream.State = models.VectoringStateFull
		case vectorFriendlyStatus.Int == 2:
			downstream.State = models.VectoringStateFriendly
			upstream.State = models.VectoringStateFriendly
		case vectorFriendlyStatus.Int == 1:
			downstream.State = models.VectoringStateFriendly
		}
	}

	return
}

func parseStatusState(status *models.Status, lsg dataItem) {
	lsgValues := parseValues(lsg.Output)

	lineStateStr := lsgValues["nLineState"]
	if strings.HasPrefix(lineStateStr, "0x") {
		lineStateStr = lineStateStr[2:]
	}

	lineState, err := strconv.ParseUint(lineStateStr, 16, 64)
	if err != nil {
		return
	}

	status.State = parseLineState(lineState)
}

func parseStatusMode(status *models.Status, g997xtusesg, bpstg dataItem) {
	g977xtusesgValues := parseValues(g997xtusesg.Output)
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

		bpstgValues := parseValues(bpstg.Output)
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

func parseStatusInventory(status *models.Status, vig, g997ligFar dataItem) {
	vigValues := parseValues(vig.Output)
	status.NearEndInventory.Vendor = "Infineon"
	status.NearEndInventory.Version = vigValues["DSL_ChipSetFWVersion"]

	g997ligFarValues := parseValues(g997ligFar.Output)
	vendorID := interpretStatusBytes(g997ligFarValues, "G994VendorID")
	if len(vendorID) == 8 {
		status.FarEndInventory.Vendor = helpers.FormatVendor(string(vendorID[2:6]))
		status.FarEndInventory.Version = helpers.FormatVersion(status.FarEndInventory.Vendor, vendorID[6:8])
	}
}

func parseStatusChannelStatus(status *models.Status, g997csgUS, g997csgDS dataItem, apiVersion string) {
	g997csgUSValues := parseValues(g997csgUS.Output)
	g997csgDSValues := parseValues(g997csgDS.Output)

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

	// Per the standard, this value should be provided in 0.1 symbol granularity.
	// It seems that for Vinax (API version 2), the value is given with 0.5 symbol granularity instead.
	status.UpstreamImpulseNoiseProtection.FloatValue = interpretStatusFloatValue(g997csgUSValues, "ActualImpulseNoiseProtection", 10)
	status.DownstreamImpulseNoiseProtection.FloatValue = interpretStatusFloatValue(g997csgDSValues, "ActualImpulseNoiseProtection", 10)
	if strings.HasPrefix(apiVersion, "2") {
		status.UpstreamImpulseNoiseProtection.FloatValue.Float *= 5
		status.DownstreamImpulseNoiseProtection.FloatValue.Float *= 5
	}
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

func parseStatusLineStatus(status *models.Status, bins *models.Bins, g997lsgUS, g997lsgDS dataItem) {
	g997lsgUSValues := parseValues(g997lsgUS.Output)
	g997lsgDSValues := parseValues(g997lsgDS.Output)

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

func parseStatusLineFeatures(status *models.Status, lfsgUS, lfsgDS dataItem) {
	lfsgUSValues := parseValues(lfsgUS.Output)
	lfsgDSValues := parseValues(lfsgDS.Output)

	status.UpstreamRetransmissionEnabled = interpretStatusBoolValue(lfsgUSValues, "bReTxEnable")
	status.DownstreamRetransmissionEnabled = interpretStatusBoolValue(lfsgDSValues, "bReTxEnable")

	status.UpstreamBitswap.Enabled = interpretStatusBoolValue(lfsgUSValues, "bBitswapEnable")
	status.DownstreamBitswap.Enabled = interpretStatusBoolValue(lfsgDSValues, "bBitswapEnable")
}

func parseStatusRateAdaptationStatus(status *models.Status, g997rasgUS, g997rasgDS dataItem) {
	g997rasgUSValues := parseValues(g997rasgUS.Output)
	g997rasgDSValues := parseValues(g997rasgDS.Output)

	status.UpstreamSeamlessRateAdaptation.Enabled = interpretStatusRAMode(g997rasgUSValues, "RA_MODE")
	status.DownstreamSeamlessRateAdaptation.Enabled = interpretStatusRAMode(g997rasgDSValues, "RA_MODE")
}

func parseStatusOlrStatistics(status *models.Status, osgUS, osgDS dataItem) {
	osgUSValues := parseValues(osgUS.Output)
	osgDSValues := parseValues(osgDS.Output)

	status.UpstreamBitswap.Executed = interpretStatusIntValue(osgUSValues, "nBitswapExecuted", 1)
	status.DownstreamBitswap.Executed = interpretStatusIntValue(osgDSValues, "nBitswapExecuted", 1)

	status.UpstreamSeamlessRateAdaptation.Executed = interpretStatusIntValue(osgUSValues, "nSraExecuted", 1)
	status.DownstreamSeamlessRateAdaptation.Executed = interpretStatusIntValue(osgDSValues, "nSraExecuted", 1)
}

func parseStatusDSMStatus(status *models.Status, dsmsg dataItem) {
	dsmsgValues := parseValues(dsmsg.Output)

	status.DownstreamVectoringState, status.UpstreamVectoringState = interpretVectoringState(dsmsgValues)
}

func parseStatusChannelCounters(status *models.Status, pmccsgNear, pmccsgFar dataItem) {
	pmccsgNearValues := parseValues(pmccsgNear.Output)
	pmccsgFarValues := parseValues(pmccsgFar.Output)

	// elapsed time is only meaningful for showtime stats, not for total stats
	if strings.Contains(pmccsgNear.Command, "pmccsg") {
		status.Uptime = interpretStatusDuration(pmccsgNearValues, "nElapsedTime")
	}

	status.UpstreamFECCount = interpretStatusIntValue(pmccsgFarValues, "nFEC", 1)
	status.DownstreamFECCount = interpretStatusIntValue(pmccsgNearValues, "nFEC", 1)

	status.UpstreamCRCCount = interpretStatusIntValue(pmccsgFarValues, "nCodeViolations", 1)
	status.DownstreamCRCCount = interpretStatusIntValue(pmccsgNearValues, "nCodeViolations", 1)
}

func parseStatusLineSecCounters(status *models.Status, pmlscsgNear, pmlscsgFar dataItem) {
	pmlscsgNearValues := parseValues(pmlscsgNear.Output)
	pmlscsgFarValues := parseValues(pmlscsgFar.Output)

	status.UpstreamESCount = interpretStatusIntValue(pmlscsgFarValues, "nES", 1)
	status.DownstreamESCount = interpretStatusIntValue(pmlscsgNearValues, "nES", 1)

	status.UpstreamSESCount = interpretStatusIntValue(pmlscsgFarValues, "nSES", 1)
	status.DownstreamSESCount = interpretStatusIntValue(pmlscsgNearValues, "nSES", 1)
}

func parseStatusReTxCounters(status *models.Status, pmrtcsgNear, pmrtcsgFar dataItem) {
	pmrtcsgNearValues := parseValues(pmrtcsgNear.Output)
	pmrtcsgFarValues := parseValues(pmrtcsgFar.Output)

	status.UpstreamMinimumErrorFreeThroughput.IntValue = interpretStatusEFTRMin(pmrtcsgFarValues, "nEftrMin")
	status.DownstreamMinimumErrorFreeThroughput.IntValue = interpretStatusEFTRMin(pmrtcsgNearValues, "nEftrMin")
}

func parseStatusReTxStatistics(status *models.Status, rtsgNear, rtsgFar dataItem) {
	rtsgNearValues := parseValues(rtsgNear.Output)
	rtsgFarValues := parseValues(rtsgFar.Output)

	status.UpstreamRTXTXCount = interpretStatusIntValue(rtsgNearValues, "nTxRetransmitted", 1)
	status.DownstreamRTXTXCount = interpretStatusIntValue(rtsgFarValues, "nTxRetransmitted", 1)

	status.UpstreamRTXCCount = interpretStatusIntValue(rtsgFarValues, "nRxCorrected", 1)
	status.DownstreamRTXCCount = interpretStatusIntValue(rtsgNearValues, "nRxCorrected", 1)

	status.UpstreamRTXUCCount = interpretStatusIntValue(rtsgFarValues, "nRxUncorrectedProtected", 1)
	status.DownstreamRTXUCCount = interpretStatusIntValue(rtsgNearValues, "nRxUncorrectedProtected", 1)
}
