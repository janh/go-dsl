// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package history

import (
	"errors"
	"math"
	"time"

	"3e8.eu/go/dsl/models"
)

const (
	minValidSNR   = -32.0
	maxValidSNR   = 95.0
	defaultSNRMin = maxValidSNR + 1
	defaultSNRMax = minValidSNR - 1
)

type BinsConfig struct {
	PeriodLength time.Duration
	PeriodCount  int
	MaxBinCount  int
}

var DefaultBinsConfig = BinsConfig{
	PeriodLength: 1 * time.Hour,
	PeriodCount:  24,
	MaxBinCount:  1024,
}

type snrMinMax struct {
	OriginalGroupSize int
	OriginalCount     int
	Periods           []models.BinsFloatMinMax
	Total             models.BinsFloatMinMax
}

func (m *snrMinMax) resetBinsFloatMinMax(data *models.BinsFloatMinMax, groupSize, count int) {
	data.GroupSize = groupSize

	data.Min = make([]float64, count, count)
	for i := range data.Min {
		data.Min[i] = defaultSNRMin
	}

	data.Max = make([]float64, count, count)
	for i := range data.Max {
		data.Max[i] = defaultSNRMax
	}
}

func (m *snrMinMax) Reset(groupSize, count, maxBinCount, periodCount int) {
	m.OriginalGroupSize = groupSize
	m.OriginalCount = count

	factor := 1
	minmaxCount := m.OriginalCount
	for minmaxCount > maxBinCount {
		factor *= 2
		minmaxCount = (m.OriginalCount + factor - 1) / factor
	}

	minmaxGroupSize := groupSize * factor

	m.resetBinsFloatMinMax(&m.Total, minmaxGroupSize, minmaxCount)

	if periodCount != 0 {
		m.Periods = make([]models.BinsFloatMinMax, periodCount, periodCount)

		for i := range m.Periods {
			m.resetBinsFloatMinMax(&m.Periods[i], minmaxGroupSize, minmaxCount)
		}
	}
}

func (m *snrMinMax) ClearPeriods(startIndex int, count int) {
	periodCount := len(m.Periods)
	minmaxCount := len(m.Total.Min)

	for i := 0; i < count && i < periodCount; i++ {
		idx := (startIndex + i) % periodCount

		for j := 0; j < minmaxCount; j++ {
			m.Periods[idx].Min[j] = defaultSNRMin
			m.Periods[idx].Max[j] = defaultSNRMax
		}
	}
}

func (m *snrMinMax) RecalculateTotal() {
	for i := range m.Total.Min {
		minTotal := defaultSNRMin
		maxTotal := defaultSNRMax

		for j := range m.Periods {
			minTotal = math.Min(minTotal, m.Periods[j].Min[i])
			maxTotal = math.Max(maxTotal, m.Periods[j].Max[i])
		}

		m.Total.Min[i] = minTotal
		m.Total.Max[i] = maxTotal
	}
}

type snrMinMaxDownUp struct {
	Downstream snrMinMax
	Upstream   snrMinMax
}

type Bins struct {
	config      BinsConfig
	mode        models.Mode
	periodStart time.Time
	periodIndex int
	snr         snrMinMaxDownUp
}

func updateBinsFloatMinMax(minmax *models.BinsFloatMinMax, snr models.BinsFloat) {
	if snr.GroupSize == 0 {
		return
	}

	factor := minmax.GroupSize / snr.GroupSize

	for i, val := range snr.Data {
		num := i / factor

		if val >= minValidSNR && val <= maxValidSNR {
			minmax.Min[num] = math.Min(minmax.Min[num], val)
			minmax.Max[num] = math.Max(minmax.Max[num], val)
		}
	}
}

func NewBins(config BinsConfig) (*Bins, error) {
	if (config.PeriodLength != 0) != (config.PeriodCount != 0) {
		return nil, errors.New("either both or neither of period length and count must be zero")
	}

	h := Bins{config: config}

	return &h, nil
}

func (h *Bins) Update(status models.Status, bins models.Bins, now time.Time) {
	currentPeriodStart := now.Truncate(h.config.PeriodLength)

	if status.State != models.StateShowtime {
		return
	}

	if status.Uptime.Valid && status.Uptime.Duration < 1*time.Minute {
		return
	}

	if bins.Mode.Type == models.ModeTypeUnknown {
		return
	}

	if h.needsReset(bins) || h.periodStart.After(currentPeriodStart) {
		h.mode = bins.Mode

		if h.config.PeriodCount != 0 {
			h.periodStart = currentPeriodStart
		}

		h.snr.Downstream.Reset(bins.SNR.Downstream.GroupSize, len(bins.SNR.Downstream.Data),
			h.config.MaxBinCount, h.config.PeriodCount)
		h.snr.Upstream.Reset(bins.SNR.Upstream.GroupSize, len(bins.SNR.Upstream.Data),
			h.config.MaxBinCount, h.config.PeriodCount)
	}

	if h.config.PeriodCount != 0 {
		elapsedPeriodTime := currentPeriodStart.Sub(h.periodStart)
		elapsedPeriods := int(elapsedPeriodTime / h.config.PeriodLength)

		if elapsedPeriods > 0 {
			h.snr.Downstream.ClearPeriods(h.periodIndex+1, elapsedPeriods)
			h.snr.Upstream.ClearPeriods(h.periodIndex+1, elapsedPeriods)

			h.snr.Downstream.RecalculateTotal()
			h.snr.Upstream.RecalculateTotal()

			h.periodStart = currentPeriodStart
			h.periodIndex = (h.periodIndex + elapsedPeriods) % h.config.PeriodCount
		}

		updateBinsFloatMinMax(&h.snr.Downstream.Periods[h.periodIndex], bins.SNR.Downstream)
		updateBinsFloatMinMax(&h.snr.Upstream.Periods[h.periodIndex], bins.SNR.Upstream)
	}

	updateBinsFloatMinMax(&h.snr.Downstream.Total, bins.SNR.Downstream)
	updateBinsFloatMinMax(&h.snr.Upstream.Total, bins.SNR.Upstream)
}

func (h *Bins) needsReset(bins models.Bins) bool {
	if h.mode != bins.Mode {
		return true
	}

	if (h.snr.Downstream.OriginalGroupSize != bins.SNR.Downstream.GroupSize ||
		h.snr.Downstream.OriginalCount != len(bins.SNR.Downstream.Data)) &&
		!(h.snr.Downstream.OriginalGroupSize != 0 && bins.SNR.Downstream.GroupSize == 0) {

		return true
	}

	if (h.snr.Upstream.OriginalGroupSize != bins.SNR.Upstream.GroupSize ||
		h.snr.Upstream.OriginalCount != len(bins.SNR.Upstream.Data)) &&
		!(h.snr.Upstream.OriginalGroupSize != 0 && bins.SNR.Upstream.GroupSize == 0) {

		return true
	}

	return false
}

func copyBinsFloatMinMax(dst *models.BinsFloatMinMax, src *models.BinsFloatMinMax) {
	dst.GroupSize = src.GroupSize

	dst.Min = make([]float64, len(src.Min), len(src.Min))
	copy(dst.Min, src.Min)

	dst.Max = make([]float64, len(src.Max), len(src.Max))
	copy(dst.Max, src.Max)
}

func (h *Bins) Data() (out models.BinsHistory) {
	copyBinsFloatMinMax(&out.SNR.Downstream, &h.snr.Downstream.Total)
	copyBinsFloatMinMax(&out.SNR.Upstream, &h.snr.Upstream.Total)

	return
}
