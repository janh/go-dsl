// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lancom

import (
	"strings"
	"time"

	"3e8.eu/go/dsl/internal/helpers"
	"3e8.eu/go/dsl/internal/snmp"
	"3e8.eu/go/dsl/models"
)

func parseStatus(values snmp.Values) models.Status {
	var status models.Status

	status.State = interpretState(values, lcsStatusVdslLineState)
	status.Mode = interpretMode(values, lcsStatusVdslStandard, lcsStatusVdslLineType, lcsStatusVdslVdslProfile)

	status.NearEndInventory = interpretNearEndInventory(values,
		lcsStatusVdslModemChipsetType, lcsStatusVdslModemDspFirmwareVersion)
	status.FarEndInventory = interpretFarEndInventory(values, lcsStatusVdslAdvancedDslamChipsetManufacturerDump)

	status.Uptime = interpretUptime(values, lcsStatusVdslConnectionDuration)

	status.DownstreamActualRate.IntValue = interpretIntValue(values, lcsStatusVdslDataRateDownstreamKbps)
	status.UpstreamActualRate.IntValue = interpretIntValue(values, lcsStatusVdslDataRateUpstreamKbps)

	status.DownstreamAttainableRate.IntValue = interpretIntValue(values, lcsStatusVdslAttainableDataRateDownstreamKbps)
	status.UpstreamAttainableRate.IntValue = interpretIntValue(values, lcsStatusVdslAttainableDataRateUpstreamKbps)

	status.DownstreamSNRMargin.FloatValue = interpretFloatValue(values, lcsStatusVdslSnrDownstreamDb)
	status.UpstreamSNRMargin.FloatValue = interpretFloatValue(values, lcsStatusVdslSnrUpstreamDb)

	status.DownstreamAttenuation.FloatValue = interpretFloatValue(values, lcsStatusVdslAttenuationDownstreamDb)
	status.UpstreamAttenuation.FloatValue = interpretFloatValue(values, lcsStatusVdslAttenuationUpstreamDb)

	status.DownstreamInterleavingDelay.FloatValue = interpretFloatValue(values, lcsStatusVdslInterleaveDownstreamMs)
	status.UpstreamInterleavingDelay.FloatValue = interpretFloatValue(values, lcsStatusVdslInterleaveUpstreamMs)

	status.DownstreamCRCCount = interpretIntValue(values, lcsStatusVdslAdvancedDsCrcErrors)
	status.UpstreamCRCCount = interpretIntValue(values, lcsStatusVdslAdvancedUsCrcErrors)

	status.DownstreamImpulseNoiseProtection.FloatValue = interpretFloatValue(values, lcsStatusVdslAdvancedDsInpSymbols)
	status.UpstreamImpulseNoiseProtection.FloatValue = interpretFloatValue(values, lcsStatusVdslAdvancedUsInpSymbols)

	status.DownstreamFECCount = interpretIntValue(values, lcsStatusVdslAdvancedDsFecErrors)
	status.UpstreamFECCount = interpretIntValue(values, lcsStatusVdslAdvancedUsFecErrors)

	status.DownstreamVectoringState = interpretVectoringValue(values, lcsStatusVdslVectoring)

	status.DownstreamBitswapEnabled, status.DownstreamRetransmissionEnabled =
		interpretLineOptions(values, lcsStatusVdslAdvancedDsLineOptions)

	status.UpstreamBitswapEnabled, status.UpstreamRetransmissionEnabled =
		interpretLineOptions(values, lcsStatusVdslAdvancedUsLineOptions)

	return status
}

func interpretState(values snmp.Values, oid string) (out models.State) {
	if lineState, err := values.GetUint64(oid); err == nil {
		switch lineState {
		case lineStateDown, lineStateShutdown, lineStateIdle:
			out = models.StateIdle
		case lineStateHandshake:
			out = models.StateHandshake
		case lineStateTraining:
			out = models.StateTraining
		case lineStateShowtime:
			out = models.StateShowtime
		}
	}
	return
}

func interpretMode(values snmp.Values, oidStandard, oidLineType, oidProfile string) (out models.Mode) {
	if standard, err := values.GetUint64(oidStandard); err == nil {
		switch standard {
		case dslStandardGdotDMT:
			out.Type = models.ModeTypeADSL
		case dslStandardAdsl2:
			out.Type = models.ModeTypeADSL2
		case dslStandardAdsl2plus:
			out.Type = models.ModeTypeADSL2Plus
		case dslStandardAdsl2AnnexM:
			out.Type = models.ModeTypeADSL2
			out.Subtype = models.ModeSubtypeAnnexM
		case dslStandardAdsl2plusAnnexM:
			out.Type = models.ModeTypeADSL2Plus
			out.Subtype = models.ModeSubtypeAnnexM
		case dslStandardAdsl2AnnexI:
			out.Type = models.ModeTypeADSL2
			out.Subtype = models.ModeSubtypeAnnexI
		case dslStandardAdsl2plusAnnexI:
			out.Type = models.ModeTypeADSL2Plus
			out.Subtype = models.ModeSubtypeAnnexI
		case dslStandardAdsl2AnnexJ:
			out.Type = models.ModeTypeADSL2
			out.Subtype = models.ModeSubtypeAnnexJ
		case dslStandardAdsl2plusAnnexJ:
			out.Type = models.ModeTypeADSL2Plus
			out.Subtype = models.ModeSubtypeAnnexJ
		case dslStandardAdsl2AnnexL:
			out.Type = models.ModeTypeADSL2
			out.Subtype = models.ModeSubtypeAnnexL
		case dslStandardVdsl2, dslStandardGVector:
			out.Type = models.ModeTypeVDSL2
		}
	}

	if out.Type == models.ModeTypeVDSL2 {
		if profile, err := values.GetUint64(oidProfile); err == nil {
			switch profile {
			case profile8a:
				out.Subtype = models.ModeSubtypeProfile8a
			case profile8b:
				out.Subtype = models.ModeSubtypeProfile8b
			case profile8c:
				out.Subtype = models.ModeSubtypeProfile8c
			case profile8d:
				out.Subtype = models.ModeSubtypeProfile8d
			case profile12a:
				out.Subtype = models.ModeSubtypeProfile12a
			case profile12b:
				out.Subtype = models.ModeSubtypeProfile12b
			case profile17a:
				out.Subtype = models.ModeSubtypeProfile17a
			case profile30a:
				out.Subtype = models.ModeSubtypeProfile30a
			case profile35b:
				out.Subtype = models.ModeSubtypeProfile35b
			}
		}
	}

	if (out.Type == models.ModeTypeADSL || out.Type == models.ModeTypeADSL2 || out.Type == models.ModeTypeADSL2Plus) &&
		out.Subtype == models.ModeSubtypeUnknown {

		if lineType, err := values.GetString(oidLineType); err == nil {
			switch lineType {
			case lineTypePOTS:
				out.Subtype = models.ModeSubtypeAnnexA
			case lineTypeISDN:
				out.Subtype = models.ModeSubtypeAnnexB
			}
		}
	}

	return
}

func interpretNearEndInventory(values snmp.Values, oidChipset, oidVersion string) (out models.Inventory) {
	out.Vendor = "LANCOM"
	if val, err := values.GetString(oidChipset); err == nil {
		if strings.HasPrefix(val, "Lantiq") || strings.HasPrefix(val, "Ifx") || strings.Contains(val, "VINAX") {
			out.Vendor = "Infineon"
		}
	}
	if val, err := values.GetString(oidVersion); err == nil {
		out.Version = val
	}
	return
}

func interpretFarEndInventory(values snmp.Values, oid string) (out models.Inventory) {
	if val, err := values.GetBytes(oid); err == nil {
		if len(val) == 8 {
			out.Vendor = helpers.FormatVendor(string(val[2:6]))
			out.Version = helpers.FormatVersion(out.Vendor, val[6:8])
		}
	}
	return
}

func interpretUptime(values snmp.Values, oid string) (out models.Duration) {
	if val, err := values.GetInt64(oid); err == nil {
		out.Duration = time.Duration(val) * time.Second
	}
	return
}

func interpretVectoringValue(values snmp.Values, oid string) (out models.VectoringValue) {
	if val, err := values.GetUint64(oid); err == nil {
		switch val {
		case vectoringNo:
			out.State = models.VectoringStateOff
			out.Valid = true
		case vectoringYes:
			out.State = models.VectoringStateFull
			out.Valid = true
		case vectoringFriendly:
			out.State = models.VectoringStateFriendly
			out.Valid = true
		}
	}
	return
}

func interpretLineOptions(values snmp.Values, oid string) (outBitswap, outRetransmission models.BoolValue) {
	if val, err := values.GetUint64(oid); err == nil {
		outBitswap.Bool = (val & lineOptionBitswap) != 0
		outBitswap.Valid = true
		outRetransmission.Bool = (val & lineOptionRetransmission) != 0
		outRetransmission.Valid = true
	}
	return
}

func interpretIntValue(values snmp.Values, oid string) (out models.IntValue) {
	if val, err := values.GetInt64(oid); err == nil {
		out.Int = val
		out.Valid = true
	}
	return
}

func interpretFloatValue(values snmp.Values, oid string) (out models.FloatValue) {
	if val, err := values.GetFloat64(oid); err == nil {
		out.Float = val
		out.Valid = true
	}
	return
}
