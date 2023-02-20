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

func (m *snrMinMax) Reset(snr models.BinsFloat, maxBinCount, periodCount int) {
	m.OriginalGroupSize = snr.GroupSize
	m.OriginalCount = len(snr.Data)

	factor := 1
	minmaxCount := m.OriginalCount
	for minmaxCount > maxBinCount {
		factor *= 2
		minmaxCount = m.OriginalCount / factor
	}

	minmaxGroupSize := snr.GroupSize * factor

	m.Total.GroupSize = minmaxGroupSize

	m.Total.Min = make([]float64, minmaxCount, minmaxCount)
	m.Total.Max = make([]float64, minmaxCount, minmaxCount)

	if periodCount != 0 {
		m.Periods = make([]models.BinsFloatMinMax, periodCount, periodCount)

		for i := range m.Periods {
			m.Periods[i].GroupSize = minmaxGroupSize

			m.Periods[i].Min = make([]float64, minmaxCount, minmaxCount)
			m.Periods[i].Max = make([]float64, minmaxCount, minmaxCount)
		}
	}
}

func (m *snrMinMax) ClearPeriods(startIndex int, count int) {
	periodCount := len(m.Periods)
	minmaxCount := len(m.Total.Min)

	for i := 0; i < count && i < periodCount; i++ {
		idx := (startIndex + i) % periodCount

		for j := 0; j < minmaxCount; j++ {
			m.Periods[idx].Min[j] = 0
			m.Periods[idx].Max[j] = 0
		}
	}
}

func (m *snrMinMax) RecalculateTotal() {
	for i := range m.Total.Min {
		var minTotal, maxTotal float64

		for j := range m.Periods {
			min := m.Periods[j].Min[i]
			if minTotal == 0 {
				minTotal = min
			} else if min != 0 {
				minTotal = math.Min(minTotal, min)
			}

			max := m.Periods[j].Max[i]
			if maxTotal == 0 {
				maxTotal = max
			} else if max != 0 {
				maxTotal = math.Max(maxTotal, max)
			}
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

		if minmax.Min[num] == 0 {
			minmax.Min[num] = val
		} else if val != 0 {
			minmax.Min[num] = math.Min(minmax.Min[num], val)
		}

		if minmax.Max[num] == 0 {
			minmax.Max[num] = val
		} else if val != 0 {
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

	if h.needsReset(bins) || h.periodStart.After(currentPeriodStart) {
		h.mode = bins.Mode

		if h.config.PeriodCount != 0 {
			h.periodStart = currentPeriodStart
		}

		h.snr.Downstream.Reset(bins.SNR.Downstream, h.config.MaxBinCount, h.config.PeriodCount)
		h.snr.Upstream.Reset(bins.SNR.Upstream, h.config.MaxBinCount, h.config.PeriodCount)
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
