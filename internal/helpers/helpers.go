// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package helpers

import (
	"encoding/hex"
	"fmt"
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
	"TCCN": "TrendChip",
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

func FormatVersion(vendor string, version []byte) string {
	if len(version) != 2 {
		return ""
	}

	if version[0] == 0 && version[1] == 0 {
		return ""
	}

	if vendor == "Infineon" {
		if version[0]&0xf0 == 0x90 {
			return fmt.Sprintf("%d.%d.%d.%d (%d.%d)",
				version[0]>>4, version[0]&0xf<<1+version[1]>>7, version[1]>>4&0x7, version[1]&0xf,
				version[0], version[1])
		} else {
			return fmt.Sprintf("%d.%d.%d.%d (%d.%d)",
				version[0]>>4, version[0]&0xf, version[1]>>4, version[1]&0xf,
				version[0], version[1])
		}
	}

	if vendor == "Broadcom" {
		return fmt.Sprintf("%d.%d.%d (%d.%d)",
			version[0]>>4, version[0]&0xf<<1+version[1]>>7, version[1]&0x7f,
			version[0], version[1])
	}

	return fmt.Sprintf("%d.%d", version[0], version[1])
}

func ParseHexadecimal(str string) []byte {
	if strings.HasPrefix(str, "0x") {
		str = str[2:]
	}
	bytes, _ := hex.DecodeString(str)
	return bytes
}
