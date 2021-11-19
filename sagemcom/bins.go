// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package sagemcom

import (
	"strconv"
	"strings"

	"3e8.eu/go/dsl/internal/helpers"
	"3e8.eu/go/dsl/models"
)

func interpretBins(status *models.Status, data *dslObj) models.Bins {
	var bins models.Bins

	bins.Mode = status.Mode

	testParams := &data.Lines[0].TestParams

	fixBinsData(testParams)

	interpretBinsDataSNR(&bins.SNR.Downstream, testParams.SNRpsds, testParams.SNRGds)
	interpretBinsDataSNR(&bins.SNR.Upstream, testParams.SNRpsus, testParams.SNRGus)

	interpretBinsDataQLN(&bins.QLN.Downstream, testParams.QLNpsds, testParams.QLNGds)
	interpretBinsDataQLN(&bins.QLN.Upstream, testParams.QLNpsus, testParams.QLNGus)

	interpretBinsDataHlog(&bins.Hlog.Downstream, testParams.HLOGpsds, testParams.HLOGGds)
	interpretBinsDataHlog(&bins.Hlog.Upstream, testParams.HLOGpsus, testParams.HLOGGus)

	helpers.GenerateBandsData(&bins)

	guessBinsProfile(&bins)

	return bins
}

func fixBinsData(testParams *lineTestParams) {
	// at least on Speedport Pro the Hlog upstream data is a copy of the downstream data

	fixBinsDataItem(&testParams.SNRpsds)
	fixBinsDataItem(&testParams.SNRpsus)
	if testParams.SNRpsus == testParams.SNRpsds {
		testParams.SNRpsus = ""
	}

	fixBinsDataItem(&testParams.QLNpsds)
	fixBinsDataItem(&testParams.QLNpsus)
	if testParams.QLNpsus == testParams.QLNpsds {
		testParams.QLNpsus = ""
	}

	fixBinsDataItem(&testParams.HLOGpsds)
	fixBinsDataItem(&testParams.HLOGpsus)
	if testParams.HLOGpsus == testParams.HLOGpsds {
		testParams.HLOGpsus = ""
	}
}

func fixBinsDataItem(str *string) {
	val := *str

	// on Speedport Pro, HLOGpsds and HLOGpsus seem to contain a spurious "139" at the end
	if len(val) >= 3 && val[len(val)-3:] == "139" {
		val = val[:len(val)-3]
	}

	// also on Speedport Pro, HLOGpsds seems to be contain the same data twice
	if len(val)%2 == 0 {
		firstHalf := val[:len(val)/2]
		secondHalf := val[len(val)/2:]
		if firstHalf == secondHalf {
			val = firstHalf
		}
	}

	*str = val
}

func interpretBinsData(out *models.BinsFloat, data string, groupSize int,
	defaultValue float64, processValueFunc func(val float64, scale bool) float64) {

	needsScaling := !strings.ContainsRune(data, '.')
	items := strings.Split(data, ",")

	out.GroupSize = groupSize
	out.Data = make([]float64, len(items))

	for i, item := range items {
		val, err := strconv.ParseFloat(item, 64)
		if err == nil {
			out.Data[i] = processValueFunc(val, needsScaling)
		} else {
			out.Data[i] = defaultValue
		}
	}
}

func interpretBinsDataSNR(out *models.BinsFloat, data string, groupSize int) {
	processValueFunc := func(val float64, scale bool) float64 {
		if scale {
			return -32 + val/2
		}
		return val
	}

	interpretBinsData(out, data, groupSize, 0, processValueFunc)
}

func interpretBinsDataQLN(out *models.BinsFloat, data string, groupSize int) {
	processValueFunc := func(val float64, scale bool) float64 {
		if scale {
			return -23 - val/2
		}
		return val
	}

	interpretBinsData(out, data, groupSize, 0, processValueFunc)
}

func interpretBinsDataHlog(out *models.BinsFloat, data string, groupSize int) {
	processValueFunc := func(val float64, scale bool) float64 {
		if scale {
			return 6 - val/10
		}
		// on Broadcom modems, some invalid values are reported as -96 instead of -96.3
		if val > -96 {
			return val
		}
		return -96.3
	}

	interpretBinsData(out, data, groupSize, -96.3, processValueFunc)
}

func guessBinsProfile(bins *models.Bins) {
	if bins.Mode.Type == models.ModeTypeVDSL2 && bins.Mode.Subtype == models.ModeSubtypeUnknown {

		if bins.SNR.Downstream.GroupSize == 16 ||
			bins.QLN.Downstream.GroupSize == 16 ||
			bins.Hlog.Downstream.GroupSize == 16 {

			bins.Mode.Subtype = models.ModeSubtypeProfile35b
		} else {
			bins.Mode.Subtype = models.ModeSubtypeProfile17a
		}
	}
}
