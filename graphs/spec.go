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

	LegendXLabelStep   int
	LegendXLabelStart  int
	LegendXLabelEnd    int
	LegendXLabelFactor float64
	LegendXLabelFormat string
	LegendXMin         float64
	LegendXMax         float64

	LegendYLabelStep  int
	LegendYLabelStart int
	LegendYLabelEnd   int
	LegendYBottom     float64
	LegendYTop        float64
}
