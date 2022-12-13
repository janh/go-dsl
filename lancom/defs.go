// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lancom

const (
	lcsStatusVdsl = ".1.3.6.1.4.1.2356.11.1.75"

	lcsStatusVdslLineState = ".1.3.6.1.4.1.2356.11.1.75.1.0" // Integer -> lineState

	lcsStatusVdslLineType    = ".1.3.6.1.4.1.2356.11.1.75.2.0"  // OctetString -> lineType
	lcsStatusVdslStandard    = ".1.3.6.1.4.1.2356.11.1.75.3.0"  // Integer -> dslStandard
	lcsStatusVdslVdslProfile = ".1.3.6.1.4.1.2356.11.1.75.40.0" // Integer -> profile

	lcsStatusVdslModemChipsetType                     = ".1.3.6.1.4.1.2356.11.1.75.500.13.0" // OctetString -> string
	lcsStatusVdslModemDspFirmwareVersion              = ".1.3.6.1.4.1.2356.11.1.75.500.11.0" // OctetString -> string
	lcsStatusVdslAdvancedDslamChipsetManufacturerDump = ".1.3.6.1.4.1.2356.11.1.75.25.47.0"  // OctetString -> bytes

	lcsStatusVdslConnectionDuration = ".1.3.6.1.4.1.2356.11.1.75.54.0" // Integer

	lcsStatusVdslDataRateDownstreamKbps = ".1.3.6.1.4.1.2356.11.1.75.5.0" // Integer
	lcsStatusVdslDataRateUpstreamKbps   = ".1.3.6.1.4.1.2356.11.1.75.4.0" // Integer

	lcsStatusVdslAttainableDataRateDownstreamKbps = ".1.3.6.1.4.1.2356.11.1.75.36.0" // Integer
	lcsStatusVdslAttainableDataRateUpstreamKbps   = ".1.3.6.1.4.1.2356.11.1.75.37.0" // Integer

	lcsStatusVdslSnrDownstreamDb = ".1.3.6.1.4.1.2356.11.1.75.6.0" // OctetString -> float
	lcsStatusVdslSnrUpstreamDb   = ".1.3.6.1.4.1.2356.11.1.75.7.0" // OctetString -> float

	lcsStatusVdslAttenuationDownstreamDb = ".1.3.6.1.4.1.2356.11.1.75.8.0" // OctetString -> float
	lcsStatusVdslAttenuationUpstreamDb   = ".1.3.6.1.4.1.2356.11.1.75.9.0" // OctetString -> float

	lcsStatusVdslInterleaveDownstreamMs = ".1.3.6.1.4.1.2356.11.1.75.11.0" // OctetString -> float
	lcsStatusVdslInterleaveUpstreamMs   = ".1.3.6.1.4.1.2356.11.1.75.10.0" // OctetString -> float

	lcsStatusVdslAdvancedDsInpSymbols = ".1.3.6.1.4.1.2356.11.1.75.25.101.0" // OctetString -> float
	lcsStatusVdslAdvancedUsInpSymbols = ".1.3.6.1.4.1.2356.11.1.75.25.121.0" // OctetString -> float

	lcsStatusVdslAdvancedDsCrcErrors = ".1.3.6.1.4.1.2356.11.1.75.25.102.0" // Integer
	lcsStatusVdslAdvancedUsCrcErrors = ".1.3.6.1.4.1.2356.11.1.75.25.122.0" // Integer

	lcsStatusVdslAdvancedDsFecErrors = ".1.3.6.1.4.1.2356.11.1.75.25.103.0" // Integer
	lcsStatusVdslAdvancedUsFecErrors = ".1.3.6.1.4.1.2356.11.1.75.25.123.0" // Integer

	lcsStatusVdslVectoring = ".1.3.6.1.4.1.2356.11.1.75.46.0" // Integer -> vectoring

	lcsStatusVdslAdvancedDsLineOptions = ".1.3.6.1.4.1.2356.11.1.75.25.158.0" // OctetString -> lineOption
	lcsStatusVdslAdvancedUsLineOptions = ".1.3.6.1.4.1.2356.11.1.75.25.148.0" // OctetString -> lineOption

	lcsStatusVdslAdvancedDsBitLoadingTable = ".1.3.6.1.4.1.2356.11.1.75.25.13" // table
	lcsStatusVdslAdvancedUsBitLoadingTable = ".1.3.6.1.4.1.2356.11.1.75.25.25" // table
	// columns: base number (OctetString -> string), entries 0-9 (Integer), graph (OctetString -> string)

	lcsStatusVdslAdvancedDsSnrPerSubCarrierTable = ".1.3.6.1.4.1.2356.11.1.75.25.157" // table
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
