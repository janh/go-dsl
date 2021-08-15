// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package models

import (
	"fmt"
	"strings"
	"time"
)

type Duration struct {
	time.Duration
}

func (d Duration) String() string {
	if d.Duration <= 0 {
		return "-"
	}

	minutes := int64(d.Minutes()) % 60
	hours := int64(d.Hours()) % 24
	days := int64(d.Hours()) / 24

	parts := make([]string, 0, 3)

	if days == 1 {
		parts = append(parts, "1 day")
	} else if days > 1 {
		parts = append(parts, fmt.Sprintf("%d days", days))
	}

	if hours == 1 {
		parts = append(parts, "1 hour")
	} else if days > 0 || hours > 1 {
		parts = append(parts, fmt.Sprintf("%d hours", hours))
	}

	if minutes == 1 {
		parts = append(parts, "1 minute")
	} else {
		parts = append(parts, fmt.Sprintf("%d minutes", minutes))
	}

	return strings.Join(parts, ", ")
}
