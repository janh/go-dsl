// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package bintecelmeg

import (
	"3e8.eu/go/dsl/internal/exec"
	"3e8.eu/go/dsl/models"
)

func updateData(e exec.Executor) (status models.Status, bins models.Bins, rawData []byte, err error) {
	data, err := e.Execute("dsl -v status")
	if err != nil {
		return
	}

	sections := parseSections(data)

	connection := parseKeyValueItems(sections["connection"])
	localModem := parseKeyValueItems(sections["localmodem"])
	remoteModem := parseKeyValueItems(sections["remotemodem"])
	receiveStatistics := parseStatisticsItems(sections["receivestatistics"])
	transmitStatistics := parseStatisticsItems(sections["transmitstatistics"])

	status = interpretStatus(connection, localModem, remoteModem, receiveStatistics, transmitStatistics)
	bins = interpretBins(&status, receiveStatistics, transmitStatistics)
	rawData = []byte(data)

	return
}
