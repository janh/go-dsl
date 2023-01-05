// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package history

import (
	"errors"
	"time"

	"3e8.eu/go/dsl/models"
)

type ErrorsConfig struct {
	PeriodLength time.Duration
	PeriodCount  int
}

var DefaultErrorsConfig = ErrorsConfig{
	PeriodLength: 5 * time.Minute,
	PeriodCount:  288,
}

type errorsHistoryItem struct {
	DownstreamRTXTXCount models.IntValue
	UpstreamRTXTXCount   models.IntValue

	DownstreamRTXCCount models.IntValue
	UpstreamRTXCCount   models.IntValue

	DownstreamRTXUCCount models.IntValue
	UpstreamRTXUCCount   models.IntValue

	DownstreamFECCount models.IntValue
	UpstreamFECCount   models.IntValue

	DownstreamCRCCount models.IntValue
	UpstreamCRCCount   models.IntValue

	DownstreamESCount models.IntValue
	UpstreamESCount   models.IntValue

	DownstreamSESCount models.IntValue
	UpstreamSESCount   models.IntValue
}

type Errors struct {
	config      ErrorsConfig
	lastTime    time.Time
	lastStatus  models.Status
	periodStart time.Time
	periodIndex int
	data        []errorsHistoryItem
}

func updateErrorValue(out *models.IntValue, lastVal, val models.IntValue) {
	if !lastVal.Valid || !val.Valid {
		return
	}

	if lastVal.Int > val.Int {
		return
	}

	out.Valid = true
	out.Int += val.Int - lastVal.Int
}

func NewErrors(config ErrorsConfig) (*Errors, error) {
	if config.PeriodLength == 0 || config.PeriodCount == 0 {
		return nil, errors.New("period length and count must not be zero")
	}

	h := Errors{config: config}

	return &h, nil
}

func (h *Errors) updatePeriod(now time.Time) {
	periodTime := now
	if !h.lastTime.IsZero() && now.After(h.lastTime) && now.Sub(h.lastTime) <= h.config.PeriodLength {
		periodTime = now.Add(h.lastTime.Sub(now) / 2)
	}
	currentPeriodStart := periodTime.Truncate(h.config.PeriodLength)

	if len(h.data) == 0 || h.periodStart.After(currentPeriodStart) {
		h.data = make([]errorsHistoryItem, h.config.PeriodCount, h.config.PeriodCount)
		h.periodStart = currentPeriodStart
	}

	elapsedPeriodTime := currentPeriodStart.Sub(h.periodStart)
	elapsedPeriods := int(elapsedPeriodTime / h.config.PeriodLength)
	for i := 0; i < elapsedPeriods; i++ {
		h.periodIndex = (h.periodIndex + 1) % h.config.PeriodCount
		h.data[h.periodIndex] = errorsHistoryItem{}
	}

	h.periodStart = currentPeriodStart
}

func (h *Errors) shouldRejectUpdate(status models.Status, now time.Time) bool {
	if now.Sub(h.lastTime) > h.config.PeriodLength {
		return true
	}

	if h.lastStatus.State != models.StateShowtime || status.State != models.StateShowtime {
		return true
	}

	if h.lastStatus.Uptime.Duration > status.Uptime.Duration {
		return true
	}

	return false
}

func (h *Errors) Update(status models.Status, now time.Time) {
	now = now.Round(0)

	defer func() {
		h.lastTime = now
		h.lastStatus = status
	}()

	h.updatePeriod(now)

	if h.shouldRejectUpdate(status, now) {
		return
	}

	currentItem := &h.data[h.periodIndex]

	updateErrorValue(&currentItem.DownstreamRTXTXCount, h.lastStatus.DownstreamRTXTXCount, status.DownstreamRTXTXCount)
	updateErrorValue(&currentItem.UpstreamRTXTXCount, h.lastStatus.UpstreamRTXTXCount, status.UpstreamRTXTXCount)

	updateErrorValue(&currentItem.DownstreamRTXCCount, h.lastStatus.DownstreamRTXCCount, status.DownstreamRTXCCount)
	updateErrorValue(&currentItem.UpstreamRTXCCount, h.lastStatus.UpstreamRTXCCount, status.UpstreamRTXCCount)

	updateErrorValue(&currentItem.DownstreamRTXUCCount, h.lastStatus.DownstreamRTXUCCount, status.DownstreamRTXUCCount)
	updateErrorValue(&currentItem.UpstreamRTXUCCount, h.lastStatus.UpstreamRTXUCCount, status.UpstreamRTXUCCount)

	updateErrorValue(&currentItem.DownstreamFECCount, h.lastStatus.DownstreamFECCount, status.DownstreamFECCount)
	updateErrorValue(&currentItem.UpstreamFECCount, h.lastStatus.UpstreamFECCount, status.UpstreamFECCount)

	updateErrorValue(&currentItem.DownstreamCRCCount, h.lastStatus.DownstreamCRCCount, status.DownstreamCRCCount)
	updateErrorValue(&currentItem.UpstreamCRCCount, h.lastStatus.UpstreamCRCCount, status.UpstreamCRCCount)

	updateErrorValue(&currentItem.DownstreamESCount, h.lastStatus.DownstreamESCount, status.DownstreamESCount)
	updateErrorValue(&currentItem.UpstreamESCount, h.lastStatus.UpstreamESCount, status.UpstreamESCount)

	updateErrorValue(&currentItem.DownstreamSESCount, h.lastStatus.DownstreamSESCount, status.DownstreamSESCount)
	updateErrorValue(&currentItem.UpstreamSESCount, h.lastStatus.UpstreamSESCount, status.UpstreamSESCount)
}

func (h *Errors) Data() (out models.ErrorsHistory) {
	out.EndTime = h.periodStart.Add(h.config.PeriodLength)
	out.PeriodLength = h.config.PeriodLength
	out.PeriodCount = h.config.PeriodCount

	out.DownstreamRTXTXCount = make([]models.IntValue, h.config.PeriodCount, h.config.PeriodCount)
	out.UpstreamRTXTXCount = make([]models.IntValue, h.config.PeriodCount, h.config.PeriodCount)

	out.DownstreamRTXCCount = make([]models.IntValue, h.config.PeriodCount, h.config.PeriodCount)
	out.UpstreamRTXCCount = make([]models.IntValue, h.config.PeriodCount, h.config.PeriodCount)

	out.DownstreamRTXUCCount = make([]models.IntValue, h.config.PeriodCount, h.config.PeriodCount)
	out.UpstreamRTXUCCount = make([]models.IntValue, h.config.PeriodCount, h.config.PeriodCount)

	out.DownstreamFECCount = make([]models.IntValue, h.config.PeriodCount, h.config.PeriodCount)
	out.UpstreamFECCount = make([]models.IntValue, h.config.PeriodCount, h.config.PeriodCount)

	out.DownstreamCRCCount = make([]models.IntValue, h.config.PeriodCount, h.config.PeriodCount)
	out.UpstreamCRCCount = make([]models.IntValue, h.config.PeriodCount, h.config.PeriodCount)

	out.DownstreamESCount = make([]models.IntValue, h.config.PeriodCount, h.config.PeriodCount)
	out.UpstreamESCount = make([]models.IntValue, h.config.PeriodCount, h.config.PeriodCount)

	out.DownstreamSESCount = make([]models.IntValue, h.config.PeriodCount, h.config.PeriodCount)
	out.UpstreamSESCount = make([]models.IntValue, h.config.PeriodCount, h.config.PeriodCount)

	if len(h.data) != h.config.PeriodCount {
		return
	}

	for i := 0; i < h.config.PeriodCount; i++ {
		index := (h.periodIndex + 1 + i) % h.config.PeriodCount

		out.DownstreamRTXTXCount[i] = h.data[index].DownstreamRTXTXCount
		out.UpstreamRTXTXCount[i] = h.data[index].UpstreamRTXTXCount

		out.DownstreamRTXCCount[i] = h.data[index].DownstreamRTXCCount
		out.UpstreamRTXCCount[i] = h.data[index].UpstreamRTXCCount

		out.DownstreamRTXUCCount[i] = h.data[index].DownstreamRTXUCCount
		out.UpstreamRTXUCCount[i] = h.data[index].UpstreamRTXUCCount

		out.DownstreamFECCount[i] = h.data[index].DownstreamFECCount
		out.UpstreamFECCount[i] = h.data[index].UpstreamFECCount

		out.DownstreamCRCCount[i] = h.data[index].DownstreamCRCCount
		out.UpstreamCRCCount[i] = h.data[index].UpstreamCRCCount

		out.DownstreamESCount[i] = h.data[index].DownstreamESCount
		out.UpstreamESCount[i] = h.data[index].UpstreamESCount

		out.DownstreamSESCount[i] = h.data[index].DownstreamSESCount
		out.UpstreamSESCount[i] = h.data[index].UpstreamSESCount
	}

	return
}
