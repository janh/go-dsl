// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lancom

import (
	"fmt"
	"math"
	"strconv"

	"3e8.eu/go/dsl/internal/helpers"
	"3e8.eu/go/dsl/internal/snmp"
	"3e8.eu/go/dsl/models"
)

func parseBins(status *models.Status, values snmp.Values, oidBase string) models.Bins {
	var bins models.Bins

	bins.Mode = status.Mode

	bins.Bits.Downstream = interpretBitloadingTable(values, oidBase+oidAdvancedDsBitLoadingTable)
	bins.Bits.Upstream = interpretBitloadingTable(values, oidBase+oidAdvancedUsBitLoadingTable)

	helpers.GenerateBandsData(&bins)

	bins.SNR.Downstream = interpretSNRTable(values, oidBase+oidAdvancedDsSnrPerSubCarrierTable,
		bins.Mode, bins.Bands.Downstream)

	return bins
}

func interpretBitloadingTable(values snmp.Values, oid string) (out models.BinsBits) {
	interpretTable(values, oid, func(index int, val int64) {
		for len(out.Data) < index+1 {
			out.Data = append(out.Data, 0)
		}

		out.Data[index] = int8(val)
	})

	return
}

func interpretSNRTable(values snmp.Values, oid string, mode models.Mode, bands []models.Band) (out models.BinsFloat) {
	var maxValidSNRIndex int

	interpretTable(values, oid, func(index int, val int64) {
		for len(out.Data) < index+1 {
			out.Data = append(out.Data, 0)
		}

		if val != 255 {
			out.Data[index] = -32 + float64(val)/2
			maxValidSNRIndex = index
		}
	})

	if len(bands) != 0 {
		maxValidBitsIndex := bands[len(bands)-1].End
		out.GroupSize = helpers.GuessSNRGroupSize(maxValidSNRIndex, maxValidBitsIndex, mode.BinCount())
	} else {
		out.GroupSize = 1
	}

	return
}

// Creates OID suffix for table row
func getTableSuffix(digits, row int) string {
	// last item: ASCII codepoint of 'x'
	suffix := "120"

	for i := 0; i < digits-1; i++ {
		digit := row % 10
		row /= 10

		// add ASCII codepoint of a digit
		suffix = strconv.Itoa(0x30+digit) + "." + suffix
	}

	// first item: number of digits
	suffix = strconv.Itoa(digits) + "." + suffix

	return suffix
}

func detectTableDigits(values snmp.Values, oid string) (digits int) {
	for digits = 1; digits < 5; digits++ {
		suffix := getTableSuffix(digits, 0)
		oidFirstItem := fmt.Sprintf("%s.1.1.%s", oid, suffix)

		if _, err := values.GetString(oidFirstItem); err == nil {
			return digits
		}
	}

	return 0
}

func interpretTable(values snmp.Values, oid string, callback func(index int, val int64)) {
	digits := detectTableDigits(values, oid)
	if digits < 1 {
		return
	}

	limit := int(math.Pow10(digits - 1))
	for row := 1; row < limit; row++ {
		rowBase := row * 10
		oidSuffix := getTableSuffix(digits, row)

		for i := 0; i < 10; i++ {
			oidItem := fmt.Sprintf("%s.1.%d.%s", oid, i+2, oidSuffix)

			val, err := values.GetInt64(oidItem)
			if err != nil {
				break
			}

			callback(rowBase+i, val)
		}
	}
}
