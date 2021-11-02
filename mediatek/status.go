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

func parseStatus(adslStats, vdslInterfaceConfig, adslFwVer string) models.Status {
	var status models.Status

	values := parseKeyValue(adslStats, vdslInterfaceConfig)

	parseStatusBasic(&status, values)

	parseStatusRates(&status, values)
	parseStatusLine(&status, values)

	parseStatusINP(&status, values)

	parseStatusVectoring(&status, values)

	parseStatusErrors(&status, values)

	parseFarVersion(&status, values)
	parseFirmwareVersion(&status, adslFwVer)

	return status
}

func parseKeyValue(data ...string) map[string]string {
	values := make(map[string]string)

	for _, item := range data {
		scanner := bufio.NewScanner(strings.NewReader(item))

		for scanner.Scan() {
			line := scanner.Text()
			split := strings.SplitN(line, ":", 2)

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

	status.DownstreamImpulseNoiseProtection.FloatValue = interpretFloatValueINP(values, "inpdsnormal")
	status.UpstreamImpulseNoiseProtection.FloatValue = interpretFloatValueINP(values, "inpusnormal")

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
