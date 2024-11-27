// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package history

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"

	"3e8.eu/go/dsl/models"
)

const errorsStorageVersion = 1

type storageErrorsConfig struct {
	PeriodLength int64
	PeriodCount  int64
}

type storageErrorsHeader struct {
	PeriodStartSec  int64
	PeriodStartNsec uint32
}

func (h *Errors) writeErrorsBoolValue(w io.Writer, val models.BoolValue) error {
	err := binary.Write(w, binary.BigEndian, val)
	return err
}

func (h *Errors) writeErrorsIntValues(w io.Writer, values ...models.IntValue) error {
	for _, val := range values {
		err := binary.Write(w, binary.BigEndian, val)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *Errors) writeErrorsItem(w io.Writer, item errorsHistoryItem) error {
	err := h.writeErrorsBoolValue(w, item.Showtime)
	if err != nil {
		return err
	}

	err = h.writeErrorsIntValues(w, item.DownstreamRTXTXCount, item.UpstreamRTXTXCount)
	if err != nil {
		return err
	}

	err = h.writeErrorsIntValues(w, item.DownstreamRTXCCount, item.UpstreamRTXCCount)
	if err != nil {
		return err
	}

	err = h.writeErrorsIntValues(w, item.DownstreamRTXUCCount, item.UpstreamRTXUCCount)
	if err != nil {
		return err
	}

	err = h.writeErrorsIntValues(w, item.DownstreamFECCount, item.UpstreamFECCount)
	if err != nil {
		return err
	}

	err = h.writeErrorsIntValues(w, item.DownstreamCRCCount, item.UpstreamCRCCount)
	if err != nil {
		return err
	}

	err = h.writeErrorsIntValues(w, item.DownstreamESCount, item.UpstreamESCount)
	if err != nil {
		return err
	}

	err = h.writeErrorsIntValues(w, item.DownstreamSESCount, item.UpstreamSESCount)
	if err != nil {
		return err
	}

	return nil
}

// Save serializes the current state in an opaque binary format.
func (h *Errors) Save(w io.Writer) error {
	// Write main header

	mainHeader := storageMainHeader{
		Version:      errorsStorageVersion,
		CreationTime: time.Now().Unix(),
	}

	err := binary.Write(w, binary.BigEndian, mainHeader)
	if err != nil {
		return fmt.Errorf("failed to write main header: %w", err)
	}

	// Write config

	errorsConfig := storageErrorsConfig{
		PeriodLength: h.config.PeriodLength.Nanoseconds(),
		PeriodCount:  int64(h.config.PeriodCount),
	}

	err = binary.Write(w, binary.BigEndian, errorsConfig)
	if err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Write header

	header := storageErrorsHeader{
		PeriodStartSec:  int64(h.periodStart.Unix()),
		PeriodStartNsec: uint32(h.periodStart.Nanosecond()),
	}

	err = binary.Write(w, binary.BigEndian, header)
	if err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write data

	for i := 0; i < h.config.PeriodCount; i++ {
		var item errorsHistoryItem

		// Get item if data is already populated, otherwise write empty item
		if len(h.data) == h.config.PeriodCount {
			index := (h.periodIndex + 1 + i) % h.config.PeriodCount
			item = h.data[index]
		}

		err = h.writeErrorsItem(w, item)
		if err != nil {
			return fmt.Errorf("failed to write errors item: %w", err)
		}
	}

	return nil
}

func (h *Errors) readErrorsBoolValue(r io.Reader, val *models.BoolValue) error {
	err := binary.Read(r, binary.BigEndian, val)
	return err
}

func (h *Errors) readErrorsIntValues(r io.Reader, values ...*models.IntValue) error {
	for _, val := range values {
		err := binary.Read(r, binary.BigEndian, val)
		if err != nil {
			return err
		}
		if val.Int < 0 {
			return errors.New("unexpected negative value")
		}
	}
	return nil
}

func (h *Errors) readErrorsItem(r io.Reader, item *errorsHistoryItem) error {
	err := h.readErrorsBoolValue(r, &item.Showtime)
	if err != nil {
		return err
	}

	err = h.readErrorsIntValues(r, &item.DownstreamRTXTXCount, &item.UpstreamRTXTXCount)
	if err != nil {
		return err
	}

	err = h.readErrorsIntValues(r, &item.DownstreamRTXCCount, &item.UpstreamRTXCCount)
	if err != nil {
		return err
	}

	err = h.readErrorsIntValues(r, &item.DownstreamRTXUCCount, &item.UpstreamRTXUCCount)
	if err != nil {
		return err
	}

	err = h.readErrorsIntValues(r, &item.DownstreamFECCount, &item.UpstreamFECCount)
	if err != nil {
		return err
	}

	err = h.readErrorsIntValues(r, &item.DownstreamCRCCount, &item.UpstreamCRCCount)
	if err != nil {
		return err
	}

	err = h.readErrorsIntValues(r, &item.DownstreamESCount, &item.UpstreamESCount)
	if err != nil {
		return err
	}

	err = h.readErrorsIntValues(r, &item.DownstreamSESCount, &item.UpstreamSESCount)
	if err != nil {
		return err
	}

	return nil
}

// Load loads a serialized state. The config parameters need to match the current instance.
func (h *Errors) Load(r io.Reader) error {
	// Read and verify main header

	var mainHeader storageMainHeader
	err := binary.Read(r, binary.BigEndian, &mainHeader)
	if err != nil {
		return fmt.Errorf("failed to read main header: %w", err)
	}

	if mainHeader.Version != errorsStorageVersion {
		return fmt.Errorf("unsupported data version %d", mainHeader.Version)
	}

	if mainHeader.CreationTime > time.Now().Unix() {
		return errors.New("creation time in future")
	}

	// Read and check config

	var errorsConfig storageErrorsConfig
	err = binary.Read(r, binary.BigEndian, &errorsConfig)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	config := ErrorsConfig{
		PeriodLength: time.Duration(errorsConfig.PeriodLength) * time.Nanosecond,
		PeriodCount:  int(errorsConfig.PeriodCount),
	}

	if config != h.config {
		return errors.New("config does not match")
	}

	// Read and verify header

	var errorsHeader storageErrorsHeader
	err = binary.Read(r, binary.BigEndian, &errorsHeader)
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	newHistory := Errors{
		config:      config,
		periodStart: time.Unix(errorsHeader.PeriodStartSec, int64(errorsHeader.PeriodStartNsec)),
		periodIndex: config.PeriodCount - 1,
		data:        make([]errorsHistoryItem, config.PeriodCount, config.PeriodCount),
	}

	if newHistory.periodStart.Unix() > mainHeader.CreationTime {
		return fmt.Errorf("period start time after creation time: %s", newHistory.periodStart.String())
	}

	if !newHistory.periodStart.Equal(newHistory.periodStart.Truncate(h.config.PeriodLength)) {
		return fmt.Errorf("implausible period start time: %s", newHistory.periodStart.String())
	}

	// Read data

	for i := 0; i < config.PeriodCount; i++ {
		err = newHistory.readErrorsItem(r, &newHistory.data[i])
		if err != nil {
			return fmt.Errorf("failed to read errors item: %w", err)
		}
	}

	*h = newHistory

	err = checkEndOfFile(r)
	return err
}
