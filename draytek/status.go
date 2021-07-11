// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package draytek

import (
	"bufio"
	"fmt"
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

func parseStatus(statusStr, counts string) models.Status {
	var status models.Status

	values := readStatus(statusStr)
	interpretStatus(&status, values)

	valuesCounts := readCounts(counts)
	interpretCounts(&status, valuesCounts)

	return status
}

func readStatus(status string) map[string]string {
	values := make(map[string]string)

	scanner := bufio.NewScanner(strings.NewReader(status))

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, ":") && !strings.Contains(line, "---") {
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

func interpretStatus(status *models.Status, values map[string]string) {
	state := interpretStatusString(values, "State")
	status.State = models.ParseState(state)

	mode := interpretStatusString(values, "RunningMode")
	status.Mode = models.ParseMode(mode)

	status.ActualDownstreamRate = interpretStatusInt32Suffix(values, "DSActualRate", " bps") / 1000
	status.ActualUpstreamRate = interpretStatusInt32Suffix(values, "USActualRate", " bps") / 1000

	status.AttainableDownstreamRate = interpretStatusInt32Suffix(values, "DSAttainableRate", " bps") / 1000
	status.AttainableUpstreamRate = interpretStatusInt32Suffix(values, "USAttainableRate", " bps") / 1000

	status.DownstreamInterleavingDepth = interpretStatusInt16(values, "DSInterleaveDepth")
	status.UpstreamInterleavingDepth = interpretStatusInt16(values, "USInterleaveDepth")

	status.DownstreamAttenuation = interpretStatusFloat64Suffix(values, "NECurrentAttenuation", " dB")
	status.UpstreamAttenuation = interpretStatusFloat64Suffix(values, "FarCurrentAttenuation", " dB")

	status.DownstreamSNRMargin = interpretStatusFloat64Suffix(values, "CurSNRMargin", " dB")
	status.UpstreamSNRMargin = interpretStatusFloat64Suffix(values, "FarSNRMargin", " dB")

	// the "actual PSD" values actually seem to be the transmit power, although with wrong unit,
	// and at least for VDSL2 the upstream/downstream values are swapped
	powerUS := interpretStatusFloat64Suffix(values, "USactualPSD", " dB")
	powerDS := interpretStatusFloat64Suffix(values, "DSactualPSD", " dB")
	if status.Mode.Type == models.ModeTypeVDSL2 && powerUS > powerDS {
		status.DownstreamPower = powerUS
		status.UpstreamPower = powerDS
	} else {
		status.DownstreamPower = powerDS
		status.UpstreamPower = powerUS
	}

	status.DownstreamCRCCount = interpretStatusInt64(values, "NECRCCount")
	status.UpstreamCRCCount = interpretStatusInt64(values, "FECRCCount")

	status.DownstreamESCount = interpretStatusInt64(values, "NEESCount")
	status.UpstreamESCount = interpretStatusInt64(values, "FEESCount")

	status.LinecardVendor = interpretStatusVendor(values, "COITUVersion0", "COITUVersion1")
	status.LinecardVersion = interpretStatusLinecardVersion(values, "COITUVersion1")

	status.ModemVendor = interpretStatusVendor(values, "ITUVersion0", "ITUVersion1")
	status.ModemVersion = interpretStatusModemVersion(values, "ADSLFirmwareVersion", "VDSLFirmwareVersion")
}

func interpretStatusString(values map[string]string, key string) string {
	if val, ok := values[key]; ok {
		return val
	}
	return ""
}

func interpretStatusInt32Suffix(values map[string]string, key string, suffix string) int32 {
	if val, ok := values[key]; ok {
		if strings.HasSuffix(val, suffix) {
			val := val[:len(val)-len(suffix)]
			valInt, _ := strconv.ParseInt(val, 10, 32)
			return int32(valInt)
		}
	}
	return 0
}

func interpretStatusInt16(values map[string]string, key string) int16 {
	if val, ok := values[key]; ok {
		valInt, _ := strconv.ParseInt(val, 10, 16)
		return int16(valInt)
	}
	return 0
}

func interpretStatusFloat64Suffix(values map[string]string, key string, suffix string) float64 {
	if val, ok := values[key]; ok {
		if strings.HasSuffix(val, suffix) {
			val := val[:len(val)-len(suffix)]

			val = regexpBrokenFloat.ReplaceAllString(val, "$1$3$2.$4")
			if strings.HasPrefix(val, "--") {
				val = val[1:]
			}

			valFloat, _ := strconv.ParseFloat(val, 64)
			return valFloat
		}
	}
	return 0
}

func interpretStatusInt64(values map[string]string, key string) int64 {
	if val, ok := values[key]; ok {
		valInt, _ := strconv.ParseInt(val, 10, 64)
		return valInt
	}
	return 0
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

func interpretStatusLinecardVersion(values map[string]string, key string) string {
	v1 := helpers.ParseHexadecimal(interpretStatusString(values, key))
	if len(v1) == 4 {
		return fmt.Sprintf("%d.%d", v1[2], v1[3])
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

	version = regexpModemVersion.ReplaceAllString(version, "$1.$2.$3.$4.$5.$6")
	return version
}

func readCounts(counts string) map[string][2]string {
	values := make(map[string][2]string)

	scanner := bufio.NewScanner(strings.NewReader(counts))

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "[") && !strings.Contains(line, "Showtime") {
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

func interpretCounts(status *models.Status, values map[string][2]string) {
	status.DownstreamFECCount, status.UpstreamFECCount = interpretCountsInt64Ref(values, "FEC")
}

func interpretCountsInt64(values map[string][2]string, key string) (downstream, upstream int64) {
	if val, ok := values[key]; ok {
		downstream, _ = strconv.ParseInt(val[0], 10, 64)
		upstream, _ = strconv.ParseInt(val[1], 10, 64)
	}
	return
}

func interpretCountsInt64Ref(values map[string][2]string, key string) (downstream, upstream *int64) {
	d, u := interpretCountsInt64(values, key)
	downstream = &d
	upstream = &u
	return
}
