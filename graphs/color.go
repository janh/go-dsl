// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package graphs

import (
	"fmt"
	"math"
	"strconv"

	"html/template"
)

type Color struct {
	R uint8
	G uint8
	B uint8
	A float64
}

func (c Color) String() string {
	a := math.Round(c.A*1000) / 1000
	aStr := strconv.FormatFloat(a, 'f', -1, 64)
	return fmt.Sprintf("rgba(%d,%d,%d,%s)", c.R, c.G, c.B, aStr)
}

func (c Color) ColorHex() string {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

func (c Color) Opacity() string {
	a := math.Round(c.A*1000) / 1000
	aStr := strconv.FormatFloat(a, 'f', -1, 64)
	if len(aStr) > 1 && aStr[0] == '0' {
		aStr = aStr[1:]
	}
	return aStr
}

func (c Color) CSS() template.CSS {
	return template.CSS(c.String())
}
