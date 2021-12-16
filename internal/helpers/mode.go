// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package helpers

import (
	"strings"
	"unicode"

	"3e8.eu/go/dsl/models"
)

func ParseMode(str string) models.Mode {
	str = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || r == '_' {
			return -1
		}
		return unicode.ToLower(r)
	}, str)

	var mode models.Mode

	switch {

	case strings.Contains(str, "adsl"), strings.Contains(str, "g.dmt"), strings.Contains(str, "g.992"):
		if strings.Contains(str, "adsl2+") || strings.Contains(str, "adsl2p") || strings.Contains(str, "g.992.5") {
			mode.Type = models.ModeTypeADSL2Plus
		} else if strings.Contains(str, "adsl2") || strings.Contains(str, "g.992.3") {
			mode.Type = models.ModeTypeADSL2
		} else {
			mode.Type = models.ModeTypeADSL
		}

		switch {

		case strings.Contains(str, "annexa"):
			mode.Subtype = models.ModeSubtypeAnnexA

		case strings.Contains(str, "annexb"):
			mode.Subtype = models.ModeSubtypeAnnexB

		case strings.Contains(str, "annexi"):
			mode.Subtype = models.ModeSubtypeAnnexI

		case strings.Contains(str, "annexj"):
			mode.Subtype = models.ModeSubtypeAnnexJ

		case strings.Contains(str, "annexl"):
			mode.Subtype = models.ModeSubtypeAnnexL

		case strings.Contains(str, "annexm"):
			mode.Subtype = models.ModeSubtypeAnnexM

		}

	case strings.Contains(str, "8a"):
		mode.Type = models.ModeTypeVDSL2
		mode.Subtype = models.ModeSubtypeProfile8a

	case strings.Contains(str, "8b"):
		mode.Type = models.ModeTypeVDSL2
		mode.Subtype = models.ModeSubtypeProfile8b

	case strings.Contains(str, "8c"):
		mode.Type = models.ModeTypeVDSL2
		mode.Subtype = models.ModeSubtypeProfile8c

	case strings.Contains(str, "8d"):
		mode.Type = models.ModeTypeVDSL2
		mode.Subtype = models.ModeSubtypeProfile8d

	case strings.Contains(str, "12a"):
		mode.Type = models.ModeTypeVDSL2
		mode.Subtype = models.ModeSubtypeProfile12a

	case strings.Contains(str, "12b"):
		mode.Type = models.ModeTypeVDSL2
		mode.Subtype = models.ModeSubtypeProfile12b

	case strings.Contains(str, "17a"):
		mode.Type = models.ModeTypeVDSL2
		mode.Subtype = models.ModeSubtypeProfile17a

	case strings.Contains(str, "30a"):
		mode.Type = models.ModeTypeVDSL2
		mode.Subtype = models.ModeSubtypeProfile30a

	case strings.Contains(str, "35b"):
		mode.Type = models.ModeTypeVDSL2
		mode.Subtype = models.ModeSubtypeProfile35b

	case strings.Contains(str, "vdsl2"), strings.Contains(str, "g.993.2"), strings.Contains(str, "g.993.5"):
		mode.Type = models.ModeTypeVDSL2

	}

	return mode
}
