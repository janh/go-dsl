// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package fritzbox

import (
	"bufio"
	"strconv"
	"strings"

	"3e8.eu/go/dsl/models"
)

func parseSupportData(status *models.Status, bins *models.Bins, d *rawDataSupport) {
	if status.State != models.StateShowtime || d.Data == "" {
		return
	}

	values := parseSupportDataValues(d.Data)

	if status.FarEndInventory.Vendor == "" && status.FarEndInventory.Version == "" {
		status.FarEndInventory.Vendor = values["ATUC Vendor ID"]
		status.FarEndInventory.Version = values["ATUC Vendor Info"]
	}

	if status.DownstreamRetransmissionEnabled.Bool {
		status.DownstreamRTXTXCount = interpretSupportDataIntValue(values, "US RTX retransmitted DTUs")
		status.DownstreamRTXCCount = interpretSupportDataIntValue(values, "DS RTX corrected DTUs")
		status.DownstreamRTXUCCount = interpretSupportDataIntValue(values, "DS RTX uncorrected DTUs")
	}

	if status.UpstreamRetransmissionEnabled.Bool {
		status.UpstreamRTXTXCount = interpretSupportDataIntValue(values, "DS RTX retransmitted DTUs")
		status.UpstreamRTXCCount = interpretSupportDataIntValue(values, "US RTX corrected DTUs")
		status.UpstreamRTXUCCount = interpretSupportDataIntValue(values, "US RTX uncorrected DTUs")
	}

	batGroupSize, _ := strconv.Atoi(values["BAT Bins per Group"])

	if val, ok := values["Pilot Array"]; ok {
		parseSupportDataPilotTones(&bins.PilotTones, val, batGroupSize)
	}

	if val, ok := values["DS Bands"]; ok {
		parseSupportDataBands(&bins.Bands.Downstream, val, batGroupSize)
	}
	if val, ok := values["US Bands"]; ok {
		parseSupportDataBands(&bins.Bands.Upstream, val, batGroupSize)
	}

	if val, ok := values["HLOG DS Array"]; ok {
		bins.Hlog.Downstream = parseSupportDataBins(val, bins.Bands.Downstream)
	} else if val, ok := values["HLOG Array"]; ok {
		bins.Hlog.Downstream = parseSupportDataBins(val, bins.Bands.Downstream)
	}
	if val, ok := values["HLOG US Array"]; ok {
		bins.Hlog.Upstream = parseSupportDataBins(val, bins.Bands.Upstream)
	}

	if val, ok := values["QLN DS Array"]; ok {
		bins.QLN.Downstream = parseSupportDataBins(val, bins.Bands.Downstream)
	} else if val, ok := values["QLN Array"]; ok {
		bins.QLN.Downstream = parseSupportDataBins(val, bins.Bands.Downstream)
	}
	if val, ok := values["QLN US Array"]; ok {
		bins.QLN.Upstream = parseSupportDataBins(val, bins.Bands.Upstream)
	}
}

func parseSupportDataValues(supportData string) map[string]string {
	values := make(map[string]string)

	scanner := bufio.NewScanner(strings.NewReader(supportData))
	for scanner.Scan() {
		line := scanner.Text()
		lineSplit := strings.SplitN(line, ":", 2)

		if len(lineSplit) != 2 {
			continue
		}

		key := lineSplit[0]
		val := strings.TrimSpace(lineSplit[1])

		values[key] = val
	}

	return values
}

func interpretSupportDataIntValue(values map[string]string, key string) (out models.IntValue) {
	if val, ok := values[key]; ok {
		if valInt, err := strconv.ParseInt(val, 10, 64); err == nil {
			out.Int = valInt
			out.Valid = true
		}
	}
	return
}

func parseSupportDataPilotTones(pilotTones *[]int, val string, groupSize int) {
	data := strings.Split(val, ",")
	if len(data) <= len(*pilotTones) || groupSize == 0 {
		return
	}

	*pilotTones = make([]int, 0)

	for _, item := range data {
		itemInt, _ := strconv.Atoi(strings.TrimSpace(item))
		*pilotTones = append(*pilotTones, itemInt*groupSize)
	}
}

func parseSupportDataBands(bands *[]models.Band, val string, groupSize int) {
	data := strings.Split(val, ",")
	if len(data)%2 != 0 || groupSize == 0 {
		return
	}

	*bands = make([]models.Band, 0)

	for i := 0; i < len(data); i += 2 {
		start, _ := strconv.Atoi(strings.TrimSpace(data[i]))
		end, _ := strconv.Atoi(strings.TrimSpace(data[i+1]))

		band := models.Band{Start: start * groupSize, End: end*groupSize + groupSize - 1}
		*bands = append(*bands, band)
	}
}

func calculateSupportDataGroupSize(lastBandsIndex int) int {
	groupSize := 1
	for 512*groupSize < lastBandsIndex+1 {
		groupSize *= 2
	}
	return groupSize
}

func parseSupportDataBins(data string, bands []models.Band) (out models.BinsFloat) {
	dataSplit := strings.Split(data, ",")

	if len(dataSplit) <= 1 || len(bands) == 0 {
		return
	}

	lastBandsIndex := bands[len(bands)-1].End
	out.GroupSize = calculateSupportDataGroupSize(lastBandsIndex)

	out.Data = make([]float64, len(dataSplit))

	for num, val := range dataSplit {
		valFloat, err := strconv.ParseFloat(val, 64)
		if err == nil {
			out.Data[num] = valFloat / 10
		}
	}

	return
}
