// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package draytek

import (
	"bufio"
	"regexp"
	"strconv"
	"strings"

	"3e8.eu/go/dsl/internal/helpers"
	"3e8.eu/go/dsl/models"
)

var regexpColonWhitespace = regexp.MustCompile(`\s*:\s*`)
var regexpWhitespace = regexp.MustCompile(`\s+`)
var regexpFilterCharacters = regexp.MustCompile(`[^a-zA-Z0-9]+`)
var regexpBrokenFloat = regexp.MustCompile(`^(-?)(\d+)\.(-?)\s*(\d+)$`)
var regexpModemVersion = regexp.MustCompile(`^0([0-9A-F])-0([0-9A-F])-0([0-9A-F])-0([0-9A-F])-0([0-9A-F])-0([0-9A-F])$`)

func parseStatus(statusStr, counts, more, olr, basic string) models.Status {
	var status models.Status

	values := readStatus(statusStr)
	interpretStatus(&status, values)

	valuesCounts := readNearFar(counts)
	interpretCounts(&status, valuesCounts)

	valuesMore := readNearFar(more)
	interpretMore(&status, valuesMore)

	valuesOLR := readNearFar(olr)
	interpretOLR(&status, valuesOLR)

	valuesBasic := readBasic(basic)
	interpretBasic(&status, valuesBasic)

	normalizeOLRValues(&status)

	return status
}

func normalizeOLRValues(status *models.Status) {
	status.DownstreamBitswap.Normalize()
	status.UpstreamBitswap.Normalize()

	status.DownstreamSeamlessRateAdaptation.Normalize()
	status.UpstreamSeamlessRateAdaptation.Normalize()
}

func readStatus(status string) map[string]string {
	values := make(map[string]string)

	scanner := bufio.NewScanner(strings.NewReader(status))

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, ":") && !strings.Contains(line, "------") {
			readLine(values, line)
		}
	}

	return values
}

func readLine(values map[string]string, line string) {
	line = strings.TrimSpace(line)
	count := strings.Count(line, ":")

	if count == 2 {
		line = regexpColonWhitespace.ReplaceAllString(line, " : ")
		lineSplit := strings.SplitN(line, ":", 3)

		middle := lineSplit[1]
		if regexpWhitespace.MatchString(middle) {
			middleSplit := splitAtLongestWhitespace(middle)

			key1 := regexpFilterCharacters.ReplaceAllString(lineSplit[0], "")
			value1 := regexpWhitespace.ReplaceAllString(strings.TrimSpace(middleSplit[0]), " ")
			values[key1] = value1

			key2 := regexpFilterCharacters.ReplaceAllString(middleSplit[1], "")
			value2 := regexpWhitespace.ReplaceAllString(strings.TrimSpace(lineSplit[2]), " ")
			values[key2] = value2
		}

	} else if count == 1 {
		lineSplit := strings.SplitN(line, ":", 2)
		key := regexpFilterCharacters.ReplaceAllString(lineSplit[0], "")
		value := regexpWhitespace.ReplaceAllString(strings.TrimSpace(lineSplit[1]), " ")
		values[key] = value
	}
}

func splitAtLongestWhitespace(str string) [2]string {
	matches := regexpWhitespace.FindAllStringIndex(str, -1)

	var longest []int
	var longestCount int
	for _, m := range matches {
		count := m[1] - m[0]
		if count > longestCount {
			longest = m
			longestCount = count
		}
	}

	strA := str[:longest[0]]
	strB := str[longest[1]:]
	return [2]string{strA, strB}
}

func parseState(str string) models.State {
	switch strings.ToUpper(str) {
	case "READY":
		return models.StateDown
	case "TRAINING":
		return models.StateInit
	case "SHOWTIME":
		return models.StateShowtime
	}
	return models.StateUnknown
}

func interpretStatus(status *models.Status, values map[string]string) {
	state := interpretStatusString(values, "State")
	status.State = parseState(state)

	mode := interpretStatusString(values, "RunningMode")
	status.Mode = helpers.ParseMode(mode)

	status.DownstreamActualRate.IntValue = interpretStatusIntValueSuffixFactor(values, "DSActualRate", " bps", 1000)
	status.UpstreamActualRate.IntValue = interpretStatusIntValueSuffixFactor(values, "USActualRate", " bps", 1000)

	status.DownstreamAttainableRate.IntValue = interpretStatusIntValueSuffixFactor(values, "DSAttainableRate", " bps", 1000)
	status.UpstreamAttainableRate.IntValue = interpretStatusIntValueSuffixFactor(values, "USAttainableRate", " bps", 1000)

	status.DownstreamAttenuation.FloatValue = interpretStatusFloatValueSuffix(values, "NECurrentAttenuation", " dB")
	status.UpstreamAttenuation.FloatValue = interpretStatusFloatValueSuffix(values, "FarCurrentAttenuation", " dB")

	status.DownstreamSNRMargin.FloatValue = interpretStatusFloatValueSuffix(values, "CurSNRMargin", " dB")
	status.UpstreamSNRMargin.FloatValue = interpretStatusFloatValueSuffix(values, "FarSNRMargin", " dB")

	// the "actual PSD" values actually seem to be the transmit power, although with wrong unit,
	// and at least for VDSL2 the upstream/downstream values are swapped
	powerUS := interpretStatusFloatValueSuffix(values, "USactualPSD", " dB")
	powerDS := interpretStatusFloatValueSuffix(values, "DSactualPSD", " dB")
	if status.Mode.Type == models.ModeTypeVDSL2 && powerUS.Float > powerDS.Float {
		status.DownstreamPower.FloatValue = powerUS
		status.UpstreamPower.FloatValue = powerDS
	} else {
		status.DownstreamPower.FloatValue = powerDS
		status.UpstreamPower.FloatValue = powerUS
	}

	status.DownstreamCRCCount = interpretStatusIntValue(values, "NECRCCount")
	status.UpstreamCRCCount = interpretStatusIntValue(values, "FECRCCount")

	status.DownstreamESCount = interpretStatusIntValue(values, "NEESCount")
	status.UpstreamESCount = interpretStatusIntValue(values, "FEESCount")

	status.FarEndInventory.Vendor = interpretStatusVendor(values, "COITUVersion0", "COITUVersion1")
	status.FarEndInventory.Version = interpretStatusCOVersion(values, "COITUVersion1", status.FarEndInventory.Vendor)

	status.NearEndInventory.Vendor = interpretStatusVendor(values, "ITUVersion0", "ITUVersion1")
	status.NearEndInventory.Version = interpretStatusModemVersion(values, "ADSLFirmwareVersion", "VDSLFirmwareVersion")
}

func interpretStatusString(values map[string]string, key string) string {
	if val, ok := values[key]; ok {
		return val
	}
	return ""
}

func interpretStatusIntValueSuffixFactor(values map[string]string, key string, suffix string, factor int64) (out models.IntValue) {
	if val, ok := values[key]; ok {
		if strings.HasSuffix(val, suffix) {
			val := val[:len(val)-len(suffix)]
			if valInt, err := strconv.ParseInt(val, 10, 64); err == nil {
				out.Int = valInt / factor
				out.Valid = true
			}
		}
	}
	return
}

func interpretStatusIntValue(values map[string]string, key string) (out models.IntValue) {
	if val, ok := values[key]; ok {
		if valInt, err := strconv.ParseInt(val, 10, 64); err == nil {
			out.Int = valInt
			out.Valid = true
		}
	}
	return
}

func interpretStatusFloatValueSuffix(values map[string]string, key string, suffix string) (out models.FloatValue) {
	if val, ok := values[key]; ok {
		if strings.HasSuffix(val, suffix) {
			val := val[:len(val)-len(suffix)]

			val = regexpBrokenFloat.ReplaceAllString(val, "$1$3$2.$4")
			if strings.HasPrefix(val, "--") {
				val = val[1:]
			}

			if valFloat, err := strconv.ParseFloat(val, 64); err == nil {
				out.Float = valFloat
				out.Valid = true
			}
		}
	}
	return
}

func interpretStatusVendor(values map[string]string, key0, key1 string) string {
	v0 := helpers.ParseHexadecimal(interpretStatusString(values, key0))
	v1 := helpers.ParseHexadecimal(interpretStatusString(values, key1))
	if len(v0) == 4 && len(v1) == 4 {
		// vendor is encoded as ASCII in the last 2 bytes of COITUVersion0 and first 2 bytes of COITUVersion1
		vendor := []byte{v0[2], v0[3], v1[0], v1[1]}
		return helpers.FormatVendor(string(vendor))
	}
	return ""
}

func interpretStatusCOVersion(values map[string]string, key string, vendor string) string {
	v1 := helpers.ParseHexadecimal(interpretStatusString(values, key))
	if len(v1) == 4 {
		return helpers.FormatVersion(vendor, v1[2:4])
	}
	return ""
}

func interpretStatusModemVersion(values map[string]string, key, alternateKey string) string {
	version := interpretStatusString(values, key)
	if len(version) == 0 {
		version = interpretStatusString(values, alternateKey)
	}
	version = strings.ToUpper(version)

	posBracket := strings.IndexRune(version, '[')
	if posBracket != -1 {
		version = strings.TrimSpace(version[:posBracket])
	}

	versionSubmatch := regexpModemVersion.FindStringSubmatch(version)
	if versionSubmatch != nil {
		version = ""
		for i, item := range versionSubmatch[1:] {
			if i != 0 {
				version += "."
			}
			digit, _ := strconv.ParseUint(item, 16, 4)
			version += strconv.FormatUint(digit, 10)
		}
	}

	return version
}

func readNearFar(nearFar string) map[string][2]string {
	values := make(map[string][2]string)

	scanner := bufio.NewScanner(strings.NewReader(nearFar))

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "[") && !strings.Contains(line, "Showtime") && !strings.Contains(line, "OLR") {
			break
		}

		split := strings.SplitN(line, ":", 2)

		if len(split) == 2 {
			key := regexpFilterCharacters.ReplaceAllString(split[0], "")
			val := split[1]
			valSplit := strings.Fields(val)

			if len(valSplit) >= 2 {
				values[key] = [2]string{valSplit[0], valSplit[1]}
			}
		}
	}

	return values
}

func interpretNearFarIntValue(values map[string][2]string, key string) (near, far models.IntValue) {
	if val, ok := values[key]; ok {
		if n, err := strconv.ParseInt(val[0], 10, 64); err == nil {
			near.Int = n
			near.Valid = true
		}
		if f, err := strconv.ParseInt(val[1], 10, 64); err == nil {
			far.Int = f
			far.Valid = true
		}
	}
	return
}

func interpretNearFarBoolValue(values map[string][2]string, key string) (near, far models.BoolValue) {
	if val, ok := values[key]; ok {
		if n, err := strconv.ParseBool(val[0]); err == nil {
			near.Bool = n
			near.Valid = true
		}
		if f, err := strconv.ParseBool(val[1]); err == nil {
			far.Bool = f
			far.Valid = true
		}
	}
	return
}

func interpretNearFarFloatValueFactor(values map[string][2]string, key string, factor float64) (near, far models.FloatValue) {
	if val, ok := values[key]; ok {
		if n, err := strconv.ParseFloat(val[0], 64); err == nil {
			near.Float = n / factor
			near.Valid = true
		}
		if f, err := strconv.ParseFloat(val[1], 64); err == nil {
			far.Float = f / factor
			far.Valid = true
		}
	}
	return
}

func interpretNearFarInterleavingDelay(values map[string][2]string) (near, far models.FloatValue) {
	val, ok := values["InterleaveDelay"]
	if !ok {
		val, ok = values["INTLVDelay"]
		if !ok {
			return
		}
	}

	if n, err := strconv.ParseFloat(val[0], 64); err == nil {
		near.Float = n / 100
		near.Valid = true
	}
	if f, err := strconv.ParseFloat(val[1], 64); err == nil {
		far.Float = f / 100
		far.Valid = true
	}

	return
}

func interpretCounts(status *models.Status, values map[string][2]string) {
	status.DownstreamFECCount, status.UpstreamFECCount = interpretNearFarIntValue(values, "FEC")
	status.DownstreamSESCount, status.UpstreamSESCount = interpretNearFarIntValue(values, "SES")
}

func interpretMore(status *models.Status, values map[string][2]string) {
	status.UpstreamInterleavingDelay.FloatValue, status.DownstreamInterleavingDelay.FloatValue = interpretNearFarInterleavingDelay(values)

	status.UpstreamRetransmissionEnabled, status.DownstreamRetransmissionEnabled = interpretNearFarBoolValue(values, "ReTxEnable")

	status.UpstreamBitswap.Enabled, status.DownstreamBitswap.Enabled = interpretNearFarBoolValue(values, "Bitswap")

	// Dividing by 5 seems to give the correct value. According to the xDSL
	// specifications, this value is provided in 0.1 symbols granularity. However,
	// for old versions of the Lantiq API, it seems to have been provided with
	// 0.5 symbols granularity. Maybe Draytek then implemented a division by 2 and
	// never updated this.
	status.UpstreamImpulseNoiseProtection.FloatValue, status.DownstreamImpulseNoiseProtection.FloatValue =
		interpretNearFarFloatValueFactor(values, "INP", 5)
}

func interpretOLR(status *models.Status, values map[string][2]string) {
	status.UpstreamBitswap.Executed, status.DownstreamBitswap.Executed =
		interpretNearFarIntValue(values, "BitswapExecuted")

	status.UpstreamSeamlessRateAdaptation.Executed, status.DownstreamSeamlessRateAdaptation.Executed =
		interpretNearFarIntValue(values, "SraExecuted")
}

func readBasic(basic string) map[string]string {
	values := make(map[string]string)

	scanner := bufio.NewScanner(strings.NewReader(basic))

	for scanner.Scan() {
		line := scanner.Text()
		split := strings.SplitN(line, ":", 2)

		if len(split) == 2 {
			key := regexpFilterCharacters.ReplaceAllString(split[0], "")
			val := strings.TrimSpace(split[1])

			values[key] = val
		}
	}

	return values
}

func interpretBasicVectoringState(values map[string]string, key string) (downstream, upstream models.VectoringValue) {
	if val, ok := values[key]; ok {
		val = strings.ToLower(val)

		if strings.HasPrefix(val, "on") {
			downstream.Valid = true
			downstream.State = models.VectoringStateFull
			upstream.Valid = true
			if strings.Contains(val, "bidi") {
				upstream.State = models.VectoringStateFull
			}
		} else if strings.HasPrefix(val, "off") {
			downstream.Valid = true
			upstream.Valid = true
		}
	}
	return
}

func interpretBasic(status *models.Status, values map[string]string) {
	status.DownstreamVectoringState, status.UpstreamVectoringState = interpretBasicVectoringState(values, "GVectoringStatus")
}
