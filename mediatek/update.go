// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package mediatek

import (
	"fmt"
	"strings"

	"3e8.eu/go/dsl/models"
)

type executor interface {
	Execute(cmd string) (string, error)
}

func updateData(e executor) (status models.Status, bins models.Bins, rawData []byte, err error) {
	adslStats, err := e.Execute("cat /proc/tc3162/adsl_stats")
	if err != nil {
		return
	}

	vdslInterfaceConfig, err := e.Execute("cat /proc/tc3162/vdsl_interface_config")
	if err != nil {
		return
	}

	adslFwVer, err := e.Execute("cat /proc/tc3162/adsl_fwver")
	if err != nil {
		return
	}

	adslShowbpcDs, err := e.Execute("cat /proc/tc3162/adsl_showbpc_ds")
	if err != nil {
		return
	}

	adslShowbpcUs, err := e.Execute("cat /proc/tc3162/adsl_showbpc_us")
	if err != nil {
		return
	}

	adslShowsnr, err := e.Execute("cat /proc/tc3162/adsl_showsnr")
	if err != nil {
		return
	}

	vdslShowbpcDs, err := e.Execute("cat /proc/tc3162/vdsl_showbpc_ds")
	if err != nil {
		return
	}

	vdslShowbpcUs, err := e.Execute("cat /proc/tc3162/vdsl_showbpc_us")
	if err != nil {
		return
	}

	vdslShowsnr, err := e.Execute("cat /proc/tc3162/vdsl_showsnr")
	if err != nil {
		return
	}

	status = parseStatus(adslStats, vdslInterfaceConfig, adslFwVer)
	bins = parseBins(status, adslShowbpcDs, adslShowbpcUs, adslShowsnr, vdslShowbpcDs, vdslShowbpcUs, vdslShowsnr)

	var b strings.Builder
	fmt.Fprintln(&b, "# cat /proc/tc3162/adsl_stats")
	fmt.Fprintln(&b, adslStats)
	fmt.Fprintln(&b, "# cat /proc/tc3162/vdsl_interface_config")
	fmt.Fprintln(&b, vdslInterfaceConfig)
	fmt.Fprintln(&b, "# cat /proc/tc3162/adsl_fwver")
	fmt.Fprintln(&b, adslFwVer)
	fmt.Fprintln(&b, "# cat /proc/tc3162/adsl_showbpc_ds")
	fmt.Fprintln(&b, adslShowbpcDs)
	fmt.Fprintln(&b, "# cat /proc/tc3162/adsl_showbpc_us")
	fmt.Fprintln(&b, adslShowbpcUs)
	fmt.Fprintln(&b, "# cat /proc/tc3162/adsl_showsnr")
	fmt.Fprintln(&b, adslShowsnr)
	fmt.Fprintln(&b, "# cat /proc/tc3162/vdsl_showbpc_ds")
	fmt.Fprintln(&b, vdslShowbpcDs)
	fmt.Fprintln(&b, "# cat /proc/tc3162/vdsl_showbpc_us")
	fmt.Fprintln(&b, vdslShowbpcUs)
	fmt.Fprintln(&b, "# cat /proc/tc3162/vdsl_showsnr")
	fmt.Fprintln(&b, vdslShowsnr)
	fmt.Fprintln(&b)
	rawData = []byte(b.String())

	return
}
