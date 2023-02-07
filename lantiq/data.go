// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lantiq

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"3e8.eu/go/dsl/internal/exec"
)

type dataItem struct {
	Command string
	Output  string
}

type data struct {
	APIVersion string

	LineState                    dataItem `command:"lsg" commandLegacy:"lsg 0"`
	G997_XTUSystemEnablingStatus dataItem `command:"g997xtusesg" commandLegacy:"g997atusesg 0"`
	BandPlanSTatus               dataItem `command:"bpstg" commandLegacy:"bpcg 0"`
	VersionInformation           dataItem `command:"vig" commandLegacy:"vig"`
	G997_LineInventory_Far       dataItem `command:"g997lig 1" commandLegacy:"g997lig 0 1"`

	G997_ChannelStatus_US        dataItem `command:"g997csg 0 0" commandLegacy:"g997csg 0 0 0"`
	G997_ChannelStatus_DS        dataItem `command:"g997csg 0 1" commandLegacy:"g997csg 0 0 1"`
	G997_LineStatus_US           dataItem `command:"g997lsg 0 1" commandLegacy:"g997lsg 0 0"`
	G997_LineStatus_DS           dataItem `command:"g997lsg 1 1" commandLegacy:"g997lsg 0 1"`
	LineFeatureStatus_US         dataItem `command:"lfsg 0"`
	LineFeatureStatus_DS         dataItem `command:"lfsg 1"`
	G997_RateAdaptationStatus_US dataItem `command:"g997rasg 0"`
	G997_RateAdaptationStatus_DS dataItem `command:"g997rasg 1"`
	OlrStatistics_US             dataItem `command:"osg 0" commandLegacy:"ostg 0 0"`
	OlrStatistics_DS             dataItem `command:"osg 1" commandLegacy:"ostg 0 1"`
	DSM_Status                   dataItem `command:"dsmsg"`

	PM_ChannelCountersShowtime_Near dataItem `command:"pmccsg 0 0 0,pmcctg 0 0" commandLegacy:"pmcctg 0 0 0"`
	PM_ChannelCountersShowtime_Far  dataItem `command:"pmccsg 0 1 0,pmcctg 0 1" commandLegacy:"pmcctg 0 0 1"`
	PM_LineSecCountersShowtime_Near dataItem `command:"pmlscsg 0 0,pmlsctg 0" commandLegacy:"pmlsctg 0 0"`
	PM_LineSecCountersShowtime_Far  dataItem `command:"pmlscsg 1 0,pmlsctg 1" commandLegacy:"pmlsctg 0 1"`
	PM_ReTxCountersShowtimeGet_Near dataItem `command:"pmrtcsg 0 0,pmrtctg 0"`
	PM_ReTxCountersShowtimeGet_Far  dataItem `command:"pmrtcsg 1 0,pmrtctg 1"`
	ReTxStatistics_Near             dataItem `command:"rtsg 0"`
	ReTxStatistics_Far              dataItem `command:"rtsg 1"`

	PilotTonesStatus              dataItem `command:"ptsg"`
	BandBorderStatus_US           dataItem `command:"bbsg 0"`
	BandBorderStatus_DS           dataItem `command:"bbsg 1"`
	G997_BitAllocationNscShort_US dataItem `command:"g997bansg 0" commandLegacy:"g997banscsg 0 0"`
	G997_BitAllocationNscShort_DS dataItem `command:"g997bansg 1" commandLegacy:"g997banscsg 0 1"`
	G997_SnrAllocationNscShort_US dataItem `command:"g997sansg 0" commandLegacy:"g997snrnscsg 0 0"`
	G997_SnrAllocationNscShort_DS dataItem `command:"g997sansg 1" commandLegacy:"g997snrnscsg 0 1"`
	G997_DeltSNR_US               dataItem `command:"g997dsnrg 0 1" commandLegacy:"g997dsnrg 0 0"`
	G997_DeltSNR_DS               dataItem `command:"g997dsnrg 1 1" commandLegacy:"g997dsnrg 0 1"`
	G997_DeltQLN_US               dataItem `command:"g997dqlng 0 1" commandLegacy:"g997dqlng 0 0"`
	G997_DeltQLN_DS               dataItem `command:"g997dqlng 1 1" commandLegacy:"g997dqlng 0 1"`
	G997_DeltHLOG_US              dataItem `command:"g997dhlogg 0 1" commandLegacy:"g997dhlogg 0 0"`
	G997_DeltHLOG_DS              dataItem `command:"g997dhlogg 1 1" commandLegacy:"g997dhlogg 0 1"`
}

func (d *data) LoadData(e exec.Executor, command string) (err error) {
	var tagName = "command"

	vig, err := e.Execute(command + " vig")
	if exec.IsCommandNotFound(vig, err) {
		return errors.New("command not found, check the configuration")
	} else if err != nil {
		return err
	}
	vigData := parseValues(vig)

	if version, ok := vigData["DSL_DriverVersionApi"]; ok {
		d.APIVersion = version
	} else if version, ok := vigData["DSL_APILibraryVersion"]; ok {
		d.APIVersion = version
	} else {
		return errors.New("failed to read Lantiq API version")
	}

	if strings.HasPrefix(d.APIVersion, "2") {
		tagName = "commandLegacy"
	}

	t := reflect.TypeOf(*d)
	v := reflect.ValueOf(d)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if commands, ok := field.Tag.Lookup(tagName); ok {
			var fullCommand, out string
			var err error

			if commands == "" {
				continue
			}

			commandsSplit := strings.Split(commands, ",")
			for _, cmd := range commandsSplit {
				fullCommand = command + " " + cmd

				if cmd != "vig" {
					out, err = e.Execute(fullCommand)
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

			valCommand := val.FieldByName("Command")
			valCommand.SetString(fullCommand)

			valOutput := val.FieldByName("Output")
			valOutput.SetString(out)
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
		val := v.Elem().FieldByName(field.Name)

		if item, ok := val.Interface().(dataItem); ok {
			if item.Command != "" {
				fmt.Fprintf(&b, "# %s # %s\n", item.Command, field.Name)
				fmt.Fprintln(&b, item.Output)
			}
		}
	}

	fmt.Fprintln(&b)
	return []byte(b.String())
}
