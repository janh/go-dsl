// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package helpers

import (
	"strings"

	"3e8.eu/go/dsl/models"
)

func ParseStateTR06X(str string) models.State {
	str = strings.ToLower(strings.TrimSpace(str))

	switch str {

	case "disabled", "nosignal", "initializing":
		return models.StateDown

	case "establishinglink":
		return models.StateInit

	case "up":
		return models.StateShowtime

	case "error":
		return models.StateError

	}

	return models.StateUnknown
}
