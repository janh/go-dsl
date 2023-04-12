// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package graphs

import (
	"math"
	"sort"
)

var (
	colorGreen = Color{96, 192, 0, .75}
	colorBlue  = Color{0, 127, 255, .75}
	colorRed   = Color{204, 94, 82, .75}
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

func determineLegendStep(specSteps []int, valueRange float64, maxStepCount float64) (step int) {
	sort.Slice(specSteps, func(i, j int) bool {
		return math.Abs(float64(specSteps[i])) < math.Abs(float64(specSteps[j]))
	})

	minStep := valueRange / maxStepCount

	for _, step = range specSteps {
		if math.Abs(float64(step)) >= minStep {
			break
		}
	}

	for math.Abs(float64(step)) < minStep {
		step *= 2
	}

	return
}

func findNextStep(start, step int) int {
	if step > 0 && start >= 0 {
		return ((start + step - 1) / step) * step
	} else if step < 0 && start < 0 {
		return ((start + step + 1) / step) * step
	} else {
		return (start / step) * step
	}
}

func getBaseModel(spec graphSpec) baseModel {
	m := baseModel{}

	m.ScaledWidth = float64(spec.Width) / spec.ScaleFactor
	m.ScaledHeight = float64(spec.Height) / spec.ScaleFactor

	m.Width = float64(spec.Width)
	m.Height = float64(spec.Height)

	var fontFactor float64
	if spec.FontSize == 0 {
		factor := math.Min(m.ScaledWidth/560, m.ScaledHeight/114)
		fontFactor = math.Min(math.Max(1.0, factor), 1.35)
		m.FontSize = 10.5 * fontFactor * spec.ScaleFactor
	} else {
		fontFactor = spec.FontSize / 10.5
		m.FontSize = spec.FontSize * spec.ScaleFactor
	}

	digitWidth := 6.1
	digitHeight := 10.5

	// 23.0 for default factors and 3.75 digits
	labelYWidth := (spec.LegendYLabelDigits*digitWidth*fontFactor + 0.125) * spec.ScaleFactor
	// 13.0 for default factors and 4.0 digits
	labelXMarginWidth := (0.5*spec.LegendXLabelDigits*digitWidth*fontFactor + 0.8) * spec.ScaleFactor

	m.GraphX = math.Round(math.Max(
		labelYWidth+(6.0*fontFactor+5.0)*spec.ScaleFactor,
		labelXMarginWidth+1.0*spec.ScaleFactor))
	m.GraphY = math.Round(4.0 * fontFactor * spec.ScaleFactor)
	m.GraphWidth = m.Width - m.GraphX - math.Round(labelXMarginWidth+1.0*spec.ScaleFactor)
	m.GraphHeight = m.Height - m.GraphY - math.Round((14.0*fontFactor+5.0)*spec.ScaleFactor)

	m.ColorBackground = spec.ColorBackground
	m.ColorText = spec.ColorForeground

	m.ColorGraph, m.ColorGrid, m.ColorNeutralFill, m.ColorNeutralStroke =
		getGraphColors(spec.ColorBackground, spec.ColorForeground)

	m.ColorMinStroke = colorBlue
	m.ColorMaxStroke = colorGreen

	m.ColorUpstream = colorGreen
	m.ColorDownstream = colorBlue

	m.ColorPilotTones = colorRed

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

	loopSteps := func(start, end, step int, callback func(int)) {
		count := ((end - start) / step) + 1
		for i := 0; i < count; i++ {
			callback(start + i*step)
		}
	}

	// legend for x-axis
	maxStepXCount := w / ((spec.LegendXLabelDigits + 1) * digitWidth * ff * f)
	if maxStepXCount > 16 {
		maxStepXCount = 16 + (maxStepXCount-16)*0.4
	}
	legendXValueRange := math.Abs(spec.LegendXMax - spec.LegendXMin)
	legendXLabelStep := determineLegendStep(spec.LegendXLabelSteps, legendXValueRange, maxStepXCount)
	legendXLabelStart := findNextStep(spec.LegendXLabelStart, legendXLabelStep)
	m.PathLegend.MoveTo(x-0.5*s, y+h+0.5*s)
	m.PathLegend.LineTo(x-0.5*s+w, y+h+0.5*s)
	loopSteps(legendXLabelStart, spec.LegendXLabelEnd, legendXLabelStep, func(i int) {
		frac := (float64(i) - spec.LegendXMin) / (spec.LegendXMax - spec.LegendXMin)
		pos := x - 0.5*s + math.Round(w*frac)
		m.PathLegend.MoveTo(pos, y+h+math.Round(2*f)+0.5*s)
		m.PathLegend.LineTo(pos, y+h+math.Round(1*f)+0.5*s)
		text := spec.LegendXLabelFormatFunc(i, legendXLabelStep, legendXLabelStart, spec.LegendXLabelEnd)
		m.LabelsX = append(m.LabelsX, label{X: pos, Y: y + h + (2+8*ff)*f + textOffset, Text: text})
	})

	// legend for y-axis
	maxStepYCount := h / (digitHeight * ff * f)
	if maxStepYCount > 7.5 {
		maxStepYCount = 7.5 + (maxStepYCount-7.5)*0.2
	}
	legendYValueRange := math.Abs(spec.LegendYTop - spec.LegendYBottom)
	legendYLabelStep := determineLegendStep(spec.LegendYLabelSteps, legendYValueRange, maxStepYCount)
	legendYLabelStart := findNextStep(spec.LegendYLabelStart, legendYLabelStep)
	m.PathLegend.MoveTo(x-0.5*s, y+0.5*s)
	m.PathLegend.LineTo(x-0.5*s, y+h+0.5*s)
	loopSteps(legendYLabelStart+legendYLabelStep/2, spec.LegendYLabelEnd, legendYLabelStep, func(i int) {
		frac := (float64(i) - spec.LegendYBottom) / (spec.LegendYTop - spec.LegendYBottom)
		pos := y + h + 0.5*s - math.Round(h*frac)
		m.PathLegend.MoveTo(x-math.Round(2*f)-0.5*s, pos)
		m.PathLegend.LineTo(x-math.Round(1*f)-0.5*s, pos)
	})
	loopSteps(legendYLabelStart, spec.LegendYLabelEnd, legendYLabelStep, func(i int) {
		frac := (float64(i) - spec.LegendYBottom) / (spec.LegendYTop - spec.LegendYBottom)
		pos := y + h + 0.5*s - math.Round(h*frac)
		m.PathLegend.MoveTo(x-math.Round(4*f)-0.5*s, pos)
		m.PathLegend.LineTo(x-math.Round(1*f)-0.5*s, pos)
		if frac > 0.01 {
			m.PathGrid.MoveTo(x+0.5*s, pos)
			m.PathGrid.LineTo(x+w-0.5*s, pos)
		}
		text := spec.LegendYLabelFormatFunc(i, legendYLabelStep, legendYLabelStart, spec.LegendYLabelEnd)
		m.LabelsY = append(m.LabelsY, label{X: x - (5+5.5*ff)*f, Y: pos + textOffset, Text: text})
	})

	return m
}
