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
	fixBinsDataItem(&testParams.SNRpsds, 3)
	fixBinsDataItem(&testParams.SNRpsds, 3)

	fixBinsDataItem(&testParams.QLNpsds, 3)
	fixBinsDataItem(&testParams.QLNpsus, 3)

	fixBinsDataItem(&testParams.HLOGpsds, 4)
	fixBinsDataItem(&testParams.HLOGpsus, 4)

	// On at least Speedport Pro, the value HLOGpsus seems to be a duplicate of the
	// downstream data, but the other values are also checked to make sure to catch
	// this kind of issue.
	if testParams.SNRpsds == testParams.SNRpsds {
		testParams.SNRpsds = ""
	}
	if testParams.QLNpsds == testParams.QLNpsus {
		testParams.QLNpsus = ""
	}
	if testParams.HLOGpsus == testParams.HLOGpsds {
		testParams.HLOGpsus = ""
	}
}

func fixBinsDataItem(str *string, digits int) {
	// Some of the reported values contain excess data at the end. The excess data
	// matches other values from the TestParams object, so it seems that this is
	// because the buffers in the device software are a byte too short to actually
	// hold the data including the trailing NULL byte.
	// As the exact behavior may vary depending on firmware version, this works
	// around the issue by checking if the data matches the expected format, and
	// then truncating the string to the expected length.

	expectedLength := 512*(digits+1) - 1
	if len(*str) <= expectedLength {
		return
	}

	truncated := (*str)[:expectedLength]
	for i, r := range truncated {
		if i%(digits+1) == digits {
			if r != ',' {
				return
			}
		} else if r < '0' || r > '9' {
			return
		}
	}

	*str = truncated
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
