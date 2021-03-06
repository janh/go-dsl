// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package graphs

import (
	"fmt"
	"io"
	"math"
	"sort"

	"3e8.eu/go/dsl/models"
)

var (
	colorUpstream   = Color{96, 192, 0, .75}
	colorDownstream = Color{0, 127, 255, .75}
	colorPilotTones = Color{204, 94, 82, .75}
)

func getGraphColors(background, foreground Color) (colorGraph, colorGrid, colorNeutralFill, colorNeutralStroke Color) {
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

	var grayNeutral float64
	if brightness > 127 {
		grayNeutral = 95
	} else {
		grayNeutral = 159
	}

	colorGraph = Color{uint8(gray), uint8(gray), uint8(gray), 1.0}
	colorGrid = Color{uint8(grayGrid), uint8(grayGrid), uint8(grayGrid), 1.0}

	colorNeutralFill = Color{uint8(grayNeutral), uint8(grayNeutral), uint8(grayNeutral), .6}
	colorNeutralStroke = Color{uint8(grayNeutral), uint8(grayNeutral), uint8(grayNeutral), .75}

	return
}

func getBaseModel(spec graphSpec) baseModel {
	m := baseModel{}

	m.ScaledWidth = float64(spec.Width) / spec.ScaleFactor
	m.ScaledHeight = float64(spec.Height) / spec.ScaleFactor

	m.Width = float64(spec.Width)
	m.Height = float64(spec.Height)

	var fontFactor float64
	if spec.FontSize == 0 {
		factor := math.Min(m.ScaledWidth/554, m.ScaledHeight/114)
		fontFactor = math.Min(math.Max(1.0, factor), 1.35)
		m.FontSize = 10.5 * fontFactor * spec.ScaleFactor
	} else {
		fontFactor = spec.FontSize / 10.5
		m.FontSize = spec.FontSize * spec.ScaleFactor
	}

	m.GraphX = math.Round((23.0*fontFactor + 5.0) * spec.ScaleFactor)
	m.GraphY = math.Round(4.0 * fontFactor * spec.ScaleFactor)
	m.GraphWidth = m.Width - math.Round((38.0*fontFactor+4.0)*spec.ScaleFactor)
	m.GraphHeight = m.Height - math.Round((18.0*fontFactor+5.0)*spec.ScaleFactor)

	m.ColorBackground = spec.ColorBackground
	m.ColorText = spec.ColorForeground

	m.ColorGraph, m.ColorGrid, m.ColorNeutralFill, m.ColorNeutralStroke =
		getGraphColors(spec.ColorBackground, spec.ColorForeground)

	m.ColorMinStroke = colorDownstream
	m.ColorMaxStroke = colorUpstream

	m.ColorUpstream = colorUpstream
	m.ColorDownstream = colorDownstream

	m.ColorPilotTones = colorPilotTones

	if spec.ScaleFactor > 1.0 {
		m.StrokeWidthBase = math.Round(spec.ScaleFactor)
	} else {
		m.StrokeWidthBase = 1.0
	}

	textOffset := 3.5 * fontFactor * spec.ScaleFactor

	x := m.GraphX
	y := m.GraphY
	w := m.GraphWidth
	h := m.GraphHeight

	f := spec.ScaleFactor
	ff := fontFactor
	s := m.StrokeWidthBase

	// legend for x-axis
	legendXStep := float64(spec.LegendXStep)
	for w*legendXStep/spec.LegendXMax < m.FontSize*2.5 {
		legendXStep *= 2
	}
	m.PathLegend.MoveTo(x-0.5*s, y+h+0.5*s)
	m.PathLegend.LineTo(x-0.5*s+w, y+h+0.5*s)
	for i := 0.0; i <= spec.LegendXMax; i += legendXStep {
		frac := i / spec.LegendXMax
		pos := x - 0.5*s + math.Round(w*frac)
		m.PathLegend.MoveTo(pos, y+h+math.Round(2*f)+0.5*s)
		m.PathLegend.LineTo(pos, y+h+math.Round(1*f)+0.5*s)
		text := fmt.Sprintf(spec.LegendXFormat, i*spec.LegendXFactor)
		m.LabelsX = append(m.LabelsX, label{X: pos, Y: y + h + (2+8*ff)*f + textOffset, Text: text})
	}

	// legend for y-axis
	legendYLabelStep := spec.LegendYLabelStep
	for h*float64(legendYLabelStep)/(spec.LegendYTop-spec.LegendYBottom) < m.FontSize {
		legendYLabelStep *= 2
	}
	if math.Max(math.Abs(float64(spec.LegendYLabelStart)), math.Abs(float64(spec.LegendYLabelEnd))) >= 100 {
		m.LabelsYTransform.Translate(x-(5+5.5*ff)*f, 0)
		m.LabelsYTransform.Scale(0.7, 1)
		m.LabelsYTransform.Translate((5+5.5*ff)*f-x, 0)
	}
	m.PathLegend.MoveTo(x-0.5*s, y+0.5*s)
	m.PathLegend.LineTo(x-0.5*s, y+h+0.5*s)
	for i := spec.LegendYLabelStart + legendYLabelStep/2; i <= spec.LegendYLabelEnd; i += legendYLabelStep {
		frac := (float64(i) - spec.LegendYBottom) / (spec.LegendYTop - spec.LegendYBottom)
		pos := y + h + 0.5*s - math.Round(h*frac)
		m.PathLegend.MoveTo(x-math.Round(2*f)-0.5*s, pos)
		m.PathLegend.LineTo(x-math.Round(1*f)-0.5*s, pos)
	}
	for i := spec.LegendYLabelStart; i <= spec.LegendYLabelEnd; i += legendYLabelStep {
		frac := (float64(i) - spec.LegendYBottom) / (spec.LegendYTop - spec.LegendYBottom)
		pos := y + h + 0.5*s - math.Round(h*frac)
		m.PathLegend.MoveTo(x-math.Round(4*f)-0.5*s, pos)
		m.PathLegend.LineTo(x-math.Round(1*f)-0.5*s, pos)
		if frac > 0.01 {
			m.PathGrid.MoveTo(x+0.5*s, pos)
			m.PathGrid.LineTo(x+w-0.5*s, pos)
		}
		text := fmt.Sprintf("%d", i)
		m.LabelsY = append(m.LabelsY, label{X: x - (5+5.5*ff)*f, Y: pos + textOffset, Text: text})
	}

	return m
}

func setBandsData(m *baseModel, data models.Bins, useColor bool) {
	if useColor {
		m.ColorBandsDownstream = colorDownstream
		m.ColorBandsUpstream = colorUpstream
	} else {
		m.ColorBandsDownstream = m.ColorNeutralFill
		m.ColorBandsUpstream = m.ColorNeutralFill
	}
	m.ColorBandsDownstream.A = 0.075
	m.ColorBandsUpstream.A = 0.075

	m.ColorBandsStroke = m.ColorNeutralStroke
	m.ColorBandsStroke.A = 0.1

	s := m.StrokeWidthBase

	type bandWithPath struct {
		models.Band
		PathFill *path
	}

	bands := make([]bandWithPath, 0, len(data.Bands.Downstream)+len(data.Bands.Upstream))
	for _, b := range data.Bands.Downstream {
		bands = append(bands, bandWithPath{Band: b, PathFill: &m.PathBandsDownstream})
	}
	for _, b := range data.Bands.Upstream {
		bands = append(bands, bandWithPath{Band: b, PathFill: &m.PathBandsUpstream})
	}
	sort.Slice(bands, func(i, j int) bool {
		return bands[i].Start < bands[j].Start
	})

	top := m.GraphY
	bottom := m.GraphY + m.GraphHeight
	scaleX := m.GraphWidth / float64(data.Mode.BinCount())

	if len(bands) > 0 {
		band := bands[0]
		start := m.GraphX + math.Floor((float64(band.Start)+0.5)*scaleX)

		band.PathFill.MoveTo(start, bottom)
		band.PathFill.LineTo(start, top)

		m.PathBandsStroke.MoveTo(start+0.5*s, bottom-0.5*s)
		m.PathBandsStroke.LineTo(start+0.5*s, top+0.5*s)
	}

	for i := 1; i < len(bands); i++ {
		band1 := bands[i-1]
		band2 := bands[i]

		end := m.GraphX + math.Ceil((float64(band1.End)+0.5)*scaleX)
		start := m.GraphX + math.Floor((float64(band2.Start)+0.5)*scaleX)

		if start-end <= 1*s {
			center := float64(band2.Start+band1.End) / 2
			pos := m.GraphX + math.Floor((center+0.5)*scaleX) + 0.5*s
			end = pos
			start = pos

			m.PathBandsStroke.MoveTo(pos, bottom-0.5*s)
			m.PathBandsStroke.LineTo(pos, top+0.5*s)
		} else {
			m.PathBandsStroke.MoveTo(end-0.5*s, bottom-0.5*s)
			m.PathBandsStroke.LineTo(end-0.5*s, top+0.5*s)

			m.PathBandsStroke.MoveTo(start+0.5*s, bottom-0.5*s)
			m.PathBandsStroke.LineTo(start+0.5*s, top+0.5*s)
		}

		band1.PathFill.LineTo(end, top)
		band1.PathFill.LineTo(end, bottom)
		band1.PathFill.Close()

		band2.PathFill.MoveTo(start, bottom)
		band2.PathFill.LineTo(start, top)
	}

	if len(bands) > 0 {
		band := bands[len(bands)-1]
		end := m.GraphX + math.Ceil((float64(band.End)+0.5)*scaleX)

		band.PathFill.LineTo(end, top)
		band.PathFill.LineTo(end, bottom)
		band.PathFill.Close()

		m.PathBandsStroke.MoveTo(end-0.5*s, bottom-0.5*s)
		m.PathBandsStroke.LineTo(end-0.5*s, top+0.5*s)
	}
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

func buildPilotTonesPath(p *path, tones []int, height float64) {
	for _, tone := range tones {
		pos := float64(tone) + 0.5

		p.MoveTo(pos, 0)
		p.LineTo(pos, height)
	}
}

func buildBitsPath(p *path, bins models.BinsBits, scaleY float64) {
	var lastValid bool
	var lastBits int8
	var lastPosY float64

	count := len(bins.Data)
	for i := 0; i < count; i++ {
		bits := bins.Data[i]
		valid := bits > 0
		changed := lastBits != bits

		posX := float64(i)
		posY := math.Ceil(float64(bits) * scaleY)

		if lastValid && !valid {
			p.LineTo(posX, lastPosY)
			p.LineTo(posX, 0)
			p.Close()
		}
		if !lastValid && valid {
			p.MoveTo(posX, 0)
		}
		if valid && changed {
			if lastValid {
				p.LineTo(posX, lastPosY)
			}
			p.LineTo(posX, posY)
			lastPosY = posY
		}

		lastValid = valid
		lastBits = bits
	}

	if lastValid {
		p.LineTo(float64(count), lastPosY)
		p.LineTo(float64(count), 0)
		p.Close()
	}
}

func DrawBitsGraph(out io.Writer, data models.Bins, params GraphParams) error {
	bins, step, _ := getLegendX(data.Mode)

	params.normalize()

	spec := graphSpec{
		Width:             params.Width,
		Height:            params.Height,
		ScaleFactor:       params.ScaleFactor,
		FontSize:          params.FontSize,
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

	setBandsData(&m.baseModel, data, false)

	m.StrokeWidthPilotTones = 1
	if scaleX < 1.5 {
		m.StrokeWidthPilotTones = 1.5 / scaleX
	}
	m.StrokeWidthPilotTones *= spec.ScaleFactor

	buildPilotTonesPath(&m.PathPilotTones, data.PilotTones, h)

	buildBitsPath(&m.PathDownstream, data.Bits.Downstream, scaleY)
	buildBitsPath(&m.PathUpstream, data.Bits.Upstream, scaleY)

	m.Transform.Translate(x, y+h)
	m.Transform.Scale(scaleX, -1)

	return writeTemplate(out, m, templateBase, templateBits)
}

func buildSNRQLNPath(p *path, bins models.BinsFloat, scaleY, offsetY, maxY, minYValid, maxYValid float64) {
	width := float64(bins.GroupSize)

	var lastValid, lastDrawn bool
	var last float64 = offsetY
	var lastPosY float64

	count := len(bins.Data)
	for i := 0; i < count; i++ {
		val := bins.Data[i]
		valid := val > offsetY && val >= minYValid && val <= maxYValid
		changed := last != val
		drawn := false

		posX := (float64(i) + 0.5) * width
		posY := (math.Min(maxY, val) - offsetY) * scaleY

		if lastValid && !valid {
			p.LineTo(posX-0.5*width, lastPosY)
			p.LineTo(posX-0.5*width, 0)
			p.Close()
		}
		if !lastValid && valid {
			p.MoveTo(posX-0.5*width, 0)
			p.LineTo(posX-0.5*width, posY)
		}
		if valid && changed {
			if lastValid {
				if !lastDrawn {
					p.LineTo(posX-width, lastPosY)
				}
				p.LineTo(posX, posY)
				drawn = true
			}
			lastPosY = posY
		}

		lastDrawn = drawn
		lastValid = valid
		last = val
	}

	if lastValid {
		posX := (float64(count) + 0.5) * width
		p.LineTo(posX-0.5*width, lastPosY)
		p.LineTo(posX-0.5*width, 0)
		p.Close()
	}
}

func buildSNRMinMaxPath(pMin *path, pMax *path, bins models.BinsFloatMinMax, scaleY, maxY, postScaleY float64) {
	width := float64(bins.GroupSize)

	var lastValidMin, lastValidMax, lastDrawnMin, lastDrawnMax bool
	var lastMin float64 = 0
	var lastMax float64 = 0
	var lastPosYMin, lastPosYMax float64

	iter := func(p *path, i int, val float64, valid bool, last, lastPosY *float64, lastValid, lastDrawn *bool) {
		changed := *last != val
		drawn := false

		posX := (float64(i) + 0.5) * width
		posY := math.Min(maxY, val)*scaleY - 0.5

		if *lastValid && !valid {
			p.LineTo(posX-0.5*width, *lastPosY*postScaleY)
		}
		if !*lastValid && valid {
			p.MoveTo(posX-0.5*width, posY*postScaleY)
			*lastPosY = posY
		}
		if valid && changed {
			if *lastValid {
				if !*lastDrawn {
					p.LineTo(posX-width, *lastPosY*postScaleY)
				}
				p.LineTo(posX, posY*postScaleY)
				drawn = true
			}
			*lastPosY = posY
		}

		*lastDrawn = drawn
		*lastValid = valid
		*last = val
	}

	count := len(bins.Min)
	for i := 0; i < count; i++ {
		min := bins.Min[i]
		max := bins.Max[i]
		valid := (min > 0 && min <= 95) || (max > 0 && max <= 95)

		iter(pMin, i, min, valid, &lastMin, &lastPosYMin, &lastValidMin, &lastDrawnMin)
		iter(pMax, i, max, valid, &lastMax, &lastPosYMax, &lastValidMax, &lastDrawnMax)
	}

	if lastValidMin {
		pMin.LineTo(float64(count*bins.GroupSize), lastPosYMin*postScaleY)
	}
	if lastValidMax {
		pMax.LineTo(float64(count*bins.GroupSize), lastPosYMax*postScaleY)
	}

}

func DrawSNRGraph(out io.Writer, data models.Bins, params GraphParams) error {
	return DrawSNRGraphWithHistory(out, data, models.BinsHistory{}, params)
}

func DrawSNRGraphWithHistory(out io.Writer, data models.Bins, history models.BinsHistory, params GraphParams) error {
	bins, step, freq := getLegendX(data.Mode)

	params.normalize()

	spec := graphSpec{
		Width:             params.Width,
		Height:            params.Height,
		ScaleFactor:       params.ScaleFactor,
		FontSize:          params.FontSize,
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

	setBandsData(&m.baseModel, data, true)

	m.Path.SetPrecision(1)

	buildSNRQLNPath(&m.Path, data.SNR.Downstream, scaleY, 0, spec.LegendYTop, -32, 95)
	buildSNRQLNPath(&m.Path, data.SNR.Upstream, scaleY, 0, spec.LegendYTop, -32, 95)

	m.PathMin.SetPrecision(1)
	m.PathMax.SetPrecision(1)

	buildSNRMinMaxPath(&m.PathMin, &m.PathMax, history.SNR.Downstream, scaleY, spec.LegendYTop, 1/scaleX)
	buildSNRMinMaxPath(&m.PathMin, &m.PathMax, history.SNR.Upstream, scaleY, spec.LegendYTop, 1/scaleX)

	m.Transform.Translate(x, y+h)
	m.Transform.Scale(scaleX, -1)

	// scaling of y by scaleX in order to simulate vector-effect="non-scaling-stroke" for non-supporting renderers
	m.TransformMinMax.Translate(x, y+h)
	m.TransformMinMax.Scale(scaleX, -scaleX)

	m.StrokeWidth = spec.ScaleFactor / scaleX

	return writeTemplate(out, m, templateBase, templateSNR)
}

func DrawQLNGraph(out io.Writer, data models.Bins, params GraphParams) error {
	bins, step, freq := getLegendX(data.Mode)

	params.normalize()

	spec := graphSpec{
		Width:             params.Width,
		Height:            params.Height,
		ScaleFactor:       params.ScaleFactor,
		FontSize:          params.FontSize,
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

	setBandsData(&m.baseModel, data, true)

	m.Path.SetPrecision(1)

	buildSNRQLNPath(&m.Path, data.QLN.Downstream, scaleY, spec.LegendYBottom, spec.LegendYTop, -150, -23)
	buildSNRQLNPath(&m.Path, data.QLN.Upstream, scaleY, spec.LegendYBottom, spec.LegendYTop, -150, -23)

	m.Transform.Translate(x, y+h)
	m.Transform.Scale(scaleX, -1)

	return writeTemplate(out, m, templateBase, templateQLN)
}

func buildHlogPath(p *path, bins models.BinsFloat, scaleY, offsetY, maxY, postScaleY float64) {
	width := float64(bins.GroupSize)

	var lastValid, lastDrawn bool
	var last float64 = -96.3
	var lastPosY float64

	count := len(bins.Data)
	for i := 0; i < count; i++ {
		hlog := bins.Data[i]
		valid := hlog >= -96.2 && hlog <= 6
		changed := last != hlog
		drawn := false

		posX := (float64(i) + 0.5) * width
		posY := math.Max(0, math.Min(maxY, hlog)-offsetY)*scaleY - 0.5

		reset := lastValid && math.Abs(hlog-last) >= 10

		if (lastValid && !valid) || reset {
			p.LineTo(posX-0.5*width, lastPosY*postScaleY)
		}
		if (!lastValid && valid) || reset {
			p.MoveTo(posX-0.5*width, posY*postScaleY)
			lastPosY = posY
		}
		if valid && changed {
			if lastValid && !reset {
				if !lastDrawn {
					p.LineTo(posX-width, lastPosY*postScaleY)
				}
				p.LineTo(posX, posY*postScaleY)
				drawn = true
			}
			lastPosY = posY
		}

		lastDrawn = drawn
		lastValid = valid
		last = hlog
	}

	if lastValid {
		p.LineTo(float64(count*bins.GroupSize), lastPosY*postScaleY)
	}

}

func DrawHlogGraph(out io.Writer, data models.Bins, params GraphParams) error {
	bins, step, freq := getLegendX(data.Mode)

	params.normalize()

	spec := graphSpec{
		Width:             params.Width,
		Height:            params.Height,
		ScaleFactor:       params.ScaleFactor,
		FontSize:          params.FontSize,
		ColorBackground:   params.ColorBackground,
		ColorForeground:   params.ColorForeground,
		LegendXMax:        float64(bins),
		LegendXStep:       step,
		LegendXFactor:     freq / 1000,
		LegendXFormat:     "%.1f",
		LegendYBottom:     -100,
		LegendYTop:        7,
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

	setBandsData(&m.baseModel, data, true)

	m.Path.SetPrecision(1)

	buildHlogPath(&m.Path, data.Hlog.Downstream, scaleY, spec.LegendYBottom, spec.LegendYTop, 1/scaleX)
	buildHlogPath(&m.Path, data.Hlog.Upstream, scaleY, spec.LegendYBottom, spec.LegendYTop, 1/scaleX)

	// scaling of y by scaleX in order to simulate vector-effect="non-scaling-stroke" for non-supporting renderers
	m.Transform.Translate(x, y+h)
	m.Transform.Scale(scaleX, -scaleX)

	m.StrokeWidth = spec.ScaleFactor / scaleX

	return writeTemplate(out, m, templateBase, templateHlog)
}
