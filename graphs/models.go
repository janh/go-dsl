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

	PathLegend Path
	PathGrid   Path

	LabelsX          []label
	LabelsY          []label
	LabelsYTransform Transform
}

type label struct {
	X    float64
	Y    float64
	Text string
}

type bitsModel struct {
	baseModel
	Transform      Transform
	PathNeutral    Path
	PathUpstream   Path
	PathDownstream Path
}

type snrModel struct {
	baseModel
	Transform Transform
	Path      Path
}

type qlnModel struct {
	baseModel
	Transform Transform
	Path      Path
}

type hlogModel struct {
	baseModel
	StrokeWidth float64
	Transform   Transform
	Path        Path
}
