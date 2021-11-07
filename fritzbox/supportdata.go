// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package fritzbox

import (
	"bufio"
	"strconv"
	"strings"

	"3e8.eu/go/dsl/internal/helpers"
	"3e8.eu/go/dsl/models"
)

func parseSupportData(status *models.Status, bins *models.Bins, supportData string) {
	if status.State != models.StateShowtime || supportData == "" {
		return
	}

	values := parseSupportDataValues(supportData)

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

	if val, ok := values["DS Bands"]; ok {
		parseSupportDataBands(&bins.Bands.Downstream, val, batGroupSize)
	}
	if val, ok := values["US Bands"]; ok {
		parseSupportDataBands(&bins.Bands.Upstream, val, batGroupSize)
	}

	if val, ok := values["HLOG Array"]; ok {
		bins.Hlog = parseSupportDataBinsDownUp(val, bins.Mode.BinCount(), bins.Bands, -96.3)
	} else {
		bins.Hlog.Downstream = parseSupportDataBins(values["HLOG DS Array"], bins.Mode.BinCount())
		bins.Hlog.Upstream = parseSupportDataBins(values["HLOG US Array"], bins.Mode.BinCount())
	}

	if val, ok := values["QLN Array"]; ok {
		bins.QLN = parseSupportDataBinsDownUp(val, bins.Mode.BinCount(), bins.Bands, 0)
	} else {
		bins.QLN.Downstream = parseSupportDataBins(values["QLN DS Array"], bins.Mode.BinCount())
		bins.QLN.Upstream = parseSupportDataBins(values["QLN US Array"], bins.Mode.BinCount())
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

func calculateSupportDataGroupSize(binCount, dataLength int) int {
	groupSize := 1
	for dataLength*groupSize < binCount {
		groupSize *= 2
	}
	return groupSize
}

func parseSupportDataBins(data string, binCount int) (out models.BinsFloat) {
	dataSplit := strings.Split(data, ",")

	if len(dataSplit) <= 1 {
		return
	}

	out.GroupSize = calculateSupportDataGroupSize(binCount, len(dataSplit))
	out.Data = make([]float64, len(dataSplit))

	for num, val := range dataSplit {
		valFloat, err := strconv.ParseFloat(val, 64)
		if err == nil {
			out.Data[num] = valFloat / 10
		}
	}

	return
}

func parseSupportDataBinsDownUp(data string, binCount int, bands models.BandsDownUp, defaultValue float64) (out models.BinsFloatDownUp) {
	dataSplit := strings.Split(data, ",")

	if len(dataSplit) <= 1 || len(bands.Downstream) == 0 || len(bands.Upstream) == 0 {
		return
	}

	groupSize := calculateSupportDataGroupSize(binCount, len(dataSplit))

	out.Downstream.GroupSize = groupSize
	out.Downstream.Data = make([]float64, len(dataSplit))

	out.Upstream.GroupSize = groupSize
	out.Upstream.Data = make([]float64, len(dataSplit))

	if defaultValue != 0 {
		for i := range out.Downstream.Data {
			out.Downstream.Data[i] = defaultValue
		}
		for i := range out.Upstream.Data {
			out.Upstream.Data[i] = defaultValue
		}
	}

	bandDecider, err := helpers.NewBandDecider(bands)
	if err != nil {
		return
	}

	for num, val := range dataSplit {
		valFloat, err := strconv.ParseFloat(val, 64)
		if err == nil {
			if bandDecider.IsDownstream(num * groupSize) {
				out.Downstream.Data[num] = valFloat / 10
			} else {
				out.Upstream.Data[num] = valFloat / 10
			}
		}
	}

	return
}
