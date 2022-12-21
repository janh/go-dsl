// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package common

import (
	"archive/zip"
	"io"

	"3e8.eu/go/dsl/graphs"
)

func WriteArchive(w io.Writer, filenameBase string, state StateChange, rawData bool) (err error) {
	archive := zip.NewWriter(w)
	defer func() {
		if closeErr := archive.Close(); closeErr != nil {
			err = closeErr
		}
	}()

	var fileWriter io.Writer

	fileWriter, err = archive.Create(filenameBase + "_summary.txt")
	if err != nil {
		return
	}
	_, err = io.WriteString(fileWriter, state.Status.Summary())
	if err != nil {
		return
	}

	if rawData {
		fileWriter, err = archive.Create(filenameBase + "_raw.txt")
		if err != nil {
			return
		}
		_, err = fileWriter.Write(state.RawData)
		if err != nil {
			return
		}
	}

	fileWriter, err = archive.Create(filenameBase + "_bits.svg")
	if err != nil {
		return
	}
	err = graphs.DrawBitsGraph(fileWriter, state.Bins, graphs.DefaultGraphParams)
	if err != nil {
		return
	}

	fileWriter, err = archive.Create(filenameBase + "_snr.svg")
	if err != nil {
		return
	}
	err = graphs.DrawSNRGraph(fileWriter, state.Bins, graphs.DefaultGraphParams)
	if err != nil {
		return
	}

	fileWriter, err = archive.Create(filenameBase + "_snr_minmax.svg")
	if err != nil {
		return
	}
	err = graphs.DrawSNRGraphWithHistory(fileWriter, state.Bins, state.BinsHistory, graphs.DefaultGraphParams)
	if err != nil {
		return
	}

	fileWriter, err = archive.Create(filenameBase + "_qln.svg")
	if err != nil {
		return
	}
	err = graphs.DrawQLNGraph(fileWriter, state.Bins, graphs.DefaultGraphParams)
	if err != nil {
		return
	}

	fileWriter, err = archive.Create(filenameBase + "_hlog.svg")
	if err != nil {
		return
	}
	err = graphs.DrawHlogGraph(fileWriter, state.Bins, graphs.DefaultGraphParams)
	if err != nil {
		return
	}

	return
}
