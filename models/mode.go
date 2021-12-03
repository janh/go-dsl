// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package models

import (
	"strings"
	"unicode"
)

type ModeType int

const (
	ModeTypeUnknown ModeType = iota

	ModeTypeADSL
	ModeTypeADSL2
	ModeTypeADSL2Plus
	ModeTypeVDSL2
)

func (t ModeType) String() string {
	switch t {
	case ModeTypeADSL:
		return "ADSL"
	case ModeTypeADSL2:
		return "ADSL2"
	case ModeTypeADSL2Plus:
		return "ADSL2+"
	case ModeTypeVDSL2:
		return "VDSL2"
	}
	return "Unknown"
}

type ModeSubtype int

const (
	ModeSubtypeUnknown ModeSubtype = iota

	ModeSubtypeAnnexA
	ModeSubtypeAnnexB
	ModeSubtypeAnnexI
	ModeSubtypeAnnexJ
	ModeSubtypeAnnexL
	ModeSubtypeAnnexM

	ModeSubtypeProfile8a
	ModeSubtypeProfile8b
	ModeSubtypeProfile8c
	ModeSubtypeProfile8d

	ModeSubtypeProfile12a
	ModeSubtypeProfile12b

	ModeSubtypeProfile17a

	ModeSubtypeProfile30a

	ModeSubtypeProfile35b
)

func (s ModeSubtype) String() string {
	switch s {

	case ModeSubtypeAnnexA:
		return "Annex A"
	case ModeSubtypeAnnexB:
		return "Annex B"
	case ModeSubtypeAnnexI:
		return "Annex I"
	case ModeSubtypeAnnexJ:
		return "Annex J"
	case ModeSubtypeAnnexL:
		return "Annex L"
	case ModeSubtypeAnnexM:
		return "Annex M"

	case ModeSubtypeProfile8a:
		return "Profile 8a"
	case ModeSubtypeProfile8b:
		return "Profile 8b"
	case ModeSubtypeProfile8c:
		return "Profile 8c"
	case ModeSubtypeProfile8d:
		return "Profile 8d"

	case ModeSubtypeProfile12a:
		return "Profile 12a"
	case ModeSubtypeProfile12b:
		return "Profile 12b"

	case ModeSubtypeProfile17a:
		return "Profile 17a"

	case ModeSubtypeProfile30a:
		return "Profile 30a"

	case ModeSubtypeProfile35b:
		return "Profile 35b"

	}

	return "Unknown"
}

type Mode struct {
	Type    ModeType
	Subtype ModeSubtype
}

func (m Mode) String() string {
	if m.Subtype != ModeSubtypeUnknown {
		return m.Type.String() + " " + m.Subtype.String()
	}
	return m.Type.String()
}

func (m Mode) BinCount() int {
	switch m.Type {

	case ModeTypeADSL, ModeTypeADSL2:
		return 256

	case ModeTypeADSL2Plus:
		return 512

	case ModeTypeVDSL2:

		switch m.Subtype {

		case ModeSubtypeProfile8a, ModeSubtypeProfile8b, ModeSubtypeProfile8c, ModeSubtypeProfile8d:
			return 1972

		case ModeSubtypeProfile12a, ModeSubtypeProfile12b:
			return 2783

		case ModeSubtypeProfile17a:
			return 4096

		case ModeSubtypeProfile30a:
			return 3479

		case ModeSubtypeProfile35b:
			return 8192

		}

	}

	return 8192
}

func (m Mode) CarrierSpacing() float64 {
	if m.Type == ModeTypeVDSL2 && m.Subtype == ModeSubtypeProfile30a {
		return 8.625
	}
	return 4.3125
}

func ParseMode(str string) Mode {
	str = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || r == '_' {
			return -1
		}
		return unicode.ToLower(r)
	}, str)

	var mode Mode

	switch {

	case strings.Contains(str, "adsl"), strings.Contains(str, "g.dmt"), strings.Contains(str, "g.992"):
		if strings.Contains(str, "adsl2+") || strings.Contains(str, "adsl2p") || strings.Contains(str, "g.992.5") {
			mode.Type = ModeTypeADSL2Plus
		} else if strings.Contains(str, "adsl2") || strings.Contains(str, "g.992.3") {
			mode.Type = ModeTypeADSL2
		} else {
			mode.Type = ModeTypeADSL
		}

		switch {

		case strings.Contains(str, "annexa"):
			mode.Subtype = ModeSubtypeAnnexA

		case strings.Contains(str, "annexb"):
			mode.Subtype = ModeSubtypeAnnexB

		case strings.Contains(str, "annexi"):
			mode.Subtype = ModeSubtypeAnnexI

		case strings.Contains(str, "annexj"):
			mode.Subtype = ModeSubtypeAnnexJ

		case strings.Contains(str, "annexl"):
			mode.Subtype = ModeSubtypeAnnexL

		case strings.Contains(str, "annexm"):
			mode.Subtype = ModeSubtypeAnnexM

		}

	case strings.Contains(str, "8a"):
		mode.Type = ModeTypeVDSL2
		mode.Subtype = ModeSubtypeProfile8a

	case strings.Contains(str, "8b"):
		mode.Type = ModeTypeVDSL2
		mode.Subtype = ModeSubtypeProfile8b

	case strings.Contains(str, "8c"):
		mode.Type = ModeTypeVDSL2
		mode.Subtype = ModeSubtypeProfile8c

	case strings.Contains(str, "8d"):
		mode.Type = ModeTypeVDSL2
		mode.Subtype = ModeSubtypeProfile8d

	case strings.Contains(str, "12a"):
		mode.Type = ModeTypeVDSL2
		mode.Subtype = ModeSubtypeProfile12a

	case strings.Contains(str, "12b"):
		mode.Type = ModeTypeVDSL2
		mode.Subtype = ModeSubtypeProfile12b

	case strings.Contains(str, "17a"):
		mode.Type = ModeTypeVDSL2
		mode.Subtype = ModeSubtypeProfile17a

	case strings.Contains(str, "30a"):
		mode.Type = ModeTypeVDSL2
		mode.Subtype = ModeSubtypeProfile30a

	case strings.Contains(str, "35b"):
		mode.Type = ModeTypeVDSL2
		mode.Subtype = ModeSubtypeProfile35b

	case strings.Contains(str, "vdsl2"), strings.Contains(str, "g.993.2"), strings.Contains(str, "g.993.5"):
		mode.Type = ModeTypeVDSL2

	}

	return mode
}
