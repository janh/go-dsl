// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package draytek

import (
	"fmt"
	"strings"

	"3e8.eu/go/dsl/models"
)

type executor interface {
	Execute(cmd string) (string, error)
}

func updateData(e executor) (statusData models.Status, bins models.Bins, rawData []byte, err error) {
	status, err := e.Execute("adsl status")
	if err != nil {
		return
	}

	counts, err := e.Execute("adsl status counts")
	if err != nil {
		return
	}

	more, err := e.Execute("adsl status more")
	if err != nil {
		return
	}

	olr, err := e.Execute("adsl status olr")
	if err != nil {
		return
	}

	bandinfo, err := e.Execute("adsl status bandinfo")
	if err != nil {
		return
	}

	downstream, err := e.Execute("adsl showbins")
	if err != nil {
		return
	}

	upstream, err := e.Execute("adsl showbins up")
	if err != nil {
		return
	}

	qln, err := e.Execute("adsl status qln")
	if err != nil {
		return
	}

	hlog, err := e.Execute("adsl status hlog")
	if err != nil {
		return
	}

	statusData = parseStatus(status, counts, more, olr)
	bins = parseBins(statusData, bandinfo, downstream, upstream, qln, hlog)

	var b strings.Builder
	fmt.Fprintln(&b, "# adsl status")
	fmt.Fprintln(&b, status)
	fmt.Fprintln(&b, "# adsl status counts")
	fmt.Fprintln(&b, counts)
	fmt.Fprintln(&b, "# adsl status more")
	fmt.Fprintln(&b, more)
	fmt.Fprintln(&b, "# adsl status olr")
	fmt.Fprintln(&b, olr)
	fmt.Fprintln(&b, "# adsl status bandinfo")
	fmt.Fprintln(&b, bandinfo)
	fmt.Fprintln(&b, "# adsl showbins")
	fmt.Fprintln(&b, downstream)
	fmt.Fprintln(&b, "# adsl showbins up")
	fmt.Fprintln(&b, upstream)
	fmt.Fprintln(&b, "# adsl status qln")
	fmt.Fprintln(&b, qln)
	fmt.Fprintln(&b, "# adsl status hlog")
	fmt.Fprintln(&b, hlog)
	fmt.Fprintln(&b)
	rawData = []byte(b.String())

	return
}
