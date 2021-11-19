// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package sagemcom

import (
	"3e8.eu/go/dsl/models"
)

type dslWrapper struct {
	DSL dslObj `json:"DSL"`
}

type dslObj struct {
	Lines    []line    `json:"Lines"`
	Channels []channel `json:"Channels"`
}

type line struct {
	FirmwareVersion       string          `json:"FirmwareVersion"`
	LinkStatus            string          `json:"LinkStatus"`
	StandardUsed          string          `json:"StandardUsed"`
	VectoringState        string          `json:"VectoringState"`
	UpstreamMaxBitRate    models.IntValue `json:"UpstreamMaxBitRate"`
	DownstreamMaxBitRate  models.IntValue `json:"DownstreamMaxBitRate"`
	UpstreamNoiseMargin   models.IntValue `json:"UpstreamNoiseMargin"`
	DownstreamNoiseMargin models.IntValue `json:"DownstreamNoiseMargin"`
	UpstreamAttenuation   models.IntValue `json:"UpstreamAttenuation"`
	DownstreamAttenuation models.IntValue `json:"DownstreamAttenuation"`
	UpstreamPower         models.IntValue `json:"UpstreamPower"`
	DownstreamPower       models.IntValue `json:"DownstreamPower"`
	XTURVendor            string          `json:"XTURVendor"`
	XTUCVendor            string          `json:"XTUCVendor"`
	Stats                 lineStats       `json:"Stats"`
	TestParams            lineTestParams  `json:"TestParams"`
	IDDSLAM               string          `json:"IDDSLAM"`
	ModemChip             string          `json:"ModemChip"`
}

type lineStats struct {
	ShowtimeStart models.IntValue   `json:"ShowtimeStart"`
	Showtime      lineStatsCounters `json:"Showtime"`
}

type lineStatsCounters struct {
	ErroredSecs           models.IntValue `json:"ErroredSecs"`
	TxErroredSecs         models.IntValue `json:"TxErroredSecs"`
	SeverelyErroredSecs   models.IntValue `json:"SeverelyErroredSecs"`
	TxSeverelyErroredSecs models.IntValue `json:"TxSeverelyErroredSecs"`
}

type lineTestParams struct {
	HLOGGds  int    `json:"HLOGGds"`
	HLOGGus  int    `json:"HLOGGus"`
	HLOGpsds string `json:"HLOGpsds"`
	HLOGpsus string `json:"HLOGpsus"`
	QLNGds   int    `json:"QLNGds"`
	QLNGus   int    `json:"QLNGus"`
	QLNpsds  string `json:"QLNpsds"`
	QLNpsus  string `json:"QLNpsus"`
	SNRGds   int    `json:"SNRGds"`
	SNRGus   int    `json:"SNRGus"`
	SNRpsds  string `json:"SNRpsds"`
	SNRpsus  string `json:"SNRpsus"`
}

type channel struct {
	ActualInterleavingDelay   models.IntValue `json:"ActualInterleavingDelay"`
	ActualInterleavingDelayus models.IntValue `json:"ActualInterleavingDelayus"`
	ACTINP                    models.IntValue `json:"ACTINP"`
	ACTINPus                  models.IntValue `json:"ACTINPus"`
	UpstreamCurrRate          models.IntValue `json:"UpstreamCurrRate"`
	DownstreamCurrRate        models.IntValue `json:"DownstreamCurrRate"`
	Stats                     channelStats    `json:"Stats"`
}

type channelStats struct {
	Showtime channelStatsCounters `json:"Showtime"`
}

type channelStatsCounters struct {
	XTURFECErrors models.IntValue `json:"XTURFECErrors"`
	XTUCFECErrors models.IntValue `json:"XTUCFECErrors"`
	XTURCRCErrors models.IntValue `json:"XTURCRCErrors"`
	XTUCCRCErrors models.IntValue `json:"XTUCCRCErrors"`
}
