// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package broadcom

import (
	"fmt"
	"strings"

	"3e8.eu/go/dsl/models"
)

type executor interface {
	Execute(cmd string) (string, error)
}

func updateData(e executor) (status models.Status, bins models.Bins, rawData []byte, err error) {
	stats, err := e.Execute("xdslctl info --stats")
	if err != nil {
		return
	}

	vendor, err := e.Execute("xdslctl info --vendor")
	if err != nil {
		return
	}

	version, err := e.Execute("xdslctl --version")
	if err != nil {
		return
	}

	pbParams, err := e.Execute("xdslctl info --pbParams")
	if err != nil {
		return
	}

	bits, err := e.Execute("xdslctl info --Bits")
	if err != nil {
		return
	}

	snr, err := e.Execute("xdslctl info --SNR")
	if err != nil {
		return
	}

	qln, err := e.Execute("xdslctl info --QLN")
	if err != nil {
		return
	}

	hlog, err := e.Execute("xdslctl info --Hlog")
	if err != nil {
		return
	}

	status = parseStatus(stats, vendor, version)
	bins = parseBins(status, pbParams, bits, snr, qln, hlog)

	var b strings.Builder
	fmt.Fprintln(&b, "# xdslctl info --stats")
	fmt.Fprintln(&b, stats)
	fmt.Fprintln(&b, "# xdslctl info --vendor")
	fmt.Fprintln(&b, vendor)
	fmt.Fprintln(&b, "# xdslctl info --version")
	fmt.Fprintln(&b, version)
	fmt.Fprintln(&b, "# xdslctl info --pbParams")
	fmt.Fprintln(&b, pbParams)
	fmt.Fprintln(&b, "# xdslctl info --Bits")
	fmt.Fprintln(&b, bits)
	fmt.Fprintln(&b, "# xdslctl info --SNR")
	fmt.Fprintln(&b, snr)
	fmt.Fprintln(&b, "# xdslctl info --QLN")
	fmt.Fprintln(&b, qln)
	fmt.Fprintln(&b, "# xdslctl info --Hlog")
	fmt.Fprintln(&b, hlog)
	fmt.Fprintln(&b)
	rawData = []byte(b.String())

	return
}
