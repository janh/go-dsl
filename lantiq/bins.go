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

	parsePilotTones(&bins.PilotTones, data.PilotTonesStatus)

	parseBandBorders(&bins.Bands.Upstream, data.BandBorderStatus_US)
	parseBandBorders(&bins.Bands.Downstream, data.BandBorderStatus_DS)

	parseBitAllocation(&bins.Bits.Upstream, data.G997_BitAllocationNscShort_US)
	parseBitAllocation(&bins.Bits.Downstream, data.G997_BitAllocationNscShort_DS)

	helpers.GenerateBandsData(&bins)

	parseSNRAllocation(&bins.SNR.Upstream, data.G997_SnrAllocationNscShort_US)
	parseSNRAllocation(&bins.SNR.Downstream, data.G997_SnrAllocationNscShort_DS)

	if len(bins.SNR.Upstream.Data) != len(bins.Bits.Upstream.Data) {
		parseDELTSNR(&bins.SNR.Upstream, bins.Bands.Upstream, data.G997_DeltSNR_US)
	}
	if len(bins.SNR.Downstream.Data) != len(bins.Bits.Downstream.Data) {
		parseDELTSNR(&bins.SNR.Downstream, bins.Bands.Downstream, data.G997_DeltSNR_DS)
	}

	parseDELTQLN(&bins.QLN.Upstream, bins.Bands.Upstream, data.G997_DeltQLN_US)
	parseDELTQLN(&bins.QLN.Downstream, bins.Bands.Downstream, data.G997_DeltQLN_DS)

	parseDELTHlog(&bins.Hlog.Upstream, bins.Bands.Upstream, data.G997_DeltHLOG_US)
	parseDELTHlog(&bins.Hlog.Downstream, bins.Bands.Downstream, data.G997_DeltHLOG_DS)

	fixHlogScaling(status, &bins.Hlog.Downstream)

	return bins
}

func parsePilotTones(tones *[]int, data dataItem) {
	v := parseValues(data.Output)
	items := strings.Fields(v["nData"])

	for _, item := range items {
		itemSplit := strings.Split(item, ",")
		if len(itemSplit) != 2 {
			continue
		}

		index := itemSplit[1]
		if len(index) < 2 || index[len(index)-1] != ')' {
			continue
		}

		if valInt, err := strconv.Atoi(index[:len(index)-1]); err == nil {
			*tones = append(*tones, valInt)
		}
	}
}

func parseBandBorders(bands *[]models.Band, data dataItem) {
	v := parseValues(data.Output)
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

func parseBitAllocation(out *models.BinsBits, data dataItem) {
	rawValues := parseBinsShort(data.Output)

	out.Data = make([]int8, len(rawValues))

	for num, val := range rawValues {
		valInt, err := strconv.ParseInt(val, 16, 8)
		if err == nil && valInt > 0 {
			out.Data[num] = int8(valInt)
		}
	}
}

func parseSNRAllocation(out *models.BinsFloat, data dataItem) {
	rawValues := parseBinsShort(data.Output)

	out.GroupSize = 1
	out.Data = parseBinsHelper(rawValues, 16, 8, 255, -32, 2)
}

func parseBinsDELT(data string, bands []models.Band) (rawItems []string, groupSize int) {
	v := parseValues(data)

	numData, err := strconv.Atoi(v["nNumData"])
	if err != nil {
		return
	}

	groupSize, err = strconv.Atoi(v["nGroupSize"])
	if err != nil {
		return
	}
	// apparently the group size is sometimes not reported correctly, try to calculate it
	if groupSize == 1 && len(bands) != 0 && numData != 0 {
		lastBin := bands[len(bands)-1].End
		for numData*groupSize < lastBin+1 {
			groupSize *= 2
		}
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

func parseDELTSNR(out *models.BinsFloat, bands []models.Band, data dataItem) {
	rawValues, groupSize := parseBinsDELT(data.Output, bands)

	out.GroupSize = groupSize
	out.Data = parseBinsHelper(rawValues, 10, 8, 255, -32, 2)
}

func parseDELTQLN(out *models.BinsFloat, bands []models.Band, data dataItem) {
	rawValues, groupSize := parseBinsDELT(data.Output, bands)

	out.GroupSize = groupSize
	out.Data = parseBinsHelper(rawValues, 10, 8, 255, -23, -2)
}

func parseDELTHlog(out *models.BinsFloat, bands []models.Band, data dataItem) {
	rawValues, groupSize := parseBinsDELT(data.Output, bands)

	out.GroupSize = groupSize
	out.Data = parseBinsHelper(rawValues, 10, 10, 1023, 6, -10)
}

func parseBinsHelper(rawValues []string, base, bitSize int, invalid uint64, offset, divisor float64) (out []float64) {
	out = make([]float64, len(rawValues))

	allZeros := true

	for num, val := range rawValues {
		valUint, err := strconv.ParseUint(val, base, bitSize)
		if valUint != 0 {
			allZeros = false
		}
		var valFloat float64
		if err == nil && valUint != invalid {
			valFloat = offset + float64(valUint)/divisor
		} else {
			valFloat = offset + float64(invalid)/divisor
		}
		out[num] = valFloat
	}

	// Sometimes all zeros are reported (this has been seen on Vinax, as well as
	// VR9 with ADSL2+. While zero is a valid value, assume that the entire data is
	// invalid when all values are zero.
	if allZeros {
		out = nil
	}

	return
}

func fixHlogScaling(status *models.Status, data *models.BinsFloat) {
	// Issues with the Hlog data are known to exist for ADSL2+ on VR9, so for now
	// the correction is limited to that case.
	if !strings.HasPrefix(status.NearEndInventory.Version, "5.") || status.Mode.Type != models.ModeTypeADSL2Plus {
		return
	}

	// It looks like the first half contains the actual data for all carriers, the
	// third quarter is a copy of the second quarter, and the rest is garbage.

	if len(data.Data) == 512 {
		// Verify that second and third quarter are actually equal
		for i := 128; i < 256; i++ {
			if data.Data[i] != data.Data[i+128] {
				return
			}
		}

		data.GroupSize = 2
		data.Data = data.Data[:256]
	}
}
