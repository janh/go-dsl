// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package broadcom

import (
	"bufio"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"3e8.eu/go/dsl/helpers"
	"3e8.eu/go/dsl/models"
)

var regexpFilterCharacters = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func parseStatus(stats, vendor, version string) models.Status {
	var status models.Status

	parseStats(&status, stats)
	parseVendor(&status, vendor)
	parseVersion(&status, version)

	return status
}

func parseStats(status *models.Status, stats string) {
	basicStats := parseBasicStats(stats)
	interpretBasicStats(status, basicStats)

	extendedStats := parseExtendedStats(stats)
	interpretExtendedStats(status, extendedStats)
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

		if len(line) == 0 || line[0] == ' ' {
			break
		}
	}

	return values
}

func interpretBasicStats(status *models.Status, values map[string]string) {
	state := interpretBasicStatsString(values, "trainingstatus")
	status.State = models.ParseState(state)

	mode := interpretBasicStatsString(values, "mode")
	if strings.HasPrefix(strings.ToUpper(mode), "VDSL2") {
		profile := interpretBasicStatsString(values, "vdsl2profile")
		if strings.HasSuffix(strings.ToLower(profile), "brcmpriv1") {
			profile = "35b"
		}
		status.Mode = models.ParseMode("VDSL2 " + profile)
	} else {
		status.Mode = models.ParseMode(mode)
	}

	status.AttainableDownstreamRate, status.AttainableUpstreamRate = interpretBasicStatsRate(values, "max")
	status.ActualDownstreamRate, status.ActualUpstreamRate = interpretBasicStatsRate(values, "bearer")
}

func interpretBasicStatsString(values map[string]string, key string) string {
	if val, ok := values[key]; ok {
		return val
	}
	return ""
}

func interpretBasicStatsRate(values map[string]string, key string) (downstream, upstream int32) {
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

func interpretBasicStatsRateNumber(item string) int32 {
	separatorIndex := strings.LastIndexByte(item, '=')
	valueWithUnit := strings.TrimSpace(item[separatorIndex+1:])
	spaceIndex := strings.IndexRune(valueWithUnit, ' ')
	value := valueWithUnit[0:spaceIndex]
	valueInt, _ := strconv.ParseInt(value, 10, 32)
	return int32(valueInt)
}

func parseExtendedStats(stats string) map[string][2]string {
	values := make(map[string][2]string)

	scanner := bufio.NewScanner(strings.NewReader(stats))
	ignore := true

	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 || line[0] == ' ' || strings.Contains(line, " time") {
			lineLower := strings.ToLower(line)
			if strings.Contains(lineLower, "bearer") && !strings.Contains(lineLower, "bearer 0") {
				ignore = true
			} else if strings.Contains(lineLower, " time") && !strings.Contains(lineLower, "link time") {
				ignore = true
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

	return values
}

func interpretExtendedStats(status *models.Status, values map[string][2]string) {
	status.DownstreamInterleavingDepth, status.UpstreamInterleavingDepth = interpretExtendedStatsInt16(values, "d")

	status.DownstreamAttenuation, status.UpstreamAttenuation = interpretExtendedStatsFloat64(values, "attndb")
	status.DownstreamSNRMargin, status.UpstreamSNRMargin = interpretExtendedStatsFloat64(values, "snrdb")
	status.DownstreamPower, status.UpstreamPower = interpretExtendedStatsFloat64(values, "pwrdbm")

	status.DownstreamFECCount, status.UpstreamFECCount = interpretExtendedStatsInt64Ref(values, "fec")

	status.DownstreamRTXTXCount, status.UpstreamRTXTXCount = interpretExtendedStatsInt64Ref(values, "rtxtx")
	status.DownstreamRTXCCount, status.UpstreamRTXCCount = interpretExtendedStatsInt64Ref(values, "rtxc")
	status.DownstreamRTXUCCount, status.UpstreamRTXUCCount = interpretExtendedStatsInt64Ref(values, "rtxuc")

	status.DownstreamCRCCount, status.UpstreamCRCCount = interpretExtendedStatsInt64(values, "crc")
	status.DownstreamESCount, status.UpstreamESCount = interpretExtendedStatsInt64(values, "es")
}

func interpretExtendedStatsInt64(values map[string][2]string, key string) (downstream, upstream int64) {
	if val, ok := values[key]; ok {
		downstream, _ = strconv.ParseInt(val[0], 10, 64)
		upstream, _ = strconv.ParseInt(val[1], 10, 64)
	}
	return
}

func interpretExtendedStatsInt64Ref(values map[string][2]string, key string) (downstream, upstream *int64) {
	d, u := interpretExtendedStatsInt64(values, key)
	downstream = &d
	upstream = &u
	return
}

func interpretExtendedStatsInt16(values map[string][2]string, key string) (downstream, upstream int16) {
	if val, ok := values[key]; ok {
		d, _ := strconv.ParseInt(val[0], 10, 16)
		downstream = int16(d)
		u, _ := strconv.ParseInt(val[1], 10, 16)
		upstream = int16(u)
	}
	return
}

func interpretExtendedStatsFloat64(values map[string][2]string, key string) (downstream, upstream float64) {
	if val, ok := values[key]; ok {
		downstream, _ = strconv.ParseFloat(val[0], 64)
		upstream, _ = strconv.ParseFloat(val[1], 64)
	}
	return
}

func parseVendor(status *models.Status, vendor string) {
	scanner := bufio.NewScanner(strings.NewReader(vendor))

	for scanner.Scan() {
		line := scanner.Text()
		lineLower := strings.ToLower(line)

		if strings.HasPrefix(lineLower, "chipset vendor id:") {
			vendor := strings.TrimSpace(strings.Split(line, ":")[1])
			status.LinecardVendor = helpers.FormatVendor(vendor)
		} else if strings.HasPrefix(lineLower, "chipset versionnumber:") {
			version := strings.TrimSpace(strings.Split(line, ":")[1])
			versionByte := helpers.ParseHexadecimal(version)
			if len(versionByte) == 2 {
				status.LinecardVersion = fmt.Sprintf("%d.%d", versionByte[0], versionByte[1])
			}
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
			status.ModemVendor = "Broadcom"
			status.ModemVersion = strings.TrimSpace(line[index+1:])

			break
		}
	}
}
