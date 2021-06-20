// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package graphs

import (
	"fmt"
	"io"
	"math"

	"github.com/ajstarks/svgo/float"

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

func drawGraphBackground(s *svg.SVG, spec graphSpec) (x, y, w, h float64) {
	x = 28.0
	y = 4.0
	w = float64(spec.Width) - 42.0
	h = float64(spec.Height) - 23.0

	colorBackground := spec.ColorBackground
	colorText := spec.ColorForeground
	colorGraph, colorGrid := getGraphColors(spec.ColorBackground, spec.ColorForeground)

	textOffset := 3.5

	var pathLegend, pathGrid Path

	// background
	s.Rect(0, 0, float64(spec.Width), float64(spec.Height), "fill:"+colorBackground.String())
	s.Rect(x, y, w, h, "fill:"+colorGraph.String())

	// legend for x-axis
	s.Gstyle("text-anchor:middle;font-family:Arial,Helvetica,sans-serif;font-size:10.5px;fill:" + colorText.String())
	pathLegend.MoveTo(x-0.5, y+h+0.5)
	pathLegend.LineTo(x-0.5+w, y+h+0.5)
	for i := 0.0; i <= spec.LegendXMax; i += float64(spec.LegendXStep) {
		frac := i / spec.LegendXMax
		pos := x - 0.5 + math.Round(w*frac)
		pathLegend.MoveTo(pos, y+h+2.5)
		pathLegend.LineTo(pos, y+h+1.5)
		text := fmt.Sprintf(spec.LegendXFormat, i*spec.LegendXFactor)
		s.Text(pos, y+h+10.5+textOffset, text)
	}
	s.Gend()

	// legend for y-axis
	s.Gstyle("text-anchor:end;font-family:Arial,Helvetica,sans-serif;font-size:10.5px;fill:" + colorText.String())
	needsTransform := math.Max(math.Abs(float64(spec.LegendYLabelStart)), math.Abs(float64(spec.LegendYLabelEnd))) >= 100
	if needsTransform {
		var transform Transform
		transform.Translate(10.5-x, 0)
		transform.Scale(0.7, 1)
		transform.Translate(x-10.5, 0)
		s.Gtransform(transform.String())
	}
	pathLegend.MoveTo(x-0.5, y+0.5)
	pathLegend.LineTo(x-0.5, y+h+0.5)
	for i := spec.LegendYLabelStart + spec.LegendYLabelStep/2; i <= spec.LegendYLabelEnd; i += spec.LegendYLabelStep {
		frac := (float64(i) - spec.LegendYBottom) / (spec.LegendYTop - spec.LegendYBottom)
		pos := y + h + 0.5 - math.Round(h*frac)
		pathLegend.MoveTo(x-2.5, pos)
		pathLegend.LineTo(x-1.5, pos)
	}
	for i := spec.LegendYLabelStart; i <= spec.LegendYLabelEnd; i += spec.LegendYLabelStep {
		frac := (float64(i) - spec.LegendYBottom) / (spec.LegendYTop - spec.LegendYBottom)
		pos := y + h + 0.5 - math.Round(h*frac)
		pathLegend.MoveTo(x-4.5, pos)
		pathLegend.LineTo(x-1.5, pos)
		if frac > 0.01 {
			pathGrid.MoveTo(x+0.5, pos)
			pathGrid.LineTo(x+w-0.5, pos)
		}
		text := fmt.Sprintf("%d", i)
		s.Text(x-10.5, pos+textOffset, text)
	}
	if needsTransform {
		s.Gend()
	}
	s.Gend()

	s.Path(pathLegend.String(), "fill:none;stroke-linecap:square;stroke:"+colorText.String())
	s.Path(pathGrid.String(), "fill:none;stroke-linecap:square;stroke:"+colorGrid.String())

	return
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

func DrawBitsGraph(out io.Writer, data models.Bins, params GraphParams) {
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

	s := svg.New(out)
	s.Start(float64(params.Width), float64(params.Height))

	x, y, w, h := drawGraphBackground(s, spec)

	scaleX := w / spec.LegendXMax
	scaleY := h / spec.LegendYTop

	var pathNone, pathUpstream, pathDownstream Path

	var lastBits int8
	var lastPosY float64
	var lastPath *Path

	for i := 0; i < bins; i++ {
		bin := data.Bins[i]
		bits := bin.Bits
		if bits < 0 {
			bits = 0
		}

		var path *Path
		switch bin.Type {
		case models.BinTypeNone:
			path = &pathNone
		case models.BinTypeUpstream:
			path = &pathUpstream
		case models.BinTypeDownstream:
			path = &pathDownstream
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

	var transform Transform
	transform.Scale(scaleX, 1)
	transform.Translate(x, y)
	s.Gtransform(transform.String())
	s.Path(pathNone.String(), "fill:"+colorNeutral.String())
	s.Path(pathUpstream.String(), "fill:"+colorUpstream.String())
	s.Path(pathDownstream.String(), "fill:"+colorDownstream.String())
	s.Gend()

	s.End()
}

func DrawSNRGraph(out io.Writer, data models.Bins, params GraphParams) {
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

	s := svg.New(out)
	s.Start(float64(params.Width), float64(params.Height))

	x, y, w, h := drawGraphBackground(s, spec)

	scaleX := w / spec.LegendXMax
	scaleY := h / spec.LegendYTop

	var path Path
	path.SetPrecision(1)

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
			path.LineTo(posX, lastPosY)
			path.LineTo(posX, h)
			path.Close()
		}
		if last == 0 && snr > 0 {
			path.MoveTo(posX, h)
		}
		if snr > 0 && last != snr {
			posY := h - math.Min(h, snr*scaleY)
			if last != 0 {
				path.LineTo(posX, lastPosY)
			}
			path.LineTo(posX, posY)
			lastPosY = posY
		}

		last = snr
	}

	if last > 0 {
		path.LineTo(spec.LegendXMax, lastPosY)
		path.LineTo(spec.LegendXMax, h)
		path.Close()
	}

	var transform Transform
	transform.Scale(scaleX, 1)
	transform.Translate(x, y)
	s.Gtransform(transform.String())
	s.Path(path.String(), "fill:"+colorNeutral.String())
	s.Gend()

	s.End()
}

func DrawQLNGraph(out io.Writer, data models.Bins, params GraphParams) {
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
		LegendYTop:        -70,
		LegendYLabelStart: -160,
		LegendYLabelEnd:   -80,
		LegendYLabelStep:  20,
	}

	s := svg.New(out)
	s.Start(float64(params.Width), float64(params.Height))

	x, y, w, h := drawGraphBackground(s, spec)

	scaleX := w / spec.LegendXMax
	scaleY := h / (spec.LegendYTop - spec.LegendYBottom)
	offsetY := spec.LegendYBottom

	var path Path
	path.SetPrecision(1)

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
			path.LineTo(posX, lastPosY)
			path.LineTo(posX, h)
			path.Close()
		}
		if last <= offsetY && qln > offsetY {
			path.MoveTo(posX, h)
		}
		if qln > offsetY && last != qln {
			posY := h - math.Max(0, math.Min(h, (qln-offsetY)*scaleY))
			if last > offsetY {
				path.LineTo(posX, lastPosY)
			}
			path.LineTo(posX, posY)
			lastPosY = posY
		}

		last = qln
	}

	if last > offsetY {
		path.LineTo(spec.LegendXMax, lastPosY)
		path.LineTo(spec.LegendXMax, h)
		path.Close()
	}

	var transform Transform
	transform.Scale(scaleX, 1)
	transform.Translate(x, y)
	s.Gtransform(transform.String())
	s.Path(path.String(), "fill:"+colorNeutral.String())
	s.Gend()

	s.End()
}

func DrawHlogGraph(out io.Writer, data models.Bins, params GraphParams) {
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

	s := svg.New(out)
	s.Start(float64(params.Width), float64(params.Height))

	x, y, w, h := drawGraphBackground(s, spec)

	scaleX := w / spec.LegendXMax
	scaleY := h / (spec.LegendYTop - spec.LegendYBottom)
	offsetY := spec.LegendYBottom

	var path Path
	path.SetPrecision(1)

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
			path.LineTo(posX-0.5, lastPosY/scaleX)
		}
		if (last >= 0 && hlog < 0) || reset {
			path.MoveTo(posX-0.5, posY/scaleX)
			lastPosY = posY
		}
		if hlog < 0 && last != hlog {
			if lastValid && !reset {
				path.LineTo(posX-1, lastPosY/scaleX)
				path.LineTo(posX, posY/scaleX)
			}
			lastPosY = posY
		}

		lastValid = hlog < 0
		last = hlog
	}

	if last < 0 {
		path.LineTo(spec.LegendXMax, lastPosY/scaleX)
		path.Close()
	}

	// scaling of y by scaleX in order to simulate vector-effect="non-scaling-stroke" for non-supporting renderers
	var transform Transform
	transform.Scale(scaleX, scaleX)
	transform.Translate(x, y)
	s.Gtransform(transform.String())
	s.Path(path.String(), fmt.Sprintf("fill:none;stroke-width:%f;stroke-linecap:butt;stroke:", 1.25/scaleX)+colorNeutral.String())
	s.Gend()

	s.End()
}
