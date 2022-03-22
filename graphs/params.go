// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package graphs

type GraphParams struct {
	Width           int
	Height          int
	ScaleFactor     float64
	ColorBackground Color
	ColorForeground Color
}

func (p *GraphParams) normalize() {
	if p.ScaleFactor == 0 {
		p.ScaleFactor = DefaultScaleFactor
	}
}

var (
	DefaultWidth  = 554
	DefaultHeight = 114

	DefaultScaleFactor = 1.0

	DefaultColorBackground = Color{255, 255, 255, 1.0}
	DefaultColorForeground = Color{0, 0, 0, 1.0}

	DefaultGraphParams = GraphParams{
		Width:           DefaultWidth,
		Height:          DefaultHeight,
		ScaleFactor:     DefaultScaleFactor,
		ColorBackground: DefaultColorBackground,
		ColorForeground: DefaultColorForeground,
	}
)
