// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package mediatek

import (
	"bufio"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"3e8.eu/go/dsl/internal/helpers"
	"3e8.eu/go/dsl/models"
)

var regexpFilterCharacters = regexp.MustCompile(`[^a-zA-Z0-9]+`)
var regexpNumber = regexp.MustCompile("[0-9]+")

func parseStatus(adslStats, vdslInterfaceConfig, adslFwVer,
	wanVdsl2Mgcnt, wanVdsl2PmsPmdRx, wanVdsl2PmsPmdTx string) models.Status {

	var status models.Status

	values := parseKeyValue(":", adslStats, vdslInterfaceConfig)

	parseStatusBasic(&status, values)

	parseStatusRates(&status, values)
	parseStatusLine(&status, values)

	parseStatusINP(&status, values)

	parseStatusVectoring(&status, values)

	parseStatusErrors(&status, values)

	parseFarVersion(&status, values)
	parseFirmwareVersion(&status, adslFwVer)

	valuesMgcnt := parseKeyValue(":", wanVdsl2Mgcnt)
	valuesPmsPmdRx := parseKeyValue("=", wanVdsl2PmsPmdRx)
	valuesPmsPmdTx := parseKeyValue("=", wanVdsl2PmsPmdTx)

	parseExtendedStatusCounters(&status, valuesMgcnt)
	parseExtendedStatusRetransmission(&status, valuesMgcnt)
	parseExtendedStatusINPDelay(&status, valuesMgcnt, valuesPmsPmdRx, valuesPmsPmdTx)

	return status
}

func parseKeyValue(separator string, data ...string) map[string]string {
	values := make(map[string]string)

	for _, item := range data {
		scanner := bufio.NewScanner(strings.NewReader(item))

		for scanner.Scan() {
			line := scanner.Text()
			split := strings.SplitN(line, separator, 2)

			if len(split) == 2 {
				key := strings.ToLower(regexpFilterCharacters.ReplaceAllString(split[0], ""))
				val := strings.TrimSpace(split[1])

				values[key] = val
			}
		}
	}

	return values
}

func interpretString(values map[string]string, key string) string {
	if val, ok := values[key]; ok {
		return val
	}
	return ""
}

func interpretDuration(values map[string]string, key string) models.Duration {
	var duration models.Duration

	if val, ok := values[key]; ok {
		numbers := regexpNumber.FindAllString(val, -1)

		for i, num := range numbers {
			count := len(numbers) - i
			numInt, _ := strconv.ParseInt(num, 10, 64)

			var factor time.Duration
			if count == 1 {
				factor = time.Second
			} else if count == 2 {
				factor = time.Minute
			} else if count == 3 {
				factor = time.Hour
			} else if count == 4 {
				factor = 24 * time.Hour
			}

			duration.Duration += time.Duration(numInt) * factor
		}
	}

	return duration
}

func interpretIntValue(values map[string]string, suffix string, keys ...string) (out models.IntValue) {
	for _, key := range keys {
		if val, ok := values[key]; ok {
			if suffix != "" && strings.HasSuffix(val, suffix) {
				val = val[:len(val)-len(suffix)]
			}

			if valInt, err := strconv.ParseInt(val, 10, 64); err == nil {
				out.Int = valInt
				out.Valid = true

				if valInt != 0 {
					return
				}
			}
		}
	}

	return
}

func interpretIntValueError(values map[string]string, keys ...string) (out models.IntValue) {
	for _, key := range keys {
		if val, ok := values[key]; ok {
			if bracketIndex := strings.IndexByte(val, '('); bracketIndex != -1 {
				val = strings.TrimSpace(val[:bracketIndex])
			}

			if valInt, err := strconv.ParseInt(val, 10, 64); err == nil {
				out.Int = valInt
				out.Valid = true

				if valInt != 0 {
					return
				}
			}
		}
	}

	return
}

func interpretFloatValue(values map[string]string, suffix string, keys ...string) (out models.FloatValue) {
	for _, key := range keys {
		if val, ok := values[key]; ok {
			if suffix != "" && strings.HasSuffix(val, suffix) {
				val = val[:len(val)-len(suffix)]
			}

			if valFloat, err := strconv.ParseFloat(val, 64); err == nil {
				out.Float = valFloat
				out.Valid = true
			}

			if out.Valid && out.Float != 0 {
				break
			}
		}
	}

	return
}

func interpretFloatValueINP(values map[string]string, key string) (out models.FloatValue) {
	if val, ok := values[key]; ok {
		if strings.HasPrefix(val, "0x") {
			// the value seems to actually be a 64-bit floating point number truncated to 32 bits, encoded in hex
			// (this was verified using the output from "wan vdsl2 show pms_pmd" and the data from the other end)
			if valUint, err := strconv.ParseUint(val[2:], 16, 32); err == nil {
				out.Float = float64(math.Float64frombits(valUint << 32))
				out.Valid = true
			}
		} else {
			if valFloat, err := strconv.ParseFloat(val, 64); err == nil {
				out.Float = valFloat
				out.Valid = true
			}
		}
	}

	return
}

func interpretFloatValueINPSumProc(values map[string]string, keyNormal, keySHINE, keyREIN string) (out models.FloatValue) {
	inpNormal := interpretFloatValueINP(values, keyNormal)
	inpSHINE := interpretFloatValue(values, "", keySHINE)
	inpREIN := interpretFloatValue(values, "", keyREIN)

	out.Float = inpNormal.Float + inpSHINE.Float + inpREIN.Float
	out.Valid = inpNormal.Valid || inpSHINE.Valid || inpREIN.Valid

	return
}

func interpretFloatValueINPSumWAN(valuesPmsPmd, valuesMgcnt map[string]string,
	keyNormal, keySHINE, keyREIN string) (out models.FloatValue) {

	inpNormal := interpretFloatValue(valuesPmsPmd, " (symbols)", keyNormal)
	inpSHINE := interpretFloatValue(valuesMgcnt, "", keySHINE)
	inpREIN := interpretFloatValue(valuesMgcnt, "", keyREIN)

	out.Float = inpNormal.Float + 0.1*inpSHINE.Float + 0.1*inpREIN.Float
	out.Valid = inpNormal.Valid || inpSHINE.Valid || inpREIN.Valid

	return
}

func parseStatusBasic(status *models.Status, values map[string]string) {

	status.State = models.ParseState(interpretString(values, "adsllinkstatus"))

	opmode := interpretString(values, "opmode")
	annex := interpretString(values, "adsltype")
	profile := interpretString(values, "currentprofiles")
	status.Mode = models.ParseMode(opmode + " " + annex + " " + profile)

	status.Uptime = interpretDuration(values, "adsluptime")
}

func parseStatusRates(status *models.Status, values map[string]string) {

	status.DownstreamActualRate.IntValue = interpretIntValue(values, " kbps",
		"nearendinterleavedchannelbitrate", "nearendfastchannelbitrate")
	status.UpstreamActualRate.IntValue = interpretIntValue(values, " kbps",
		"farendinterleavedchannelbitrate", "farendfastchannelbitrate")

	status.DownstreamAttainableRate.IntValue = interpretIntValue(values, " kbps", "attaindownstream")
	status.UpstreamAttainableRate.IntValue = interpretIntValue(values, " kbps", "attainupstream")
}

func parseStatusLine(status *models.Status, values map[string]string) {

	status.DownstreamAttenuation.FloatValue = interpretFloatValue(values, " dB", "attenuationdownstream")
	status.UpstreamAttenuation.FloatValue = interpretFloatValue(values, " dB", "attenuationupstream")

	status.DownstreamSNRMargin.FloatValue = interpretFloatValue(values, " dB", "noisemargindownstream")
	status.UpstreamSNRMargin.FloatValue = interpretFloatValue(values, " dB", "noisemarginupstream")

	status.DownstreamPower.FloatValue = interpretFloatValue(values, " dbm", "outputpowerdownstream")
	status.UpstreamPower.FloatValue = interpretFloatValue(values, " dbm", "outputpowerupstream")
}

func parseStatusINP(status *models.Status, values map[string]string) {

	// Non-Asus devices seem to only report a single "Interleave Depth" value instead
	interleavingDepthDown := interpretIntValue(values, "", "interleavedepthdownstream")
	if interleavingDepthDown.Valid && interleavingDepthDown.Int == 1 {
		status.DownstreamInterleavingDelay.FloatValue.Float = 0
		status.DownstreamInterleavingDelay.FloatValue.Valid = true
	}
	interleavingDepthUp := interpretIntValue(values, "", "interleavedepthupstream")
	if interleavingDepthUp.Valid && interleavingDepthUp.Int == 1 {
		status.UpstreamInterleavingDelay.FloatValue.Float = 0
		status.UpstreamInterleavingDelay.FloatValue.Valid = true
	}

	// These seem to be Asus-specific values
	status.DownstreamImpulseNoiseProtection.FloatValue = interpretFloatValueINPSumProc(values,
		"inpdsnormal", "inpdsginpshine", "inpdsginprein")
	status.UpstreamImpulseNoiseProtection.FloatValue = interpretFloatValueINPSumProc(values,
		"inpusnormal", "inpusginpshine", "inpusginprein")

	// The "G.INP, Upstream only" and "G.INP, Downstream only" information seems to be Asus-specific
	opmode := interpretString(values, "opmode")
	opmode = strings.ToLower(regexpFilterCharacters.ReplaceAllString(opmode, ""))
	status.DownstreamRetransmissionEnabled.Bool = strings.Contains(opmode, "ginp") && !strings.Contains(opmode, "ginpu")
	status.DownstreamRetransmissionEnabled.Valid = true
	status.UpstreamRetransmissionEnabled.Bool = strings.Contains(opmode, "ginp") && !strings.Contains(opmode, "ginpd")
	status.UpstreamRetransmissionEnabled.Valid = true
}

func parseStatusVectoring(status *models.Status, values map[string]string) {

	opmode := interpretString(values, "opmode")
	opmode = strings.ToLower(regexpFilterCharacters.ReplaceAllString(opmode, ""))
	if strings.Contains(opmode, "vector") {
		status.DownstreamVectoringState.State = models.VectoringStateFull
		status.DownstreamVectoringState.Valid = true
	}
}

func parseStatusErrors(status *models.Status, values map[string]string) {

	status.DownstreamFECCount = interpretIntValue(values, "", "nearendfecerrorfast", "nearendfecerrorinterleaved")
	status.UpstreamFECCount = interpretIntValue(values, "", "farendfecerrorfast", "farendfecerrorinterleaved")

	status.DownstreamCRCCount = interpretIntValue(values, "", "nearendcrcerrorfast", "nearendcrcerrorinterleaved")
	status.UpstreamCRCCount = interpretIntValue(values, "", "farendcrcerrorfast", "farendcrcerrorinterleaved")
}

func parseFarVersion(status *models.Status, values map[string]string) {
	data := interpretString(values, "farendituidentification")

	if len(data) == 16 {
		vendorByte := helpers.ParseHexadecimal(data[4:12])
		status.FarEndInventory.Vendor = helpers.FormatVendor(string(vendorByte))

		versionByte := helpers.ParseHexadecimal(data[12:16])
		status.FarEndInventory.Version = helpers.FormatVersion(status.FarEndInventory.Vendor, versionByte)
	}
}

func parseFirmwareVersion(status *models.Status, fwVer string) {
	status.NearEndInventory.Vendor = "TrendChip"

	startIndex := strings.Index(fwVer, "FwVer:")
	if startIndex != -1 {
		fwVer = fwVer[startIndex+6:]

		endIndex := strings.IndexByte(fwVer, '_')
		if endIndex != -1 {
			fwVer = fwVer[:endIndex]
		}

		status.NearEndInventory.Version = fwVer
	}
}

func parseExtendedStatusCounters(status *models.Status, values map[string]string) {
	status.DownstreamFECCount = interpretIntValueError(values, "nearendpath1fec", "nearendpath0fec")
	status.UpstreamFECCount = interpretIntValueError(values, "farendpath1fec", "farendpath0fec")

	status.DownstreamCRCCount = interpretIntValueError(values, "nearendpath1crc", "nearendpath0crc")
	status.UpstreamCRCCount = interpretIntValueError(values, "farendpath1crc", "farendpath0crc")

	status.DownstreamESCount = interpretIntValueError(values, "nearenderrsec")
	status.UpstreamESCount = interpretIntValueError(values, "farenderrsec")

	status.DownstreamSESCount = interpretIntValueError(values, "nearendsessec")
	status.UpstreamSESCount = interpretIntValueError(values, "farendsessec")
}

func parseExtendedStatusRetransmission(status *models.Status, values map[string]string) {
	// The retransmission status is only set to false here, as some driver versions don't print these values, i.e.
	// absence does not necessarily mean that retransmission is enabled. Checking for the presence of counter values
	// also doesn't work, as those are always printed for both near and far end, or not at all. However, both status
	// values should already be set to true if the opmode contains "G.INP".
	if _, ok := values["dsretxdisable"]; ok {
		status.DownstreamRetransmissionEnabled.Bool = false
	}
	if _, ok := values["usretxdisable"]; ok {
		status.UpstreamRetransmissionEnabled.Bool = false
	}

	// This actually needs to be swapped to give correct results
	status.DownstreamRTXTXCount = interpretIntValue(values, "", "farendtxdtucounter")
	status.UpstreamRTXTXCount = interpretIntValue(values, "", "nearendtxdtucounter")

	status.DownstreamRTXCCount = interpretIntValue(values, "", "nearendcorrectdtucounter")
	status.UpstreamRTXCCount = interpretIntValue(values, "", "farendcorrectdtucounter")

	status.DownstreamRTXUCCount = interpretIntValue(values, "", "nearenderrdtucounter")
	status.UpstreamRTXUCCount = interpretIntValue(values, "", "farenderrdtucounter")

	status.DownstreamMinimumErrorFreeThroughput.IntValue = interpretIntValue(values, "", "nearendminerrorfreethroughputrate")
	status.UpstreamMinimumErrorFreeThroughput.IntValue = interpretIntValue(values, "", "farendminerrorfreethroughputrate")
}

func parseExtendedStatusINPDelay(status *models.Status, valuesMgcnt, valuesPmsPmdRx, valuesPmsPmdTx map[string]string) {
	status.DownstreamImpulseNoiseProtection.FloatValue = interpretFloatValueINPSumWAN(
		valuesPmsPmdRx, valuesMgcnt, "inp", "nearendactshinevalue", "nearendactreinvalue")
	status.UpstreamImpulseNoiseProtection.FloatValue = interpretFloatValueINPSumWAN(
		valuesPmsPmdTx, valuesMgcnt, "inp", "farendactshinevalue", "farendactreinvalue")

	status.DownstreamInterleavingDelay.FloatValue = interpretFloatValue(valuesPmsPmdRx, " (ms)", "delay")
	status.UpstreamInterleavingDelay.FloatValue = interpretFloatValue(valuesPmsPmdTx, " (ms)", "delay")
}
