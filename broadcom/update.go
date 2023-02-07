// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package broadcom

import (
	"fmt"
	"strings"

	"3e8.eu/go/dsl/internal/exec"
	"3e8.eu/go/dsl/models"
)

func updateData(e exec.Executor, command string) (status models.Status, bins models.Bins, rawData []byte, err error) {
	if command == "" {
		command = "xdslctl"
	}

	stats, err := e.Execute(command + " info --stats")
	if err != nil {
		return
	}

	vectoring, err := e.Execute(command + " info --vectoring")
	if err != nil {
		return
	}

	vendor, err := e.Execute(command + " info --vendor")
	if err != nil {
		return
	}

	version, err := e.Execute(command + " --version")
	if err != nil {
		return
	}

	pbParams, err := e.Execute(command + " info --pbParams")
	if err != nil {
		return
	}

	bits, err := e.Execute(command + " info --Bits")
	if err != nil {
		return
	}

	snr, err := e.Execute(command + " info --SNR")
	if err != nil {
		return
	}

	qln, err := e.Execute(command + " info --QLN")
	if err != nil {
		return
	}

	hlog, err := e.Execute(command + " info --Hlog")
	if err != nil {
		return
	}

	status = parseStatus(stats, vectoring, vendor, version)
	bins = parseBins(status, pbParams, bits, snr, qln, hlog)

	var b strings.Builder
	fmt.Fprintln(&b, "# xdslctl info --stats")
	fmt.Fprintln(&b, stats)
	fmt.Fprintln(&b, "# xdslctl info --vectoring")
	fmt.Fprintln(&b, vectoring)
	fmt.Fprintln(&b, "# xdslctl info --vendor")
	fmt.Fprintln(&b, vendor)
	fmt.Fprintln(&b, "# xdslctl --version")
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
