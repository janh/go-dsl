// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lantiq

import (
	"fmt"
	"reflect"
	"strings"
)

type data struct {
	LineState                    string `command:"lsg" commandLegacy:"lsg 0"`
	G997_XTUSystemEnablingStatus string `command:"g997xtusesg" commandLegacy:"g997atusecg"`
	BandPlanSTatus               string `command:"bpstg" commandLegacy:"bpcg 0"`
	VersionInformation           string `command:"vig" commandLegacy:"vig"`
	G997_LineInventory_Far       string `command:"g997lig 1" commandLegacy:"g997lig 0 1"`

	G997_ChannelStatus_US string `command:"g997csg 0 0" commandLegacy:"g997csg 0 0 0"`
	G997_ChannelStatus_DS string `command:"g997csg 0 1" commandLegacy:"g997csg 0 0 1"`
	G997_LineStatus_US    string `command:"g997lsg 0 1" commandLegacy:"g997lsg 0 0"`
	G997_LineStatus_DS    string `command:"g997lsg 1 1" commandLegacy:"g997lsg 0 1"`
	LineFeatureStatus_US  string `command:"lfsg 0"`
	LineFeatureStatus_DS  string `command:"lfsg 1"`

	PM_ChannelCountersShowtime_Near string `command:"pmccsg 0 0 0,pmcctg 0 0" commandLegacy:"pmcctg 0 0 0"`
	PM_ChannelCountersShowtime_Far  string `command:"pmccsg 0 1 0,pmcctg 0 1" commandLegacy:"pmcctg 0 0 1"`
	PM_LineSecCountersShowtime_Near string `command:"pmlscsg 0 0,pmlscsg 0" commandLegacy:"pmlsctg 0 0"`
	PM_LineSecCountersShowtime_Far  string `command:"pmlscsg 1 0,pmlscsg 1" commandLegacy:"pmlsctg 0 1"`
	ReTxStatistics_Near             string `command:"rtsg 0"`
	ReTxStatistics_Far              string `command:"rtsg 1"`

	BandBorderStatus_US           string `command:"bbsg 0"`
	BandBorderStatus_DS           string `command:"bbsg 1"`
	G997_BitAllocationNscShort_US string `command:"g997bansg 0" commandLegacy:"g997banscsg 0 0"`
	G997_BitAllocationNscShort_DS string `command:"g997bansg 1" commandLegacy:"g997banscsg 0 1"`
	G997_SnrAllocationNscShort_US string `command:"g997sansg 0" commandLegacy:"g997snrnscsg 0 0"`
	G997_SnrAllocationNscShort_DS string `command:"g997sansg 1" commandLegacy:"g997snrnscsg 0 1"`
	G997_DeltSNR_US               string `command:"g997dsnrg 0 1" commandLegacy:"g997dsnrg 0 0"`
	G997_DeltSNR_DS               string `command:"g997dsnrg 1 1" commandLegacy:"g997dsnrg 0 1"`
	G997_DeltQLN_US               string `command:"g997dqlng 0 1" commandLegacy:"g997dqlng 0 0"`
	G997_DeltQLN_DS               string `command:"g997dqlng 1 1" commandLegacy:"g997dqlng 0 1"`
	G997_DeltHLOG_US              string `command:"g997dhlogg 0 1" commandLegacy:"g997dhlogg 0 0"`
	G997_DeltHLOG_DS              string `command:"g997dhlogg 1 1" commandLegacy:"g997dhlogg 0 1"`
}

type dataItemDesc struct {
	Data           *string
	Commands       []string
	CommandsLegacy []string
}

func (d *data) LoadData(e executor, command string) (err error) {
	var tagName = "command"

	vig, err := e.Execute(command + " vig")
	if err != nil {
		return err
	}
	if strings.Contains(vig, "DSL_APILibraryVersion=2") {
		tagName = "commandLegacy"
	}

	t := reflect.TypeOf(*d)
	v := reflect.ValueOf(d)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if commands, ok := field.Tag.Lookup(tagName); ok {
			var out string
			var err error

			commandsSplit := strings.Split(commands, ",")
			for _, cmd := range commandsSplit {
				if cmd != "vig" {
					out, err = e.Execute(command + " " + cmd)
					if err != nil {
						return err
					}
				} else {
					out = vig
				}

				truncate := 100
				if len(out) < truncate {
					truncate = len(out)
				}
				if !strings.Contains(out[:truncate], "command not found") {
					break
				}
			}

			val := v.Elem().FieldByName(field.Name)
			val.SetString(out)
		}
	}

	return nil
}

func (d *data) RawData() []byte {
	var b strings.Builder

	t := reflect.TypeOf(*d)
	v := reflect.ValueOf(d)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if cmd, ok := field.Tag.Lookup("command"); ok {
			if separator := strings.IndexRune(cmd, ','); separator != -1 {
				cmd = cmd[:separator]
			}

			fmt.Fprintf(&b, "# dsl_pipe %s # %s\n", cmd, field.Name)

			val := v.Elem().FieldByName(field.Name)
			fmt.Fprintln(&b, val.String())
		}
	}

	fmt.Fprintln(&b)
	return []byte(b.String())
}
