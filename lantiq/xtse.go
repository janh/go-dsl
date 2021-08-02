// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lantiq

import (
	"3e8.eu/go/dsl/models"
)

const (
	// ADSL Annex A, non-overlapped spectrum
	XTSE_1_03_A_1_NO = 0x04
	// ADSL Annex A, overlapped spectrum
	XTSE_1_04_A_1_O = 0x08
	// ADSL Annex B, non-overlapped spectrum
	XTSE_1_05_B_1_NO = 0x10
	// ADSL Annex B, overlapped spectrum
	XTSE_1_06_B_1_O = 0x20

	// ADSL2 Annex A, non-overlapped spectrum
	XTSE_3_03_A_3_NO = 0x04
	// ADSL2 Annex A, overlapped spectrum
	XTSE_3_04_A_3_O = 0x08
	// ADSL2 Annex B, non-overlapped spectrum
	XTSE_3_05_B_3_NO = 0x10
	// ADSL2 Annex B, overlapped spectrum
	XTSE_3_06_B_3_O = 0x20

	// ADSL2 Annex I, non-overlapped spectrum
	XTSE_4_05_I_3_NO = 0x10
	// ADSL2 Annex I, overlapped spectrum
	XTSE_4_06_I_3_O = 0x20
	// ADSL2 Annex J, non-overlapped spectrum
	XTSE_4_07_J_3_NO = 0x40
	// ADSL2 Annex J, overlapped spectrum
	XTSE_4_08_J_3_O = 0x80

	// ADSL2 Annex L, non-overlapped, wide upstream
	XTSE_5_03_L_3_NO = 0x04
	// ADSL2 Annex L, non-overlapped, narrow upstream
	XTSE_5_04_L_3_NO = 0x08
	// ADSL2 Annex L, overlapped, wide upstream
	XTSE_5_05_L_3_O = 0x10
	// ADSL2 Annex L, overlapped, narrow upstream
	XTSE_5_06_L_3_O = 0x20
	// ADSL2 Annex M, non-overlapped spectrum
	XTSE_5_07_M_3_NO = 0x40
	// ADSL2 Annex M, overlapped spectrum
	XTSE_5_08_M_3_O = 0x80

	// ADSL2+ Annex A, non-overlapped spectrum
	XTSE_6_01_A_5_NO = 0x01
	// ADSL2+ Annex A, overlapped spectrum
	XTSE_6_02_A_5_O = 0x02
	// ADSL2+ Annex B, non-overlapped spectrum
	XTSE_6_03_B_5_NO = 0x04
	// ADSL2+ Annex B, overlapped spectrum
	XTSE_6_04_B_5_O = 0x08
	// ADSL2+ Annex I, non-overlapped spectrum
	XTSE_6_07_I_5_NO = 0x40
	// ADSL2+ Annex I, overlapped spectrum
	XTSE_6_08_I_5_O = 0x80

	// ADSL2+ Annex J, non-overlapped spectrum
	XTSE_7_01_J_5_NO = 0x01
	// ADSL2+ Annex J, overlapped spectrum
	XTSE_7_02_J_5_O = 0x02
	// ADSL2+ Annex M, non-overlapped spectrum
	XTSE_7_03_M_5_NO = 0x04
	// ADSL2+ Annex M, overlapped spectrum
	XTSE_7_04_M_5_O = 0x08

	// VDSL2 Annex A (North America)
	XTSE_8_01_A = 0x01
	// VDSL2 Annex B (Europe)
	XTSE_8_02_B = 0x02
	// VDSL2 Annex C (Japan)
	XTSE_8_03_C = 0x04
)

func getStatusModeType(xtse1, xtse2, xtse3, xtse4, xtse5, xtse6, xtse7, xtse8 byte) models.ModeType {
	switch {

	case xtse1&XTSE_1_03_A_1_NO != 0,
		xtse1&XTSE_1_04_A_1_O != 0,
		xtse1&XTSE_1_05_B_1_NO != 0,
		xtse1&XTSE_1_06_B_1_O != 0:

		return models.ModeTypeADSL

	case xtse3&XTSE_3_03_A_3_NO != 0,
		xtse3&XTSE_3_04_A_3_O != 0,
		xtse3&XTSE_3_05_B_3_NO != 0,
		xtse3&XTSE_3_06_B_3_O != 0,
		xtse4&XTSE_4_05_I_3_NO != 0,
		xtse4&XTSE_4_06_I_3_O != 0,
		xtse4&XTSE_4_07_J_3_NO != 0,
		xtse4&XTSE_4_08_J_3_O != 0,
		xtse5&XTSE_5_03_L_3_NO != 0,
		xtse5&XTSE_5_04_L_3_NO != 0,
		xtse5&XTSE_5_05_L_3_O != 0,
		xtse5&XTSE_5_06_L_3_O != 0,
		xtse5&XTSE_5_07_M_3_NO != 0,
		xtse5&XTSE_5_08_M_3_O != 0:

		return models.ModeTypeADSL2

	case xtse6&XTSE_6_01_A_5_NO != 0,
		xtse6&XTSE_6_02_A_5_O != 0,
		xtse6&XTSE_6_03_B_5_NO != 0,
		xtse6&XTSE_6_04_B_5_O != 0,
		xtse6&XTSE_6_07_I_5_NO != 0,
		xtse6&XTSE_6_08_I_5_O != 0,
		xtse7&XTSE_7_01_J_5_NO != 0,
		xtse7&XTSE_7_02_J_5_O != 0,
		xtse7&XTSE_7_03_M_5_NO != 0,
		xtse7&XTSE_7_04_M_5_O != 0:

		return models.ModeTypeADSL2Plus

	case xtse8&XTSE_8_01_A != 0,
		xtse8&XTSE_8_02_B != 0,
		xtse8&XTSE_8_03_C != 0:

		return models.ModeTypeVDSL2

	}

	return models.ModeTypeUnknown
}

func getStatusModeSubtype(xtse1, xtse2, xtse3, xtse4, xtse5, xtse6, xtse7, xtse8 byte) models.ModeSubtype {
	switch {

	case xtse1&XTSE_1_03_A_1_NO != 0,
		xtse1&XTSE_1_04_A_1_O != 0,
		xtse3&XTSE_3_03_A_3_NO != 0,
		xtse3&XTSE_3_04_A_3_O != 0,
		xtse6&XTSE_6_01_A_5_NO != 0,
		xtse6&XTSE_6_02_A_5_O != 0:

		return models.ModeSubtypeAnnexA

	case xtse1&XTSE_1_05_B_1_NO != 0,
		xtse1&XTSE_1_06_B_1_O != 0,
		xtse3&XTSE_3_05_B_3_NO != 0,
		xtse3&XTSE_3_06_B_3_O != 0,
		xtse6&XTSE_6_03_B_5_NO != 0,
		xtse6&XTSE_6_04_B_5_O != 0:

		return models.ModeSubtypeAnnexB

	case xtse4&XTSE_4_05_I_3_NO != 0,
		xtse4&XTSE_4_06_I_3_O != 0,
		xtse6&XTSE_6_07_I_5_NO != 0,
		xtse6&XTSE_6_08_I_5_O != 0:

		return models.ModeSubtypeAnnexI

	case xtse4&XTSE_4_07_J_3_NO != 0,
		xtse4&XTSE_4_08_J_3_O != 0,
		xtse7&XTSE_7_01_J_5_NO != 0,
		xtse7&XTSE_7_02_J_5_O != 0:

		return models.ModeSubtypeAnnexJ

	case xtse5&XTSE_5_03_L_3_NO != 0,
		xtse5&XTSE_5_04_L_3_NO != 0,
		xtse5&XTSE_5_05_L_3_O != 0,
		xtse5&XTSE_5_06_L_3_O != 0:

		return models.ModeSubtypeAnnexL

	case xtse5&XTSE_5_07_M_3_NO != 0,
		xtse5&XTSE_5_08_M_3_O != 0,
		xtse7&XTSE_7_03_M_5_NO != 0,
		xtse7&XTSE_7_04_M_5_O != 0:

		return models.ModeSubtypeAnnexM

	}

	return models.ModeSubtypeUnknown
}
