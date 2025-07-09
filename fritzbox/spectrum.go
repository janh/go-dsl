// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package fritzbox

import (
	"encoding/json"
	"strconv"
	"strings"

	"3e8.eu/go/dsl/internal/helpers"
	"3e8.eu/go/dsl/models"
)

type spectrumDataWrapper struct {
	Data struct {
		Ports []json.RawMessage `json:"ports"`
	} `json:"data"`
}

type spectrumDataWrapperLegacy struct {
	Ports []json.RawMessage `json:"port"`
}

type number int

func (n *number) UnmarshalJSON(data []byte) (err error) {
	if len(data) > 0 && data[0] == '"' {
		var s string
		err = json.Unmarshal(data, &s)
		if err != nil {
			return err
		}
		var i int
		i, err = strconv.Atoi(strings.TrimSpace(s))
		*n = number(i)
		return err
	}

	var i int
	err = json.Unmarshal(data, &i)
	*n = number(i)
	return
}

type spectrumData struct {
	BitBandconfig         []spectrumDataBandItem `json:"BIT_BANDCONFIG"`
	BitUpstreamBandconfig []spectrumDataBandItem `json:"BIT_US_BANDCONFIG"`

	PilotValues []number `json:"PILOT_VALUES"`
	Pilot       number   `json:"PILOT"`

	TonesPerBATValue number   `json:"TONES_PER_BAT_VALUE"`
	MaxBATTone       number   `json:"MAX_BAT_TONE"`
	ActualBITValues  []number `json:"ACT_BIT_VALUES"`

	TonesPerSNRValue number   `json:"TONES_PER_SNR_VALUE"`
	MaxSNRTone       number   `json:"MAX_SNR_TONE"`
	ActualSNRValues  []number `json:"ACT_SNR_VALUES"`
}

type spectrumDataBandItem struct {
	First number `json:"FIRST"`
	Last  number `json:"LAST"`
}

func parseSpectrum(bins *models.Bins, status *models.Status, d *rawDataSpectrum) {
	bins.Mode = status.Mode

	var portList []json.RawMessage
	if !d.Legacy {
		var dataWrapper spectrumDataWrapper
		json.Unmarshal([]byte(d.Data), &dataWrapper)
		portList = dataWrapper.Data.Ports
	} else {
		var dataWrapper spectrumDataWrapperLegacy
		json.Unmarshal([]byte(d.Data), &dataWrapper)
		portList = dataWrapper.Ports
	}

	if len(portList) < 1 {
		return
	}

	data := portList[0]

	var directions map[string]json.RawMessage
	json.Unmarshal(data, &directions)
	if dataUpstream, ok := directions["us"]; ok {
		data = dataUpstream
	}

	var portData spectrumData
	err := json.Unmarshal(data, &portData)
	if err != nil {
		return
	}

	if len(portData.BitBandconfig) == 0 && len(portData.BitUpstreamBandconfig) > 0 {
		portData.BitBandconfig = portData.BitUpstreamBandconfig
	}

	processPilotTones(&bins.PilotTones, int(portData.TonesPerBATValue), portData.PilotValues, portData.Pilot)
	processSpectrumBits(&bins.Bits, portData.BitBandconfig, int(portData.TonesPerBATValue), int(portData.MaxBATTone), portData.ActualBITValues)
	processSpectrumSNR(&bins.SNR, portData.BitBandconfig, int(portData.TonesPerSNRValue), int(portData.MaxSNRTone), portData.ActualSNRValues)

	helpers.GenerateBandsData(bins)
}

func isBinUpstream(usBands []spectrumDataBandItem, bin int) bool {
	for _, band := range usBands {
		if bin >= int(band.First) && bin <= int(band.Last) {
			return true
		}
	}
	return false
}

func processPilotTones(tones *[]int, groupSize int, values []number, val number) {
	*tones = make([]int, len(values))

	for i, val := range values {
		(*tones)[i] = int(val)*groupSize + groupSize/2
	}

	if len(*tones) == 0 && val != 0 {
		*tones = append(*tones, int(val)*groupSize+groupSize/2)
	}
}

func processSpectrumBits(bits *models.BinsBitsDownUp, usBands []spectrumDataBandItem, groupSize, binCount int, values []number) {
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

func initSNRData(data *models.BinsFloat, binCount, groupSize int) {
	data.Data = make([]float64, binCount/groupSize)
	for i := range data.Data {
		data.Data[i] = -32.5
	}

	data.GroupSize = groupSize
}

func processSpectrumSNR(snr *models.BinsFloatDownUp, usBands []spectrumDataBandItem, groupSize, binCount int, values []number) {
	if groupSize == 0 {
		return
	}

	if binCount == 0 {
		binCount = len(values) * groupSize
	}

	initSNRData(&snr.Downstream, binCount, groupSize)
	initSNRData(&snr.Upstream, binCount, groupSize)

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
