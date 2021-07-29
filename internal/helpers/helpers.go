// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package helpers

import (
	"encoding/hex"
	"strings"
)

var vendorMapping = map[string]string{
	"ALCB": "Alcatel",
	"ANDV": "Analog Devices",
	"BDCM": "Broadcom",
	"CENT": "Centillium",
	"CNXT": "Conexant",
	"DRAY": "DrayTek",
	"GSPN": "Globespan",
	"IFTN": "Infineon",
	"IKNS": "Ikanos",
	"RETK": "Realtek",
	"META": "Metanoia",
	"MTIA": "Metanoia",
	"STMI": "STMicro",
	"TCTN": "TrendChip",
	"TMMB": "Thomson",
	"TSTC": "Texas Instruments",
}

func FormatVendor(vendor string) string {
	if formattedVendor, ok := vendorMapping[vendor]; ok {
		return formattedVendor
	}

	end := strings.IndexByte(vendor, 0)
	if end != -1 {
		vendor = vendor[:end]
	}

	return vendor
}

func ParseHexadecimal(str string) []byte {
	if strings.HasPrefix(str, "0x") {
		str = str[2:]
	}
	bytes, _ := hex.DecodeString(str)
	return bytes
}
