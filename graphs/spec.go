// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package graphs

type graphSpec struct {
	Width  int
	Height int

	ScaleFactor float64

	FontSize float64

	ColorBackground Color
	ColorForeground Color

	LegendXStep   int
	LegendXMax    float64
	LegendXFactor float64
	LegendXFormat string

	LegendYLabelStep  int
	LegendYLabelStart int
	LegendYLabelEnd   int
	LegendYBottom     float64
	LegendYTop        float64
}
