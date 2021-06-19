// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package graphs

import (
	"fmt"
)

type Color struct {
	R uint8
	G uint8
	B uint8
	A float64
}

func (c Color) String() string {
	return fmt.Sprintf("rgba(%d,%d,%d,%f)", c.R, c.G, c.B, c.A)
}
