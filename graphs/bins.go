// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package graphs

import (
	"fmt"
	"io"
	"math"

	"3e8.eu/go/dsl/models"
)

var (
	colorUpstream   = Color{96, 192, 0, .75}
	colorDownstream = Color{0, 127, 255, .75}
	colorNeutral    = Color{127, 127, 127, .75}
)

func getGraphColors(background, foreground Color) (colorGraph, colorGrid Color) {
	brightnessBackground := 0.299*float64(background.R) + 0.587*float64(background.G) + 0.114*float64(background.B)
	brightnessForeground := 0.299*float64(foreground.R) + 0.587*float64(foreground.G) + 0.114*float64(foreground.B)
	brightness := brightnessBackground
	if background.A < 0.75 {
		brightness = 255 - brightnessForeground
	}

	var gray float64
	if brightness > 223 {
		gray = brightness - 20
	} else if brightness > 127 {
		gray = 255 - (223-brightness)/2
	} else if brightness > 31 {
		gray = 0 + (brightness-32)/2
	} else {
		gray = brightness + 20
	}

	var grayGrid float64
	if brightnessForeground < brightnessBackground {
		grayGrid = math.Max(gray-20, 0)
	} else {
		grayGrid = math.Min(gray+20, 255)
	}

	colorGraph = Color{uint8(gray), uint8(gray), uint8(gray), 1.0}
	colorGrid = Color{uint8(grayGrid), uint8(grayGrid), uint8(grayGrid), 1.0}

	return
}

func getBaseModel(spec graphSpec) baseModel {
	m := baseModel{}

	m.Width = float64(spec.Width)
	m.Height = float64(spec.Height)

	m.GraphX = 28.0
	m.GraphY = 4.0
	m.GraphWidth = m.Width - 42.0
	m.GraphHeight = m.Height - 23.0

	m.ColorBackground = spec.ColorBackground
	m.ColorText = spec.ColorForeground

	m.ColorGraph, m.ColorGrid = getGraphColors(spec.ColorBackground, spec.ColorForeground)

	m.ColorNeutral = colorNeutral
	m.ColorUpstream = colorUpstream
	m.ColorDownstream = colorDownstream

	textOffset := 3.5

	x := m.GraphX
	y := m.GraphY
	w := m.GraphWidth
	h := m.GraphHeight

	// legend for x-axis
	m.PathLegend.MoveTo(x-0.5, y+h+0.5)
	m.PathLegend.LineTo(x-0.5+w, y+h+0.5)
	for i := 0.0; i <= spec.LegendXMax; i += float64(spec.LegendXStep) {
		frac := i / spec.LegendXMax
		pos := x - 0.5 + math.Round(w*frac)
		m.PathLegend.MoveTo(pos, y+h+2.5)
		m.PathLegend.LineTo(pos, y+h+1.5)
		text := fmt.Sprintf(spec.LegendXFormat, i*spec.LegendXFactor)
		m.LabelsX = append(m.LabelsX, label{X: pos, Y: y + h + 10.5 + textOffset, Text: text})
	}

	// legend for y-axis
	if math.Max(math.Abs(float64(spec.LegendYLabelStart)), math.Abs(float64(spec.LegendYLabelEnd))) >= 100 {
		m.LabelsYTransform.Translate(10.5-x, 0)
		m.LabelsYTransform.Scale(0.7, 1)
		m.LabelsYTransform.Translate(x-10.5, 0)
	}
	m.PathLegend.MoveTo(x-0.5, y+0.5)
	m.PathLegend.LineTo(x-0.5, y+h+0.5)
	for i := spec.LegendYLabelStart + spec.LegendYLabelStep/2; i <= spec.LegendYLabelEnd; i += spec.LegendYLabelStep {
		frac := (float64(i) - spec.LegendYBottom) / (spec.LegendYTop - spec.LegendYBottom)
		pos := y + h + 0.5 - math.Round(h*frac)
		m.PathLegend.MoveTo(x-2.5, pos)
		m.PathLegend.LineTo(x-1.5, pos)
	}
	for i := spec.LegendYLabelStart; i <= spec.LegendYLabelEnd; i += spec.LegendYLabelStep {
		frac := (float64(i) - spec.LegendYBottom) / (spec.LegendYTop - spec.LegendYBottom)
		pos := y + h + 0.5 - math.Round(h*frac)
		m.PathLegend.MoveTo(x-4.5, pos)
		m.PathLegend.LineTo(x-1.5, pos)
		if frac > 0.01 {
			m.PathGrid.MoveTo(x+0.5, pos)
			m.PathGrid.LineTo(x+w-0.5, pos)
		}
		text := fmt.Sprintf("%d", i)
		m.LabelsY = append(m.LabelsY, label{X: x - 10.5, Y: pos + textOffset, Text: text})
	}

	return m
}

func getLegendX(mode models.Mode) (bins int, step int, freq float64) {
	bins = mode.BinCount()
	freq = mode.CarrierSpacing()

	switch bins {
	case 3479:
		step = 256
	case 2783:
		step = 192
	case 1972:
		step = 128
	default:
		step = bins / 16
	}

	return
}

func DrawBitsGraph(out io.Writer, data models.Bins, params GraphParams) error {
	bins, step, _ := getLegendX(data.Mode)

	spec := graphSpec{
		Width:             params.Width,
		Height:            params.Height,
		ColorBackground:   params.ColorBackground,
		ColorForeground:   params.ColorForeground,
		LegendXMax:        float64(bins),
		LegendXStep:       step,
		LegendXFactor:     1.0,
		LegendXFormat:     "%.0f",
		LegendYBottom:     0,
		LegendYTop:        15.166666667,
		LegendYLabelStart: 0,
		LegendYLabelEnd:   15,
		LegendYLabelStep:  2,
	}

	m := bitsModel{}
	m.baseModel = getBaseModel(spec)

	x := m.GraphX
	y := m.GraphY
	w := m.GraphWidth
	h := m.GraphHeight

	scaleX := w / spec.LegendXMax
	scaleY := h / spec.LegendYTop

	var lastBits int8
	var lastPosY float64
	var lastPath *path

	for i := 0; i < bins; i++ {
		bin := data.Bins[i]
		bits := bin.Bits
		if bits < 0 {
			bits = 0
		}

		var path *path
		switch bin.Type {
		case models.BinTypeNone:
			path = &m.PathNeutral
		case models.BinTypeUpstream:
			path = &m.PathUpstream
		case models.BinTypeDownstream:
			path = &m.PathDownstream
		default:
			continue
		}

		posX := float64(i)

		if lastBits > 0 && (bits == 0 || lastPath != path) {
			lastPath.LineTo(posX, lastPosY)
			lastPath.LineTo(posX, h)
			lastPath.Close()
		}
		if (lastPath != path || lastBits == 0) && bits > 0 {
			path.MoveTo(posX, h)
		}
		if bits > 0 {
			posY := h - math.Ceil(float64(bits)*scaleY)
			if lastPath != path || lastBits != bits {
				if lastPath == path && lastBits != 0 {
					path.LineTo(posX, lastPosY)
				}
				path.LineTo(posX, posY)
				lastPosY = posY
			}
		}

		lastBits = bits
		lastPath = path
	}

	if lastBits > 0 {
		lastPath.LineTo(spec.LegendXMax, lastPosY)
		lastPath.LineTo(spec.LegendXMax, h)
		lastPath.Close()
	}

	m.Transform.Scale(scaleX, 1)
	m.Transform.Translate(x, y)

	return writeTemplate(out, m, templateBase, templateBits)
}

func DrawSNRGraph(out io.Writer, data models.Bins, params GraphParams) error {
	bins, step, freq := getLegendX(data.Mode)

	spec := graphSpec{
		Width:             params.Width,
		Height:            params.Height,
		ColorBackground:   params.ColorBackground,
		ColorForeground:   params.ColorForeground,
		LegendXMax:        float64(bins),
		LegendXStep:       step,
		LegendXFactor:     freq / 1000,
		LegendXFormat:     "%.1f",
		LegendYBottom:     0,
		LegendYTop:        65,
		LegendYLabelStart: 0,
		LegendYLabelEnd:   65,
		LegendYLabelStep:  10,
	}

	m := snrModel{}
	m.baseModel = getBaseModel(spec)

	x := m.GraphX
	y := m.GraphY
	w := m.GraphWidth
	h := m.GraphHeight

	scaleX := w / spec.LegendXMax
	scaleY := h / spec.LegendYTop

	m.Path.SetPrecision(1)

	var last float64
	var lastPosY float64

	for i := 0; i < bins; i++ {
		bin := data.Bins[i]
		snr := bin.SNR
		if snr < 0 {
			snr = 0
		}

		posX := float64(i)

		if last > 0 && snr == 0 {
			m.Path.LineTo(posX, lastPosY)
			m.Path.LineTo(posX, h)
			m.Path.Close()
		}
		if last == 0 && snr > 0 {
			m.Path.MoveTo(posX, h)
		}
		if snr > 0 && last != snr {
			posY := h - math.Min(h, snr*scaleY)
			if last != 0 {
				m.Path.LineTo(posX, lastPosY)
			}
			m.Path.LineTo(posX, posY)
			lastPosY = posY
		}

		last = snr
	}

	if last > 0 {
		m.Path.LineTo(spec.LegendXMax, lastPosY)
		m.Path.LineTo(spec.LegendXMax, h)
		m.Path.Close()
	}

	m.Transform.Scale(scaleX, 1)
	m.Transform.Translate(x, y)

	return writeTemplate(out, m, templateBase, templateSNR)
}

func DrawQLNGraph(out io.Writer, data models.Bins, params GraphParams) error {
	bins, step, freq := getLegendX(data.Mode)

	spec := graphSpec{
		Width:             params.Width,
		Height:            params.Height,
		ColorBackground:   params.ColorBackground,
		ColorForeground:   params.ColorForeground,
		LegendXMax:        float64(bins),
		LegendXStep:       step,
		LegendXFactor:     freq / 1000,
		LegendXFormat:     "%.1f",
		LegendYBottom:     -160,
		LegendYTop:        -69,
		LegendYLabelStart: -160,
		LegendYLabelEnd:   -70,
		LegendYLabelStep:  20,
	}

	m := qlnModel{}
	m.baseModel = getBaseModel(spec)

	x := m.GraphX
	y := m.GraphY
	w := m.GraphWidth
	h := m.GraphHeight

	scaleX := w / spec.LegendXMax
	scaleY := h / (spec.LegendYTop - spec.LegendYBottom)
	offsetY := spec.LegendYBottom

	m.Path.SetPrecision(1)

	var last float64 = offsetY
	var lastPosY float64

	for i := 0; i < bins; i++ {
		bin := data.Bins[i]
		qln := bin.QLN
		if qln >= 0 {
			qln = offsetY
		}

		posX := float64(i)

		if last > offsetY && qln <= offsetY {
			m.Path.LineTo(posX, lastPosY)
			m.Path.LineTo(posX, h)
			m.Path.Close()
		}
		if last <= offsetY && qln > offsetY {
			m.Path.MoveTo(posX, h)
		}
		if qln > offsetY && last != qln {
			posY := h - math.Max(0, math.Min(h, (qln-offsetY)*scaleY))
			if last > offsetY {
				m.Path.LineTo(posX, lastPosY)
			}
			m.Path.LineTo(posX, posY)
			lastPosY = posY
		}

		last = qln
	}

	if last > offsetY {
		m.Path.LineTo(spec.LegendXMax, lastPosY)
		m.Path.LineTo(spec.LegendXMax, h)
		m.Path.Close()
	}

	m.Transform.Scale(scaleX, 1)
	m.Transform.Translate(x, y)

	return writeTemplate(out, m, templateBase, templateQLN)
}

func DrawHlogGraph(out io.Writer, data models.Bins, params GraphParams) error {
	bins, step, freq := getLegendX(data.Mode)

	spec := graphSpec{
		Width:             params.Width,
		Height:            params.Height,
		ColorBackground:   params.ColorBackground,
		ColorForeground:   params.ColorForeground,
		LegendXMax:        float64(bins),
		LegendXStep:       step,
		LegendXFactor:     freq / 1000,
		LegendXFormat:     "%.1f",
		LegendYBottom:     -100,
		LegendYTop:        5,
		LegendYLabelStart: -100,
		LegendYLabelEnd:   0,
		LegendYLabelStep:  20,
	}

	m := hlogModel{}
	m.baseModel = getBaseModel(spec)

	x := m.GraphX
	y := m.GraphY
	w := m.GraphWidth
	h := m.GraphHeight

	scaleX := w / spec.LegendXMax
	scaleY := h / (spec.LegendYTop - spec.LegendYBottom)
	offsetY := spec.LegendYBottom

	m.Path.SetPrecision(1)

	var lastValid bool
	var last float64
	var lastPosY float64

	for i := 0; i < bins; i++ {
		bin := data.Bins[i]
		hlog := bin.Hlog

		posX := float64(i) + 0.5
		posY := h + 0.5 - math.Max(0, math.Min(h, (hlog-offsetY)*scaleY))

		reset := lastValid && math.Abs(hlog-last) >= 10

		if (last < 0 && hlog >= 0) || reset {
			m.Path.LineTo(posX-0.5, lastPosY/scaleX)
		}
		if (last >= 0 && hlog < 0) || reset {
			m.Path.MoveTo(posX-0.5, posY/scaleX)
			lastPosY = posY
		}
		if hlog < 0 && last != hlog {
			if lastValid && !reset {
				m.Path.LineTo(posX-1, lastPosY/scaleX)
				m.Path.LineTo(posX, posY/scaleX)
			}
			lastPosY = posY
		}

		lastValid = hlog < 0
		last = hlog
	}

	if last < 0 {
		m.Path.LineTo(spec.LegendXMax, lastPosY/scaleX)
	}

	// scaling of y by scaleX in order to simulate vector-effect="non-scaling-stroke" for non-supporting renderers
	m.Transform.Scale(scaleX, scaleX)
	m.Transform.Translate(x, y)

	m.StrokeWidth = 1.25 / scaleX

	return writeTemplate(out, m, templateBase, templateHlog)
}
