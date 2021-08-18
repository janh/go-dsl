// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package fritzbox

import (
	"encoding/json"

	"3e8.eu/go/dsl/internal/helpers"
	"3e8.eu/go/dsl/models"
)

type spectrumData struct {
	Ports []map[string]struct {
		BitBandconfig []spectrumDataBandItem `json:"BIT_BANDCONFIG"`

		PilotValues []int `json:"PILOT_VALUES"`
		Pilot       int   `json:"PILOT"`

		TonesPerBATValue int   `json:"TONES_PER_BAT_VALUE"`
		MaxBATTone       int   `json:"MAX_BAT_TONE"`
		ActualBITValues  []int `json:"ACT_BIT_VALUES"`

		TonesPerSNRValue int   `json:"TONES_PER_SNR_VALUE"`
		MaxSNRTone       int   `json:"MAX_SNR_TONE"`
		ActualSNRValues  []int `json:"ACT_SNR_VALUES"`
	} `json:"port"`
}

type spectrumDataBandItem struct {
	First int `json:"first"`
	Last  int `json:"last"`
}

func parseSpectrum(bins *models.Bins, status *models.Status, dslSpectrum string) {
	bins.Mode = status.Mode

	var data spectrumData
	json.Unmarshal([]byte(dslSpectrum), &data)

	if len(data.Ports) < 1 {
		return
	}

	if portData, ok := data.Ports[0]["us"]; ok {
		processPilotTones(&bins.PilotTones, portData.TonesPerBATValue, portData.PilotValues, portData.Pilot)
		processSpectrumBits(&bins.Bits, portData.BitBandconfig, portData.TonesPerBATValue, portData.MaxBATTone, portData.ActualBITValues)
		processSpectrumSNR(&bins.SNR, portData.BitBandconfig, portData.TonesPerSNRValue, portData.MaxSNRTone, portData.ActualSNRValues)
	}

	helpers.GenerateBandsData(bins)
}

func isBinUpstream(usBands []spectrumDataBandItem, bin int) bool {
	for _, band := range usBands {
		if bin >= band.First && bin <= band.Last {
			return true
		}
	}
	return false
}

func processPilotTones(tones *[]int, groupSize int, values []int, val int) {
	*tones = make([]int, len(values))

	for i, val := range values {
		(*tones)[i] = val*groupSize + groupSize/2
	}

	if len(*tones) == 0 && val != 0 {
		*tones = append(*tones, val*groupSize+groupSize/2)
	}
}

func processSpectrumBits(bits *models.BinsBitsDownUp, usBands []spectrumDataBandItem, groupSize, binCount int, values []int) {
	if groupSize == 0 {
		return
	}

	if binCount == 0 {
		binCount = len(values) * groupSize
	}

	bits.Downstream.Data = make([]int8, binCount)
	bits.Upstream.Data = make([]int8, binCount)

	for i, val := range values {
		if val != 0 {
			var currentBits *models.BinsBits
			if isBinUpstream(usBands, i) {
				currentBits = &bits.Upstream
			} else {
				currentBits = &bits.Downstream
			}

			numBase := i * groupSize
			for num := numBase; num < numBase+groupSize; num += 1 {
				currentBits.Data[num] = int8(val)
			}
		}
	}
}

func processSpectrumSNR(snr *models.BinsFloatDownUp, usBands []spectrumDataBandItem, groupSize, binCount int, values []int) {
	if groupSize == 0 {
		return
	}

	if binCount == 0 {
		binCount = len(values) * groupSize
	}

	snr.Downstream.Data = make([]float64, binCount/groupSize)
	snr.Downstream.GroupSize = groupSize

	snr.Upstream.Data = make([]float64, binCount/groupSize)
	snr.Upstream.GroupSize = groupSize

	for num, val := range values {
		if val != 0 {
			var currentSNR *models.BinsFloat
			if isBinUpstream(usBands, num) {
				currentSNR = &snr.Upstream
			} else {
				currentSNR = &snr.Downstream
			}

			currentSNR.Data[num] = float64(val) / 2
		}
	}
}
