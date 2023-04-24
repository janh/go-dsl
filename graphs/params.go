// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package graphs

type GraphParams struct {
	Width                   int
	Height                  int
	ScaleFactor             float64
	FontSize                float64
	ColorBackground         Color
	ColorForeground         Color
	Legend                  bool
	PreferDynamicAxisLimits bool
}

func (p *GraphParams) normalize() {
	if p.ScaleFactor == 0 {
		p.ScaleFactor = DefaultScaleFactor
	}
}

var (
	DefaultWidth            = 560
	DefaultHeight           = 114
	DefaultHeightWithLegend = 132

	DefaultScaleFactor = 1.0

	DefaultFontSize = 0.0

	DefaultColorBackground = Color{255, 255, 255, 1.0}
	DefaultColorForeground = Color{0, 0, 0, 1.0}

	DefaultGraphParams = GraphParams{
		Width:                   DefaultWidth,
		Height:                  DefaultHeight,
		ScaleFactor:             DefaultScaleFactor,
		FontSize:                DefaultFontSize,
		ColorBackground:         DefaultColorBackground,
		ColorForeground:         DefaultColorForeground,
		Legend:                  false,
		PreferDynamicAxisLimits: false,
	}

	DefaultGraphParamsWithLegend = GraphParams{
		Width:                   DefaultWidth,
		Height:                  DefaultHeightWithLegend,
		ScaleFactor:             DefaultScaleFactor,
		FontSize:                DefaultFontSize,
		ColorBackground:         DefaultColorBackground,
		ColorForeground:         DefaultColorForeground,
		Legend:                  true,
		PreferDynamicAxisLimits: false,
	}
)
