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

func parseStatus(values snmp.Values, oidBase string) models.Status {
	var status models.Status

	status.State = interpretState(values, oidBase+oidLineState)
	status.Mode = interpretMode(values, oidBase+oidStandard, oidBase+oidLineType, oidBase+oidVdslProfile)

	status.NearEndInventory = interpretNearEndInventory(values,
		oidBase+oidModemChipsetType, oidBase+oidModemDspFirmwareVersion)
	status.FarEndInventory = interpretFarEndInventory(values, oidBase+oidAdvancedDslamChipsetManufacturerDump)

	status.Uptime = interpretUptime(values, oidBase+oidConnectionDuration)

	status.DownstreamActualRate.IntValue = interpretIntValue(values, oidBase+oidDataRateDownstreamKbps)
	status.UpstreamActualRate.IntValue = interpretIntValue(values, oidBase+oidDataRateUpstreamKbps)

	status.DownstreamAttainableRate.IntValue = interpretIntValue(values, oidBase+oidAttainableDataRateDownstreamKbps)
	status.UpstreamAttainableRate.IntValue = interpretIntValue(values, oidBase+oidAttainableDataRateUpstreamKbps)

	status.DownstreamSNRMargin.FloatValue = interpretFloatValue(values, oidBase+oidSnrDownstreamDb)
	status.UpstreamSNRMargin.FloatValue = interpretFloatValue(values, oidBase+oidSnrUpstreamDb)

	status.DownstreamAttenuation.FloatValue = interpretFloatValue(values, oidBase+oidAttenuationDownstreamDb)
	status.UpstreamAttenuation.FloatValue = interpretFloatValue(values, oidBase+oidAttenuationUpstreamDb)

	status.DownstreamInterleavingDelay.FloatValue = interpretFloatValue(values, oidBase+oidInterleaveDownstreamMs)
	status.UpstreamInterleavingDelay.FloatValue = interpretFloatValue(values, oidBase+oidInterleaveUpstreamMs)

	status.DownstreamImpulseNoiseProtection.FloatValue = interpretFloatValue(values, oidBase+oidAdvancedDsInpSymbols)
	status.UpstreamImpulseNoiseProtection.FloatValue = interpretFloatValue(values, oidBase+oidAdvancedUsInpSymbols)

	status.DownstreamCRCCount = interpretIntValue(values, oidBase+oidAdvancedDsCrcErrors)
	status.UpstreamCRCCount = interpretIntValue(values, oidBase+oidAdvancedUsCrcErrors)

	status.DownstreamFECCount = interpretIntValue(values, oidBase+oidAdvancedDsFecErrors)
	status.UpstreamFECCount = interpretIntValue(values, oidBase+oidAdvancedUsFecErrors)

	status.DownstreamVectoringState = interpretVectoringValue(values, oidBase+oidVectoring)

	status.DownstreamBitswap.Enabled, status.DownstreamRetransmissionEnabled =
		interpretLineOptions(values, oidBase+oidAdvancedDsLineOptions)

	status.UpstreamBitswap.Enabled, status.UpstreamRetransmissionEnabled =
		interpretLineOptions(values, oidBase+oidAdvancedUsLineOptions)

	status.DownstreamSeamlessRateAdaptation.Enabled = interpretSraMode(values, oidBase+oidAdvancedDsSraMode)
	status.UpstreamSeamlessRateAdaptation.Enabled = interpretSraMode(values, oidBase+oidAdvancedUsSraMode)

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

func interpretSraMode(values snmp.Values, oid string) (out models.BoolValue) {
	if val, err := values.GetUint64(oid); err == nil {
		out.Bool = (val == sraModeDynamic || val == sraModeDynamicSos)
		out.Valid = true
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
