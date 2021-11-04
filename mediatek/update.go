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

	wanVdsl2Mgcnt, err := e.Execute(`wan vdsl2 show mgcnt; dmesg | sed -n 'H;/near-end path0 fec/h;${g;p}'`)
	if err != nil {
		return
	}

	wanVdsl2PmsPmdRx, err := e.Execute(`wan vdsl2 show pms_pmd rx; dmesg | sed -n 'H;/<<< RX PMSTC Parameters >>>/h;${g;p}'`)
	if err != nil {
		return
	}

	wanVdsl2PmsPmdTx, err := e.Execute(`wan vdsl2 show pms_pmd tx; dmesg | sed -n 'H;/<<< TX PMSTC Parameters >>>/h;${g;p}'`)
	if err != nil {
		return
	}

	wanVdsl2Dmt, err := e.Execute(`wan vdsl2 show dmt; dmesg | sed -n 'H;/TX_MOD_PARAMS/h;${g;p}'`)
	if err != nil {
		return
	}

	wanVdsl2Qln, err := e.Execute(`wan vdsl2 show pmdtestparam qln; dmesg | sed -n 'H;/Qln:/h;${g;p}'`)
	if err != nil {
		return
	}

	wanVdsl2Hlog, err := e.Execute(`wan vdsl2 show pmdtestparam hlog; dmesg | sed -n 'H;/Hlog:/h;${g;p}'`)
	if err != nil {
		return
	}

	status = parseStatus(adslStats, vdslInterfaceConfig, adslFwVer,
		wanVdsl2Mgcnt, wanVdsl2PmsPmdRx, wanVdsl2PmsPmdTx)

	bins = parseBins(status,
		adslShowbpcDs, adslShowbpcUs, adslShowsnr,
		vdslShowbpcDs, vdslShowbpcUs, vdslShowsnr,
		wanVdsl2Dmt, wanVdsl2Qln, wanVdsl2Hlog)

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
	fmt.Fprintln(&b, "# wan vdsl2 show mgcnt")
	fmt.Fprintln(&b, wanVdsl2Mgcnt)
	fmt.Fprintln(&b, "# wan vdsl2 show pms_pmd rx")
	fmt.Fprintln(&b, wanVdsl2PmsPmdRx)
	fmt.Fprintln(&b, "# wan vdsl2 show pms_pmd tx")
	fmt.Fprintln(&b, wanVdsl2PmsPmdTx)
	fmt.Fprintln(&b, "# wan vdsl2 show dmt")
	fmt.Fprintln(&b, wanVdsl2Dmt)
	fmt.Fprintln(&b, "# wan vdsl2 show pmdtestparam qln")
	fmt.Fprintln(&b, wanVdsl2Qln)
	fmt.Fprintln(&b, "# wan vdsl2 show pmdtestparam hlog")
	fmt.Fprintln(&b, wanVdsl2Hlog)
	fmt.Fprintln(&b)
	rawData = []byte(b.String())

	return
}
