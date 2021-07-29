// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package helpers

import (
	"3e8.eu/go/dsl/models"
)

func generateBandsDataADSL(bins *models.Bins) {
	if bins.Mode.Subtype == models.ModeSubtypeUnknown {
		return
	}

	bins.Bands.Upstream = make([]models.Band, 1)
	bins.Bands.Downstream = make([]models.Band, 1)

	switch bins.Mode.Subtype {
	case models.ModeSubtypeAnnexA, models.ModeSubtypeAnnexL, models.ModeSubtypeAnnexM:
		bins.Bands.Upstream[0].Start = 6
	case models.ModeSubtypeAnnexB:
		bins.Bands.Upstream[0].Start = 32
	case models.ModeSubtypeAnnexI, models.ModeSubtypeAnnexJ:
		bins.Bands.Upstream[0].Start = 1
	}

	switch bins.Mode.Subtype {
	case models.ModeSubtypeAnnexA, models.ModeSubtypeAnnexI, models.ModeSubtypeAnnexL:
		bins.Bands.Upstream[0].End = 31
	case models.ModeSubtypeAnnexB, models.ModeSubtypeAnnexJ, models.ModeSubtypeAnnexM:
		bins.Bands.Upstream[0].End = 63
	}

	bins.Bands.Downstream[0].Start = bins.Bands.Upstream[0].End + 1
	if bins.Mode.Subtype == models.ModeSubtypeAnnexL {
		bins.Bands.Downstream[0].End = 127
	} else if bins.Mode.Type == models.ModeTypeADSL2Plus {
		bins.Bands.Downstream[0].End = 511
	} else {
		bins.Bands.Downstream[0].End = 255
	}
}

func generateBandsDataFromBitloading(bins *models.Bins) {
	if len(bins.Bits.Downstream.Data) != bins.Mode.BinCount() || len(bins.Bits.Upstream.Data) != bins.Mode.BinCount() {
		return
	}

	var isValid, lastValid, isDownstream, lastDownstream bool
	var firstValidIndex, lastValidIndex int

	for i := 0; i < bins.Mode.BinCount(); i++ {
		bitsDS := bins.Bits.Downstream.Data[i]
		bitsUS := bins.Bits.Upstream.Data[i]

		isValid = bitsDS > 0 || bitsUS > 0

		if isValid {
			if bitsDS > 0 {
				isDownstream = true
			} else if bitsUS > 0 {
				isDownstream = false
			}

			if isDownstream != lastDownstream && lastValid {
				band := models.Band{Start: firstValidIndex, End: lastValidIndex}
				if lastDownstream {
					bins.Bands.Downstream = append(bins.Bands.Downstream, band)
				} else {
					bins.Bands.Upstream = append(bins.Bands.Upstream, band)
				}
			}

			if isDownstream != lastDownstream || !lastValid {
				firstValidIndex = i
			}

			lastValidIndex = i
			lastValid = isValid
			lastDownstream = isDownstream
		}
	}

	if lastValid {
		band := models.Band{Start: firstValidIndex, End: lastValidIndex}
		if lastDownstream {
			bins.Bands.Downstream = append(bins.Bands.Downstream, band)
		} else {
			bins.Bands.Upstream = append(bins.Bands.Upstream, band)
		}
	}
}

func GenerateBandsData(bins *models.Bins) {
	if len(bins.Bands.Downstream) != 0 || len(bins.Bands.Upstream) != 0 {
		return
	}

	if bins.Mode.Type == models.ModeTypeADSL || bins.Mode.Type == models.ModeTypeADSL2 || bins.Mode.Type == models.ModeTypeADSL2Plus {
		generateBandsDataADSL(bins)
	} else {
		generateBandsDataFromBitloading(bins)
	}
}
