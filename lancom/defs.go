// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lancom

const (
	lcsStatusAdsl      = ".1.3.6.1.4.1.2356.11.1.41"
	lcsStatusVdsl      = ".1.3.6.1.4.1.2356.11.1.75"
	lcsStatusXdslAdsl  = ".1.3.6.1.4.1.2356.11.1.99.1"
	lcsStatusXdslVdsl1 = ".1.3.6.1.4.1.2356.11.1.99.2"
	lcsStatusXdslVdsl2 = ".1.3.6.1.4.1.2356.11.1.99.3"

	oidLineState = ".1.0" // Integer -> lineState

	oidLineType    = ".2.0"  // OctetString -> lineType
	oidStandard    = ".3.0"  // Integer -> dslStandard
	oidVdslProfile = ".40.0" // Integer -> profile

	oidModemChipsetType                     = ".500.13.0" // OctetString -> string
	oidModemDspFirmwareVersion              = ".500.11.0" // OctetString -> string
	oidAdvancedDslamChipsetManufacturerDump = ".25.47.0"  // OctetString -> bytes

	oidConnectionDuration = ".54.0" // Integer

	oidDataRateDownstreamKbps = ".5.0" // Integer
	oidDataRateUpstreamKbps   = ".4.0" // Integer

	oidAttainableDataRateDownstreamKbps = ".36.0" // Integer
	oidAttainableDataRateUpstreamKbps   = ".37.0" // Integer

	oidSnrDownstreamDb = ".6.0" // OctetString -> float
	oidSnrUpstreamDb   = ".7.0" // OctetString -> float

	oidAttenuationDownstreamDb = ".8.0" // OctetString -> float
	oidAttenuationUpstreamDb   = ".9.0" // OctetString -> float

	oidInterleaveDownstreamMs = ".11.0" // OctetString -> float
	oidInterleaveUpstreamMs   = ".10.0" // OctetString -> float

	oidAdvancedDsInpSymbols = ".25.101.0" // OctetString -> float
	oidAdvancedUsInpSymbols = ".25.121.0" // OctetString -> float

	oidAdvancedDsCrcErrors = ".25.102.0" // Integer
	oidAdvancedUsCrcErrors = ".25.122.0" // Integer

	oidAdvancedDsFecErrors = ".25.103.0" // Integer
	oidAdvancedUsFecErrors = ".25.123.0" // Integer

	oidVectoring = ".46.0" // Integer -> vectoring

	oidAdvancedDsLineOptions = ".25.158.0" // OctetString -> lineOption
	oidAdvancedUsLineOptions = ".25.148.0" // OctetString -> lineOption

	oidAdvancedDsSraMode = ".25.172.0" // Integer -> sraMode
	oidAdvancedUsSraMode = ".25.162.0" // Integer -> sraMode

	oidAdvancedDsBitLoadingTable = ".25.13" // table
	oidAdvancedUsBitLoadingTable = ".25.25" // table
	// columns: base number (OctetString -> string), entries 0-9 (Integer), graph (OctetString -> string)

	oidAdvancedDsSnrPerSubCarrierTable = ".25.157" // table
	// columns: base number (OctetString -> string), entries 0-9 (Integer)
)

const (
	lineStateUnknown   = 0
	lineStateDown      = 1
	lineStateIdle      = 2
	lineStateHandshake = 3
	lineStateTraining  = 4
	lineStateShowtime  = 5
	lineStateShutdown  = 6
	lineStateError     = 7
)

const (
	lineTypePOTS = "over-POTS"
	lineTypeISDN = "over-ISDN"
)

const (
	dslStandardOff             = 0
	dslStandardT1dot413        = 2
	dslStandardGdotLite        = 3
	dslStandardGdotDMT         = 4
	dslStandardAdsl2           = 21
	dslStandardAdsl2plus       = 22
	dslStandardAdsl2AnnexM     = 31
	dslStandardAdsl2plusAnnexM = 32
	dslStandardAdsl2AnnexI     = 33
	dslStandardAdsl2plusAnnexI = 34
	dslStandardAdsl2AnnexJ     = 35
	dslStandardAdsl2plusAnnexJ = 36
	dslStandardAdsl2AnnexL     = 37
	dslStandardVdsl2           = 40
	dslStandardGFast           = 42
	dslStandardAdsl2Lite       = 54
	dslStandardGVector         = 254
	dslStandardUnknown         = 255
)

const (
	profileUnknown = 0
	profile8a      = 1
	profile8b      = 2
	profile8c      = 3
	profile8d      = 4
	profile12a     = 5
	profile12b     = 6
	profile17a     = 7
	profile30a     = 8
	profile35b     = 9
	profile106a    = 20
	profile106b    = 21
	profile106c    = 22
	profile212a    = 23
	profile212c    = 25
)

const (
	vectoringNo       = 1
	vectoringYes      = 2
	vectoringFriendly = 3
)

const (
	lineOptionRoc            = 1 << 4
	lineOptionBitswap        = 1 << 3
	lineOptionTrellis        = 1 << 2
	lineOptionVirtualNoise   = 1 << 1
	lineOptionRetransmission = 1 << 0
)

// These values are not documented, but seem to match the ones from Lantiq drivers
// (sraModeAtInit is confirmed to be correct from command line interface)
const (
	sraModeManual     = 1
	sraModeAtInit     = 2
	sraModeDynamic    = 3
	sraModeDynamicSos = 4
)
