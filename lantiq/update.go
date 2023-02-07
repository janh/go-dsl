// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lantiq

import (
	"3e8.eu/go/dsl/internal/exec"
	"3e8.eu/go/dsl/models"
)

func updateData(e exec.Executor, command string) (status models.Status, bins models.Bins, rawData []byte, err error) {
	var data data

	err = data.LoadData(e, command)
	if err != nil {
		return
	}

	status = parseBasicStatus(&data)
	bins = parseBins(&status, &data)
	parseExtendedStatus(&status, &bins, &data)
	rawData = data.RawData()

	return
}
