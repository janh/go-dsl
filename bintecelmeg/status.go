// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package bintecelmeg

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"3e8.eu/go/dsl/internal/helpers"
	"3e8.eu/go/dsl/models"
)

var regexpDuration = regexp.MustCompile(`([0-9]+) ([0-9]{2}):([0-9]{2}):([0-9]{2})`)

func interpretStatusState(values map[string]string, key string) (out models.State) {
	if val, ok := values[key]; ok {
		switch val {
		case "Idle":
			return models.StateDown
		case "Handshake":
			return models.StateInitHandshake
		case "Training":
			return models.StateInitTraining
		case "Syncing":
			return models.StateInit
		case "Showtime":
			return models.StateShowtime
		}
	}

	return
}

func interpretStatusMode(values map[string]string, key string) (out models.Mode) {
	if val, ok := values[key]; ok {
		out = helpers.ParseMode(val)
	}

	return
}

func interpretStatusDuration(values map[string]string, key string) (out models.Duration) {
	if val, ok := values[key]; ok {
		if matches := regexpDuration.FindStringSubmatch(val); len(matches) > 0 {
			days, _ := strconv.ParseInt(matches[1], 10, 64)
			hours, _ := strconv.ParseInt(matches[2], 10, 64)
			minutes, _ := strconv.ParseInt(matches[3], 10, 64)
			seconds, _ := strconv.ParseInt(matches[4], 10, 64)

			out.Duration = time.Duration(days)*24*time.Hour +
				time.Duration(hours)*time.Hour +
				time.Duration(minutes)*time.Minute +
				time.Duration(seconds)*time.Second

			out.Valid = true
		}
	}

	return
}

func interpretStatusInventory(values map[string]string, keyVendor, keyVersion string) (out models.Inventory) {
	if val, ok := values[keyVendor]; ok {
		if strings.HasPrefix(val, "0x") && len(val) >= 10 {
			vendor := helpers.ParseHexadecimal(val[2:10])
			out.Vendor = helpers.FormatVendor(string(vendor))
		}
	}

	if val, ok := values[keyVersion]; ok {
		if strings.HasPrefix(val, "0x") && len(val) >= 6 {
			version := helpers.ParseHexadecimal(val[2:6])
			version[0], version[1] = version[1], version[0]
			out.Version = helpers.FormatVersion(out.Vendor, version)
		}
	}

	return
}

func interpretStatusDelay(values map[string]string, keyDepth, keyDelay string) (out models.FloatValue) {
	if strings.Contains(values[keyDepth], "FAST") {
		out.Float = 0.0
		out.Valid = true
	} else {
		out = interpretStatusFloatValueSuffix(values, keyDelay, " ms")
	}

	return
}

func interpretStatusIntValueSuffix(values map[string]string, key string, suffix string) (out models.IntValue) {
	if val, ok := values[key]; ok {
		if strings.HasSuffix(val, suffix) {
			val := val[:len(val)-len(suffix)]
			if valInt, err := strconv.ParseInt(val, 10, 64); err == nil {
				out.Int = valInt
				out.Valid = true
			}
		}
	}

	return
}

func interpretStatusIntValue(values map[string]string, key string) (out models.IntValue) {
	if val, ok := values[key]; ok {
		if valInt, err := strconv.ParseInt(val, 10, 64); err == nil {
			out.Int = valInt
			out.Valid = true
		}
	}

	return
}

func interpretStatusFloatValueSuffix(values map[string]string, key string, suffix string) (out models.FloatValue) {
	if val, ok := values[key]; ok {
		if strings.HasSuffix(val, suffix) {
			val := val[:len(val)-len(suffix)]
			if valFloat, err := strconv.ParseFloat(val, 64); err == nil {
				out.Float = valFloat
				out.Valid = true
			}
		}
	}

	return
}

func interpretStatusFloatValue(values map[string]string, key string) (out models.FloatValue) {
	if val, ok := values[key]; ok {
		if valFloat, err := strconv.ParseFloat(val, 64); err == nil {
			out.Float = valFloat
			out.Valid = true
		}
	}

	return
}

func interpretStatus(connection, localModem, remoteModem, receiveStatistics, transmitStatistics map[string]string) models.Status {
	var status models.Status

	status.State = interpretStatusState(connection, "state")
	status.Mode = interpretStatusMode(connection, "trainedmode")
	if status.State == models.StateShowtime {
		status.Uptime = interpretStatusDuration(connection, "lastchange")
	}

	status.NearEndInventory = interpretStatusInventory(localModem, "ituvendorid", "ituvendorspecific")
	if val, ok := localModem["versionnumber"]; ok {
		status.NearEndInventory.Version = val
	}
	status.FarEndInventory = interpretStatusInventory(remoteModem, "ituvendorid", "ituvendorspecific")

	status.DownstreamActualRate.IntValue = interpretStatusIntValueSuffix(receiveStatistics, "dsbitratefastpath", " kbps")
	if status.DownstreamActualRate.IntValue.Int == 0 {
		status.DownstreamActualRate.IntValue = interpretStatusIntValueSuffix(receiveStatistics, "dsbitrateinterleaved", " kbps")
	}
	status.UpstreamActualRate.IntValue = interpretStatusIntValueSuffix(transmitStatistics, "usbitratefastpath", " kbps")
	if status.UpstreamActualRate.IntValue.Int == 0 {
		status.UpstreamActualRate.IntValue = interpretStatusIntValueSuffix(transmitStatistics, "usbitrateinterleaved", " kbps")
	}

	status.DownstreamAttainableRate.IntValue = interpretStatusIntValueSuffix(receiveStatistics, "dsattainablerate", " kbps")
	status.UpstreamAttainableRate.IntValue = interpretStatusIntValueSuffix(transmitStatistics, "usattainablerate", " kbps")

	status.DownstreamBitswap.Executed = interpretStatusIntValue(receiveStatistics, "dsbitswapcount")
	status.DownstreamBitswap.Normalize()
	status.UpstreamBitswap.Executed = interpretStatusIntValue(transmitStatistics, "usbitswapcount")
	status.UpstreamBitswap.Normalize()

	status.DownstreamInterleavingDelay.FloatValue = interpretStatusDelay(receiveStatistics, "dsinterleaverdepth", "dsinterleaverdelay")
	status.UpstreamInterleavingDelay.FloatValue = interpretStatusDelay(transmitStatistics, "usinterleaverdepth", "usinterleaverdelay")

	status.DownstreamImpulseNoiseProtection.FloatValue = interpretStatusFloatValue(receiveStatistics, "dsinp")
	status.UpstreamImpulseNoiseProtection.FloatValue = interpretStatusFloatValue(transmitStatistics, "usinp")

	status.DownstreamAttenuation.FloatValue = interpretStatusFloatValueSuffix(receiveStatistics, "dsattenuation", " dB")
	status.UpstreamAttenuation.FloatValue = interpretStatusFloatValueSuffix(transmitStatistics, "usattenuation", " dB")

	status.DownstreamSNRMargin.FloatValue = interpretStatusFloatValueSuffix(receiveStatistics, "dsnoisemargin", " dB")
	status.UpstreamSNRMargin.FloatValue = interpretStatusFloatValueSuffix(transmitStatistics, "usnoisemargin", " dB")

	status.DownstreamPower.FloatValue = interpretStatusFloatValueSuffix(receiveStatistics, "dsoutputpower", " dBm")
	status.UpstreamPower.FloatValue = interpretStatusFloatValueSuffix(transmitStatistics, "usoutputpower", " dBm")
	// As typical for some Lantiq devices, the transmit power values may be swapped for VDSL2
	if status.Mode.Type == models.ModeTypeVDSL2 && status.UpstreamPower.Float > status.DownstreamPower.Float {
		status.DownstreamPower, status.UpstreamPower = status.UpstreamPower, status.DownstreamPower
	}

	status.DownstreamFECCount = interpretStatusIntValue(receiveStatistics, "dsfecerrors")
	status.UpstreamFECCount = interpretStatusIntValue(transmitStatistics, "usfecerrors")

	status.DownstreamCRCCount = interpretStatusIntValue(receiveStatistics, "dscrcerrors")
	status.UpstreamCRCCount = interpretStatusIntValue(transmitStatistics, "uscrcerrors")

	return status
}
