// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package graphs

type GraphParams struct {
	Width           int
	Height          int
	ColorBackground Color
	ColorForeground Color
}

var (
	DefaultWidth  = 554
	DefaultHeight = 114

	DefaultColorBackground = Color{255, 255, 255, 1.0}
	DefaultColorForeground = Color{0, 0, 0, 1.0}

	DefaultGraphParams = GraphParams{
		Width:           DefaultWidth,
		Height:          DefaultHeight,
		ColorBackground: DefaultColorBackground,
		ColorForeground: DefaultColorForeground,
	}
)
