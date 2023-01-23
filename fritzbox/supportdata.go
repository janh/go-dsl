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

	parseSupportDataOLRValue(&status.DownstreamBitswap, values, "DS Bitswap")
	parseSupportDataOLRValue(&status.UpstreamBitswap, values, "US Bitswap")

	parseSupportDataOLRValue(&status.DownstreamSeamlessRateAdaptation, values, "DS SRA")
	parseSupportDataOLRValue(&status.UpstreamSeamlessRateAdaptation, values, "US SRA")

	if status.DownstreamRetransmissionEnabled.Bool {
		parseSupportDataIntValue(&status.DownstreamRTXTXCount, values, "US RTX retransmitted DTUs")
		parseSupportDataIntValue(&status.DownstreamRTXCCount, values, "DS RTX corrected DTUs")
		parseSupportDataIntValue(&status.DownstreamRTXUCCount, values, "DS RTX uncorrected DTUs")
	}

	if status.UpstreamRetransmissionEnabled.Bool {
		parseSupportDataIntValue(&status.UpstreamRTXTXCount, values, "DS RTX retransmitted DTUs")
		parseSupportDataIntValue(&status.UpstreamRTXCCount, values, "US RTX corrected DTUs")
		parseSupportDataIntValue(&status.UpstreamRTXUCCount, values, "US RTX uncorrected DTUs")
	}

	batGroupSize, _ := strconv.Atoi(values["BAT Bins per Group"])

	parseSupportDataPilotTones(&bins.PilotTones, batGroupSize, values, "Pilot Array")

	parseSupportDataBands(&bins.Bands.Downstream, batGroupSize, values, "DS Bands")
	parseSupportDataBands(&bins.Bands.Upstream, batGroupSize, values, "US Bands")

	bins.Hlog.Downstream = parseSupportDataBins(bins.Bands.Downstream, values, "HLOG DS Array", "HLOG Array")
	bins.Hlog.Upstream = parseSupportDataBins(bins.Bands.Upstream, values, "HLOG US Array")

	bins.QLN.Downstream = parseSupportDataBins(bins.Bands.Downstream, values, "QLN DS Array", "QLN Array")
	bins.QLN.Upstream = parseSupportDataBins(bins.Bands.Upstream, values, "QLN US Array")
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

func parseSupportDataIntValue(out *models.IntValue, values map[string]string, key string) {
	if val, ok := values[key]; ok {
		if valInt, err := strconv.ParseInt(val, 10, 64); err == nil {
			out.Int = valInt
			out.Valid = true
		}
	}
	return
}

func parseSupportDataBoolValue(out *models.BoolValue, values map[string]string, key string) {
	if val, ok := values[key]; ok {
		if valInt, err := strconv.Atoi(val); err == nil && (valInt == 0 || valInt == 1) {
			out.Bool = valInt == 1
			out.Valid = true
		}
	}
	return
}

func parseSupportDataOLRValue(out *models.OLRValue, values map[string]string, key string) {
	parseSupportDataBoolValue(&out.Enabled, values, key)
	parseSupportDataIntValue(&out.Executed, values, key+" Cnt")
}

func parseSupportDataPilotTones(pilotTones *[]int, groupSize int, values map[string]string, key string) {
	val, ok := values[key]
	if !ok {
		return
	}

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

func parseSupportDataBands(bands *[]models.Band, groupSize int, values map[string]string, key string) {
	val, ok := values[key]
	if !ok {
		return
	}

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

func parseSupportDataBins(bands []models.Band, values map[string]string, keys ...string) (out models.BinsFloat) {
	var data string
	var ok bool
	for _, key := range keys {
		if data, ok = values[key]; ok {
			break
		}
	}
	if !ok {
		return
	}

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
