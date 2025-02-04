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

	"3e8.eu/go/dsl/internal/helpers"
	"3e8.eu/go/dsl/models"
)

const binsStorageVersion = 1

type storageBinsConfig struct {
	PeriodLength int64
	PeriodCount  int64
	MaxBinCount  int64
}

type storageBinsHeader struct {
	Mode            [64]byte
	PeriodStartSec  int64
	PeriodStartNsec uint32
}

type storageBinsSNRMinMaxHeader struct {
	OriginalGroupSize uint32
	OriginalCount     uint32
}

func (h *Bins) writeBinsFloatMinMax(w io.Writer, data models.BinsFloatMinMax) error {
	for _, val := range data.Min {
		err := binary.Write(w, binary.BigEndian, val)
		if err != nil {
			return err
		}
	}

	for _, val := range data.Max {
		err := binary.Write(w, binary.BigEndian, val)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *Bins) writeBinsSNRMinMax(w io.Writer, data snrMinMax) error {
	// Write header

	header := storageBinsSNRMinMaxHeader{
		OriginalGroupSize: uint32(data.OriginalGroupSize),
		OriginalCount:     uint32(data.OriginalCount),
	}

	err := binary.Write(w, binary.BigEndian, header)
	if err != nil {
		return err
	}

	// Write period data or total data, if available

	if data.OriginalGroupSize == 0 || data.OriginalCount == 0 {
		return nil
	}

	if h.config.PeriodCount > 0 {
		for i := 0; i < h.config.PeriodCount; i++ {
			index := (h.periodIndex + 1 + i) % h.config.PeriodCount
			err = h.writeBinsFloatMinMax(w, data.Periods[index])
			if err != nil {
				return err
			}
		}
	} else {
		err = h.writeBinsFloatMinMax(w, data.Total)
		if err != nil {
			return err
		}
	}

	return nil
}

// Save serializes the current state in an opaque binary format.
func (h *Bins) Save(w io.Writer) error {
	// Write main header

	mainHeader := storageMainHeader{
		Version:      binsStorageVersion,
		CreationTime: time.Now().Unix(),
	}

	err := binary.Write(w, binary.BigEndian, mainHeader)
	if err != nil {
		return fmt.Errorf("failed to write main header: %w", err)
	}

	// Write config

	binsConfig := storageBinsConfig{
		PeriodLength: h.config.PeriodLength.Nanoseconds(),
		PeriodCount:  int64(h.config.PeriodCount),
		MaxBinCount:  int64(h.config.MaxBinCount),
	}

	err = binary.Write(w, binary.BigEndian, binsConfig)
	if err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Write header

	var modeString [64]byte
	copy(modeString[:], h.mode.String())
	header := storageBinsHeader{
		Mode:            modeString,
		PeriodStartSec:  int64(h.periodStart.Unix()),
		PeriodStartNsec: uint32(h.periodStart.Nanosecond()),
	}

	err = binary.Write(w, binary.BigEndian, header)
	if err != nil {
		return fmt.Errorf("failed to write heade: %w", err)
	}

	// Write SNR data

	err = h.writeBinsSNRMinMax(w, h.snr.Downstream)
	if err != nil {
		return fmt.Errorf("failed to write downstream SNR data: %w", err)
	}

	err = h.writeBinsSNRMinMax(w, h.snr.Upstream)
	if err != nil {
		return fmt.Errorf("failed to write upstream SNR data: %w", err)
	}

	return nil
}

func (h *Bins) readBinsFloatMinMax(r io.Reader, data *models.BinsFloatMinMax) error {
	for i := range data.Min {
		err := binary.Read(r, binary.BigEndian, &data.Min[i])
		if err != nil {
			return err
		}
	}

	for i := range data.Max {
		err := binary.Read(r, binary.BigEndian, &data.Max[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *Bins) readBinsSNRMinMax(r io.Reader, data *snrMinMax) error {
	// Read header

	var header storageBinsSNRMinMaxHeader
	err := binary.Read(r, binary.BigEndian, &header)
	if err != nil {
		return err
	}

	// Ensure that data structures are initialized

	data.Reset(int(header.OriginalGroupSize), int(header.OriginalCount),
		h.config.MaxBinCount, h.config.PeriodCount)

	// Read data if available

	if header.OriginalGroupSize == 0 || header.OriginalCount == 0 {
		return nil
	}

	if h.config.PeriodCount > 0 {
		for i := 0; i < h.config.PeriodCount; i++ {
			err = h.readBinsFloatMinMax(r, &data.Periods[i])
			if err != nil {
				return err
			}
		}
		data.RecalculateTotal()
	} else {
		err = h.readBinsFloatMinMax(r, &data.Total)
		if err != nil {
			return err
		}
	}

	return nil
}

// Load loads a serialized state. The config parameters need to match the current instance.
func (h *Bins) Load(r io.Reader) error {
	// Read and verify main header

	var mainHeader storageMainHeader
	err := binary.Read(r, binary.BigEndian, &mainHeader)
	if err != nil {
		return fmt.Errorf("failed to read main header: %w", err)
	}

	if mainHeader.Version != binsStorageVersion {
		return fmt.Errorf("unsupported data version %d", mainHeader.Version)
	}

	if mainHeader.CreationTime > time.Now().Unix() {
		return errors.New("creation time in future")
	}

	// Read and check config

	var binsConfig storageBinsConfig
	err = binary.Read(r, binary.BigEndian, &binsConfig)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	config := BinsConfig{
		PeriodLength: time.Duration(binsConfig.PeriodLength) * time.Nanosecond,
		PeriodCount:  int(binsConfig.PeriodCount),
		MaxBinCount:  int(binsConfig.MaxBinCount),
	}

	if config != h.config {
		return errors.New("config does not match")
	}

	// Read and verify header

	var binsHeader storageBinsHeader
	err = binary.Read(r, binary.BigEndian, &binsHeader)
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	newHistory := Bins{
		config:      config,
		mode:        helpers.ParseMode(string(binsHeader.Mode[:])),
		periodStart: time.Unix(binsHeader.PeriodStartSec, int64(binsHeader.PeriodStartNsec)),
		periodIndex: 0,
	}

	if newHistory.periodStart.Unix() > mainHeader.CreationTime {
		return fmt.Errorf("period start time after creation time: %s", newHistory.periodStart.String())
	}

	if !newHistory.periodStart.Equal(newHistory.periodStart.Truncate(h.config.PeriodLength)) {
		return fmt.Errorf("implausible period start time: %s", newHistory.periodStart.String())
	}

	if config.PeriodCount > 0 {
		newHistory.periodIndex = config.PeriodCount - 1
	}

	// Read SNR data

	err = newHistory.readBinsSNRMinMax(r, &newHistory.snr.Downstream)
	if err != nil {
		fmt.Errorf("failed to read downstream SNR data: %w", err)
	}

	err = newHistory.readBinsSNRMinMax(r, &newHistory.snr.Upstream)
	if err != nil {
		fmt.Errorf("failed to read upstream SNR data: %w", err)
	}

	*h = newHistory

	err = checkEndOfFile(r)
	return err
}
