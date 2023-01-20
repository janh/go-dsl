// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lancom

import (
	"3e8.eu/go/dsl"
	"3e8.eu/go/dsl/internal/snmp"
)

func init() {
	newFunc := func(config dsl.Config) (dsl.Client, error) {
		clientConfig := Config{
			Host:                 config.Host,
			User:                 config.User,
			Password:             config.AuthPassword,
			EncryptionPassphrase: config.EncryptionPassphrase,
			AuthProtocol:         config.Options["AuthProtocol"],
			PrivacyProtocol:      config.Options["PrivacyProtocol"],
			Subtree:              config.Options["Subtree"],
		}
		return NewClient(clientConfig)
	}
	clientDesc := dsl.ClientDesc{
		Title:                        "LANCOM (SNMPv3)",
		RequiresUser:                 dsl.TristateYes,
		SupportedAuthTypes:           dsl.AuthTypePassword,
		SupportsEncryptionPassphrase: true,
		Options: map[string]dsl.Option{
			"AuthProtocol": snmp.GetAuthProtocolOption(
				snmp.AuthProtocolNone,
				snmp.AuthProtocolMD5,
				snmp.AuthProtocolSHA,
				snmp.AuthProtocolSHA224,
				snmp.AuthProtocolSHA256,
				snmp.AuthProtocolSHA384,
				snmp.AuthProtocolSHA512,
			),
			"PrivacyProtocol": snmp.GetPrivacyProtocolOption(
				snmp.PrivacyProtocolNone,
				snmp.PrivacyProtocolDES,
				snmp.PrivacyProtocolAES128,
				snmp.PrivacyProtocolAES192,
				snmp.PrivacyProtocolAES192C,
				snmp.PrivacyProtocolAES256,
				snmp.PrivacyProtocolAES256C,
			),
			"Subtree": dsl.Option{
				Description: "the LCOS subtree to load data from",
				Type:        dsl.OptionTypeEnum,
				Values: []dsl.OptionValue{
					dsl.OptionValue{Value: "", Title: "auto-detect"},
					dsl.OptionValue{Value: "/Status/VDSL", Title: "Status > VDSL"},
					dsl.OptionValue{Value: "/Status/xDSL/VDSL1", Title: "Status > xDSL > VDSL1"},
					dsl.OptionValue{Value: "/Status/xDSL/VDSL2", Title: "Status > xDSL > VDSL2"},
					dsl.OptionValue{Value: "/Status/ADSL", Title: "Status > ADSL"},
					dsl.OptionValue{Value: "/Status/xDSL/ADSL", Title: "Status > xDSL > ADSL"},
				},
			},
		},
	}
	dsl.RegisterClient("lancom_snmpv3", newFunc, clientDesc)
}
