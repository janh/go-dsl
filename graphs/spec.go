// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package graphs

type labelFormatFunc func(val, step, start, end int) string

type graphSpec struct {
	Width  int
	Height int

	ScaleFactor float64

	FontSize float64

	ColorBackground Color
	ColorForeground Color

	LegendXLabelDigits     float64
	LegendXLabelSteps      []int
	LegendXLabelStart      int
	LegendXLabelEnd        int
	LegendXLabelFormatFunc labelFormatFunc
	LegendXMin             float64
	LegendXMax             float64

	LegendYLabelDigits     float64
	LegendYLabelSteps      []int
	LegendYLabelStart      int
	LegendYLabelEnd        int
	LegendYLabelFormatFunc labelFormatFunc
	LegendYBottom          float64
	LegendYTop             float64

	LegendEnabled bool
	LegendData    Legend
}
