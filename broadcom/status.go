// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package broadcom

import (
	"bufio"
	"regexp"
	"strconv"
	"strings"
	"time"

	"3e8.eu/go/dsl/internal/helpers"
	"3e8.eu/go/dsl/models"
)

var regexpFilterCharacters = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func parseStatus(stats, vectoring, vendor, version string) models.Status {
	var status models.Status

	parseStats(&status, stats)
	parseVectoring(&status, vectoring)
	parseVendor(&status, vendor)
	parseVersion(&status, version)

	return status
}

func parseStats(status *models.Status, stats string) {
	basicStats := parseBasicStats(stats)
	interpretBasicStats(status, basicStats)

	extendedStats, linkTime := parseExtendedStats(stats)
	interpretExtendedStats(status, extendedStats)
	interpretLinkTime(status, linkTime)
}

func parseBasicStats(stats string) map[string]string {
	values := make(map[string]string)

	scanner := bufio.NewScanner(strings.NewReader(stats))

	for scanner.Scan() {
		line := scanner.Text()
		split := strings.SplitN(line, ":", 2)

		if len(split) == 2 {
			key := strings.ToLower(regexpFilterCharacters.ReplaceAllString(split[0], ""))
			val := strings.TrimSpace(split[1])

			if key == "bearer" && (len(val) < 1 || val[0] != '0') {
				continue
			}

			values[key] = val
		}

		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			break
		}
	}

	return values
}

func parseState(str string) models.State {
	str = strings.ToLower(str)

	switch {

	case str == "idle":
		return models.StateDown

	case strings.HasPrefix(str, "g.994"):
		return models.StateInitHandshake

	case strings.Contains(str, "channel analysis"), strings.Contains(str, "message exchange"):
		return models.StateInitChannelAnalysisExchange

	case strings.HasPrefix(str, "g.992"), strings.HasPrefix(str, "g.993"):
		return models.StateInit

	case str == "showtime":
		return models.StateShowtime

	}

	return models.StateUnknown
}

func interpretBasicStats(status *models.Status, values map[string]string) {
	state := interpretBasicStatsString(values, "status")
	status.State = parseState(state)

	mode := interpretBasicStatsString(values, "mode")
	if strings.HasPrefix(strings.ToUpper(mode), "VDSL2") {
		profile := interpretBasicStatsString(values, "vdsl2profile")
		if strings.HasSuffix(strings.ToLower(profile), "brcmpriv1") {
			profile = "35b"
		}
		status.Mode = helpers.ParseMode("VDSL2 " + profile)
	} else {
		status.Mode = helpers.ParseMode(mode)
	}

	if status.Mode.Type == models.ModeTypeUnknown {
		if strings.Contains(state, "G.993") {
			status.Mode.Type = models.ModeTypeVDSL2
		} else if strings.Contains(state, "G.992") {
			status.Mode.Type = models.ModeTypeADSL
		}
	}

	status.DownstreamAttainableRate.IntValue, status.UpstreamAttainableRate.IntValue = interpretBasicStatsRate(values, "max")
	status.DownstreamActualRate.IntValue, status.UpstreamActualRate.IntValue = interpretBasicStatsRate(values, "bearer")
}

func interpretBasicStatsString(values map[string]string, key string) string {
	if val, ok := values[key]; ok {
		return val
	}
	return ""
}

func interpretBasicStatsRate(values map[string]string, key string) (downstream, upstream models.IntValue) {
	if val, ok := values[key]; ok {
		items := strings.Split(val, ",")

		for _, item := range items {
			item = strings.ToLower(strings.TrimSpace(item))

			if strings.HasPrefix(item, "upstream") {
				upstream = interpretBasicStatsRateNumber(item)
			} else if strings.HasPrefix(item, "downstream") {
				downstream = interpretBasicStatsRateNumber(item)
			}
		}
	}

	return
}

func interpretBasicStatsRateNumber(item string) (out models.IntValue) {
	separatorIndex := strings.LastIndexByte(item, '=')
	valueWithUnit := strings.TrimSpace(item[separatorIndex+1:])
	spaceIndex := strings.IndexRune(valueWithUnit, ' ')

	value := valueWithUnit[0:spaceIndex]
	if valueInt, err := strconv.ParseInt(value, 10, 64); err == nil {
		out.Int = valueInt
		out.Valid = true
	}

	return
}

func parseExtendedStats(stats string) (values map[string][2]string, linkTime string) {
	values = make(map[string][2]string)

	scanner := bufio.NewScanner(strings.NewReader(stats))
	ignore := true

	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 || line[0] == ' ' || line[0] == '\t' || strings.Contains(line, " time") {
			lineLower := strings.ToLower(line)
			if strings.Contains(lineLower, "bearer") && !strings.Contains(lineLower, "bearer 0") {
				ignore = true
			} else if strings.Contains(lineLower, " time") {
				if strings.Contains(lineLower, "link time") {
					ignore = false

					split := strings.SplitN(line, "=", 2)
					if len(split) == 2 {
						linkTime = strings.TrimSpace(split[1])
					}
				} else {
					ignore = true
				}
			} else {
				ignore = false
			}
			continue
		}

		if !ignore {
			split := strings.SplitN(line, ":", 2)

			if len(split) == 2 {
				key := strings.ToLower(regexpFilterCharacters.ReplaceAllString(split[0], ""))
				val := strings.TrimSpace(split[1])
				valSplit := strings.Fields(val)

				if len(valSplit) == 2 {
					values[key] = [2]string{valSplit[0], valSplit[1]}
				}
			}
		}
	}

	return
}

func interpretExtendedStats(status *models.Status, values map[string][2]string) {
	status.DownstreamBitswap, status.UpstreamBitswap = interpretExtendedStatsBitswap(values, "bitswap")

	status.DownstreamInterleavingDelay.FloatValue, status.UpstreamInterleavingDelay.FloatValue = interpretExtendedStatsFloatValue(values, "delay")
	status.DownstreamImpulseNoiseProtection.FloatValue, status.UpstreamImpulseNoiseProtection.FloatValue =
		interpretExtendedStatsFloatValue(values, "inp")
	status.DownstreamRetransmissionEnabled, status.UpstreamRetransmissionEnabled = interpretExtendedStatsBoolValueNonZero(values, "q")

	status.DownstreamAttenuation.FloatValue, status.UpstreamAttenuation.FloatValue = interpretExtendedStatsFloatValue(values, "attndb")
	status.DownstreamSNRMargin.FloatValue, status.UpstreamSNRMargin.FloatValue = interpretExtendedStatsFloatValue(values, "snrdb")
	status.DownstreamPower.FloatValue, status.UpstreamPower.FloatValue = interpretExtendedStatsFloatValue(values, "pwrdbm")

	status.DownstreamFECCount, status.UpstreamFECCount = interpretExtendedStatsIntValue(values, "fec")

	status.DownstreamRTXTXCount, status.UpstreamRTXTXCount = interpretExtendedStatsIntValue(values, "rtxtx")
	status.DownstreamRTXCCount, status.UpstreamRTXCCount = interpretExtendedStatsIntValue(values, "rtxc")
	status.DownstreamRTXUCCount, status.UpstreamRTXUCCount = interpretExtendedStatsIntValue(values, "rtxuc")

	status.DownstreamMinimumErrorFreeThroughput.IntValue, status.UpstreamMinimumErrorFreeThroughput.IntValue =
		interpretExtendedStatsIntValue(values, "mineftr")

	status.DownstreamCRCCount, status.UpstreamCRCCount = interpretExtendedStatsIntValue(values, "crc")
	status.DownstreamESCount, status.UpstreamESCount = interpretExtendedStatsIntValue(values, "es")
	status.DownstreamSESCount, status.UpstreamSESCount = interpretExtendedStatsIntValue(values, "ses")
}

func interpretExtendedStatsIntValue(values map[string][2]string, key string) (downstream, upstream models.IntValue) {
	if val, ok := values[key]; ok {
		if ds, err := strconv.ParseInt(val[0], 10, 64); err == nil {
			downstream.Int = ds
			downstream.Valid = true
		}
		if us, err := strconv.ParseInt(val[1], 10, 64); err == nil {
			upstream.Int = us
			upstream.Valid = true
		}
	}
	return
}

func interpretExtendedStatsFloatValue(values map[string][2]string, key string) (downstream, upstream models.FloatValue) {
	if val, ok := values[key]; ok {
		if ds, err := strconv.ParseFloat(val[0], 64); err == nil {
			downstream.Float = ds
			downstream.Valid = true
		}
		if us, err := strconv.ParseFloat(val[1], 64); err == nil {
			upstream.Float = us
			upstream.Valid = true
		}
	}
	return
}

func interpretExtendedStatsBoolValueNonZero(values map[string][2]string, key string) (downstream, upstream models.BoolValue) {
	if val, ok := values[key]; ok {
		if ds, err := strconv.ParseInt(val[0], 10, 64); err == nil {
			downstream.Bool = ds != 0
			downstream.Valid = true
		}
		if us, err := strconv.ParseInt(val[1], 10, 64); err == nil {
			upstream.Bool = us != 0
			upstream.Valid = true
		}
	}
	return
}

func interpretExtendedStatsBitswapString(val string) (out models.OLRValue) {
	split := strings.Split(val, "/")
	if len(split) != 2 {
		return
	}
	if valInt, err := strconv.ParseInt(split[0], 10, 64); err == nil {
		out.Executed.Int = valInt
		out.Executed.Valid = true
	}
	out.Normalize()
	return
}

func interpretExtendedStatsBitswap(values map[string][2]string, key string) (downstream, upstream models.OLRValue) {
	if val, ok := values[key]; ok {
		downstream = interpretExtendedStatsBitswapString(val[0])
		upstream = interpretExtendedStatsBitswapString(val[1])
	}
	return
}

func interpretLinkTime(status *models.Status, linkTime string) {
	if status.State != models.StateShowtime {
		return
	}

	split := strings.Fields(linkTime)
	if len(split)%2 != 0 || len(split) > 8 {
		return
	}

	var duration models.Duration

	for i := 0; i < len(split); i += 2 {
		valStr := split[i]
		unit := strings.ToLower(split[i+1])

		var factor time.Duration
		switch {
		case strings.HasPrefix(unit, "sec"):
			factor = time.Second
		case strings.HasPrefix(unit, "min"):
			factor = time.Minute
		case strings.HasPrefix(unit, "hour"):
			factor = time.Hour
		case strings.HasPrefix(unit, "day"):
			factor = 24 * time.Hour
		default:
			return
		}

		val, err := strconv.ParseInt(valStr, 10, 64)
		if err != nil {
			return
		}

		duration.Valid = true
		duration.Duration += time.Duration(val) * factor
	}

	status.Uptime = duration
}

func parseVectoring(status *models.Status, vectoring string) {
	scanner := bufio.NewScanner(strings.NewReader(vectoring))

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(strings.ToLower(line), "vectoring state:") {
			state := strings.TrimSpace(strings.Split(line, ":")[1])

			if state == "1" || state == "3" {
				status.DownstreamVectoringState.State = models.VectoringStateFull
				status.DownstreamVectoringState.Valid = true
			}

			break
		}
	}
}

func parseVendor(status *models.Status, vendor string) {
	scanner := bufio.NewScanner(strings.NewReader(vendor))

	for scanner.Scan() {
		line := scanner.Text()
		lineLower := strings.ToLower(line)

		if strings.HasPrefix(lineLower, "chipset vendor id:") {
			vendor := strings.TrimSpace(strings.Split(line, ":")[1])
			status.FarEndInventory.Vendor = helpers.FormatVendor(vendor)
		} else if strings.HasPrefix(lineLower, "chipset versionnumber:") {
			version := strings.TrimSpace(strings.Split(line, ":")[1])
			versionByte := helpers.ParseHexadecimal(version)
			status.FarEndInventory.Version = helpers.FormatVersion(status.FarEndInventory.Vendor, versionByte)
		}
	}
}

func parseVersion(status *models.Status, version string) {
	scanner := bufio.NewScanner(strings.NewReader(version))

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(strings.ToLower(line), "adsl phy:") {
			separator := byte(':')
			if strings.ContainsRune(line, '-') {
				separator = '-'
			}

			index := strings.LastIndexByte(line, separator)
			status.NearEndInventory.Vendor = "Broadcom"
			status.NearEndInventory.Version = strings.TrimSpace(line[index+1:])

			break
		}
	}
}
