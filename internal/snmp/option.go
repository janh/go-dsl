// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package snmp

import (
	"3e8.eu/go/dsl"
)

func GetAuthProtocolOption(protos ...AuthProtocol) dsl.Option {
	opt := dsl.Option{
		Description: "SNMPv3 USM authentication protocol",
		Type:        dsl.OptionTypeEnum,
		Values:      []dsl.OptionValue{},
	}

	for _, proto := range protos {
		title, err := proto.desc()
		if err != nil {
			panic(err)
		}

		val := dsl.OptionValue{Value: string(proto), Title: title}
		opt.Values = append(opt.Values, val)
	}

	return opt
}

func GetPrivacyProtocolOption(protos ...PrivacyProtocol) dsl.Option {
	opt := dsl.Option{
		Description: "SNMPv3 USM privacy protocol",
		Type:        dsl.OptionTypeEnum,
		Values:      []dsl.OptionValue{},
	}

	for _, proto := range protos {
		title, err := proto.desc()
		if err != nil {
			panic(err)
		}

		val := dsl.OptionValue{Value: string(proto), Title: title}
		opt.Values = append(opt.Values, val)
	}

	return opt
}
