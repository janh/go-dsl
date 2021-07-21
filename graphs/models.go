// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package graphs

type baseModel struct {
	Width  float64
	Height float64

	GraphX      float64
	GraphY      float64
	GraphWidth  float64
	GraphHeight float64

	ColorBackground Color
	ColorText       Color
	ColorGraph      Color
	ColorGrid       Color
	ColorNeutral    Color
	ColorUpstream   Color
	ColorDownstream Color

	PathLegend path
	PathGrid   path

	LabelsX          []label
	LabelsY          []label
	LabelsYTransform transform
}

type label struct {
	X    float64
	Y    float64
	Text string
}

type bitsModel struct {
	baseModel
	Transform      transform
	PathUpstream   path
	PathDownstream path
}

type snrModel struct {
	baseModel
	Transform transform
	Path      path
}

type qlnModel struct {
	baseModel
	Transform transform
	Path      path
}

type hlogModel struct {
	baseModel
	StrokeWidth float64
	Transform   transform
	Path        path
}
