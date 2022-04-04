// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package common

import (
	"bytes"
	"html/template"

	jsgraphs "3e8.eu/go/dsl/graphs/javascript"
	"3e8.eu/go/dsl/models"
)

func getSummaryString(status models.Status) string {
	buf := new(bytes.Buffer)

	tpl := template.Must(template.ParseFS(Files, "res/summary.html"))
	tpl.Execute(buf, status)

	return buf.String()
}

func GetStateMessage(change StateChange) Message {
	msg := Message{State: string(change.State)}

	switch change.State {

	case StateReady:
		msg.Data = MessageData{
			Summary: getSummaryString(change.Status),
			Bins:    jsgraphs.EncodeBins(change.Bins),
			History: jsgraphs.EncodeBinsHistory(change.BinsHistory),
		}

	case StatePassphraseRequired:
		msg.Data = change.Fingerprint

	case StateError:
		msg.Data = "failed to load data from device: " + change.Err.Error()

	}

	return msg
}
