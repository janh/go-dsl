// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lantiq

import (
	"strconv"
	"strings"

	"3e8.eu/go/dsl/internal/helpers"
	"3e8.eu/go/dsl/models"
)

func parseBins(status *models.Status, data *data) models.Bins {
	var bins models.Bins

	bins.Mode = status.Mode

	parseBandBorders(&bins.Bands.Upstream, data.BandBorderStatus_US)
	parseBandBorders(&bins.Bands.Downstream, data.BandBorderStatus_DS)

	parseBitAllocation(&bins.Bits.Upstream, data.G997_BitAllocationNscShort_US)
	parseBitAllocation(&bins.Bits.Downstream, data.G997_BitAllocationNscShort_DS)

	parseSNRAllocation(&bins.SNR.Upstream, data.G997_SnrAllocationNscShort_US)
	parseSNRAllocation(&bins.SNR.Downstream, data.G997_SnrAllocationNscShort_DS)

	if len(bins.SNR.Upstream.Data) != len(bins.Bits.Upstream.Data) {
		parseDELTSNR(&bins.SNR.Upstream, bins.Mode.BinCount(), data.G997_DeltSNR_US)
	}
	if len(bins.SNR.Downstream.Data) != len(bins.Bits.Downstream.Data) {
		parseDELTSNR(&bins.SNR.Downstream, bins.Mode.BinCount(), data.G997_DeltSNR_DS)
	}

	parseDELTQLN(&bins.QLN.Upstream, bins.Mode.BinCount(), data.G997_DeltQLN_US)
	parseDELTQLN(&bins.QLN.Downstream, bins.Mode.BinCount(), data.G997_DeltQLN_DS)

	parseDELTHlog(&bins.Hlog.Upstream, bins.Mode.BinCount(), data.G997_DeltHLOG_US)
	parseDELTHlog(&bins.Hlog.Downstream, bins.Mode.BinCount(), data.G997_DeltHLOG_DS)

	helpers.GenerateBandsData(&bins)

	return bins
}

func parseBandBorders(bands *[]models.Band, data string) {
	v := parseValues(data)
	items := strings.Fields(v["nData"])

	for _, item := range items {
		itemSplit := strings.Split(item, ",")

		if len(itemSplit) != 5 {
			continue
		}

		limitStart := itemSplit[1]
		limitEnd := itemSplit[2]

		if len(limitStart) < 2 || limitStart[0] != '(' || len(limitEnd) < 2 || limitEnd[len(limitEnd)-1] != ')' {
			continue
		}

		limitStartInt, _ := strconv.Atoi(limitStart[1:])
		limitEndInt, _ := strconv.Atoi(limitEnd[:len(limitEnd)-1])

		band := models.Band{Start: limitStartInt, End: limitEndInt}
		*bands = append(*bands, band)
	}
}

func parseBinsShort(data string) (rawItems []string) {
	v := parseValues(data)

	numData, err := strconv.Atoi(v["nNumData"])
	if err != nil {
		return
	}

	rawItems = make([]string, numData)

	items := strings.Fields(v["nData"])

	for num, item := range items {
		rawItems[num] = item
	}

	return
}

func parseBitAllocation(out *models.BinsBits, data string) {
	rawValues := parseBinsShort(data)

	out.Data = make([]int8, len(rawValues))

	for num, val := range rawValues {
		valInt, err := strconv.ParseInt(val, 16, 8)
		if err == nil && valInt > 0 {
			out.Data[num] = int8(valInt)
		}
	}
}

func parseSNRAllocation(out *models.BinsFloat, data string) {
	rawValues := parseBinsShort(data)

	out.GroupSize = 1
	out.Data = make([]float64, len(rawValues))

	for num, val := range rawValues {
		valUint, err := strconv.ParseUint(val, 16, 8)
		if err == nil && valUint != 255 {
			out.Data[num] = -32 + float64(valUint)/2
		}
	}
}

func parseBinsDELT(data string, binCount int) (rawItems []string, groupSize int) {
	v := parseValues(data)

	numData, err := strconv.Atoi(v["nNumData"])
	if err != nil {
		return
	}

	groupSize, err = strconv.Atoi(v["nGroupSize"])
	if err != nil {
		return
	}
	// apparently the group size is sometimes not reported correctly
	if groupSize == 1 {
		groupSize = binCount / numData
	}

	rawItems = make([]string, numData)

	items := strings.Fields(v["nData"])

	for _, item := range items {
		if len(item) < 2 || item[0] != '(' || item[len(item)-1] != ')' {
			continue
		}

		itemSplit := strings.SplitN(item[1:len(item)-1], ",", 2)
		if len(itemSplit) != 2 {
			continue
		}

		num, err := strconv.Atoi(itemSplit[0])
		if err != nil {
			continue
		}

		rawItems[num] = itemSplit[1]
	}

	return
}

func parseDELTSNR(out *models.BinsFloat, binCount int, data string) {
	rawValues, groupSize := parseBinsDELT(data, binCount)

	out.GroupSize = groupSize
	out.Data = make([]float64, len(rawValues))

	for num, val := range rawValues {
		valUint, err := strconv.ParseUint(val, 10, 8)
		if err == nil && valUint != 255 {
			valFloat := -32 + float64(valUint)/2
			out.Data[num] = valFloat
		}
	}
}

func parseDELTQLN(out *models.BinsFloat, binCount int, data string) {
	rawValues, groupSize := parseBinsDELT(data, binCount)

	out.GroupSize = groupSize
	out.Data = make([]float64, len(rawValues))

	for num, val := range rawValues {
		valUint, err := strconv.ParseUint(val, 10, 8)
		if err == nil && valUint != 255 {
			valFloat := -23 - float64(valUint)/2
			out.Data[num] = valFloat
		}
	}
}

func parseDELTHlog(out *models.BinsFloat, binCount int, data string) {
	rawValues, groupSize := parseBinsDELT(data, binCount)

	out.GroupSize = groupSize
	out.Data = make([]float64, len(rawValues))

	for num, val := range rawValues {
		valUint, err := strconv.ParseUint(val, 10, 10)
		if err == nil && valUint != 1023 {
			valFloat := 6 - float64(valUint)/10
			out.Data[num] = valFloat
		} else {
			out.Data[num] = -96.3
		}
	}
}
