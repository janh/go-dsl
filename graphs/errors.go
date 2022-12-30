// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package graphs

import (
	"io"
	"math"
	"time"

	"3e8.eu/go/dsl/models"
)

type errorsGraphItem struct {
	data  []models.IntValue
	color Color
}

func getErrorsHistoryLegendX(data models.ErrorsHistory) (max float64, step int, factor float64, format string) {
	var totalDuration time.Duration
	if data.PeriodCount != 0 {
		totalDuration = time.Duration(data.PeriodCount) * data.PeriodLength
	} else {
		totalDuration = 24 * time.Hour
	}

	var stepDuration time.Duration

	if totalDuration <= 4*time.Hour {
		minStepDuration := totalDuration / 12
		for _, val := range []int64{1, 5, 10, 20} {
			stepDuration = time.Duration(val) * time.Minute
			if minStepDuration <= stepDuration {
				break
			}
		}
		factor = 1
		format = "%.0f min"
	} else if totalDuration <= 8*time.Hour {
		stepDuration = 30 * time.Minute
		factor = 1.0 / 60
		format = "%.1f h"
	} else if totalDuration <= 4*24*time.Hour {
		minStepDuration := totalDuration / 16
		for _, val := range []int64{1, 2, 3, 6} {
			stepDuration = time.Duration(val) * time.Hour
			if minStepDuration <= stepDuration {
				break
			}
		}
		factor = 1.0 / 60
		format = "%.0f h"
	} else if totalDuration <= 8*24*time.Hour {
		stepDuration = 12 * time.Hour
		factor = 1.0 / 1440
		format = "%.1f d"
	} else {
		stepDuration = 24 * time.Hour
		factor = 1.0 / 1440
		format = "%.0f d"
	}

	step = int(stepDuration / time.Minute)
	max = totalDuration.Seconds() / time.Minute.Seconds()

	return
}

func getErrorsHistoryLegendY(items []errorsGraphItem) (max float64, end, step int) {
	var maxValue int64
	for _, item := range items {
		for _, val := range item.data {
			if val.Valid && val.Int > maxValue {
				maxValue = val.Int
			}
		}
	}

	if maxValue < 5 {
		maxValue = 5
	}

	max = 1.05 * float64(maxValue)
	end = int(max)

	minStep := max / 8

	if minStep <= 50 {
		for _, val := range []int{1, 2, 4, 10, 20, 50} {
			if minStep <= float64(val) {
				step = val
				break
			}
		}
	} else {
		factor := math.Pow10(int(math.Log10(minStep)))
		for _, i := range []float64{1, 2.5, 5, 10} {
			val := i * factor
			if minStep <= val {
				step = int(val)
				break
			}
		}
	}

	return
}

func buildErrorsPath(p *path, data []models.IntValue, scaleY, maxY, postScaleY float64) {
	var lastDrawn bool
	var last models.IntValue
	var lastPosY float64

	count := len(data)
	for i := 0; i < count; i++ {
		val := data[i]
		changed := last != val
		drawn := false

		posX := float64(i) + 0.5
		posY := math.Min(maxY, float64(val.Int))*scaleY - 0.5

		if last.Valid && !val.Valid {
			p.LineTo(posX-0.5, lastPosY*postScaleY)
		}
		if !last.Valid && val.Valid {
			p.MoveTo(posX-0.5, posY*postScaleY)
			lastPosY = posY
		}
		if val.Valid && changed {
			if last.Valid {
				if !lastDrawn {
					p.LineTo(posX-1, lastPosY*postScaleY)
				}
				p.LineTo(posX, posY*postScaleY)
				drawn = true
			}
			lastPosY = posY
		}

		lastDrawn = drawn
		last = val
	}

	if last.Valid {
		p.LineTo(float64(count), lastPosY*postScaleY)
	}
}

func drawErrorsGraph(out io.Writer, data models.ErrorsHistory, params GraphParams, items []errorsGraphItem) error {

	maxX, stepX, factorX, formatX := getErrorsHistoryLegendX(data)
	maxY, endY, stepY := getErrorsHistoryLegendY(items)

	params.normalize()

	spec := graphSpec{
		Width:              params.Width,
		Height:             params.Height,
		ScaleFactor:        params.ScaleFactor,
		FontSize:           params.FontSize,
		ColorBackground:    params.ColorBackground,
		ColorForeground:    params.ColorForeground,
		LegendXMin:         maxX,
		LegendXMax:         0,
		LegendXLabelStart:  0,
		LegendXLabelEnd:    int(maxX),
		LegendXLabelStep:   stepX,
		LegendXLabelFactor: factorX,
		LegendXLabelFormat: formatX,
		LegendYBottom:      0,
		LegendYTop:         maxY,
		LegendYLabelStart:  0,
		LegendYLabelEnd:    endY,
		LegendYLabelStep:   stepY,
	}

	m := errorsModel{}
	m.baseModel = getBaseModel(spec)

	x := m.GraphX
	y := m.GraphY
	w := m.GraphWidth
	h := m.GraphHeight

	scaleX := w / float64(data.PeriodCount)
	scaleY := h / spec.LegendYTop

	for _, item := range items {
		p := coloredPath{}

		p.Color = item.color

		p.Path.SetPrecision(1)
		buildErrorsPath(&p.Path, item.data, scaleY, spec.LegendYTop, 1/scaleX)

		m.Paths = append(m.Paths, p)
	}

	// scaling of y by scaleX in order to simulate vector-effect="non-scaling-stroke" for non-supporting renderers
	m.Transform.Translate(x, y+h)
	m.Transform.Scale(scaleX, -scaleX)

	m.StrokeWidth = spec.ScaleFactor / scaleX

	return writeTemplate(out, m, templateBase, templateErrors)
}

func DrawDownstreamRetransmissionGraph(out io.Writer, data models.ErrorsHistory, params GraphParams) error {
	return drawErrorsGraph(out, data, params,
		[]errorsGraphItem{
			{data: data.DownstreamRTXTXCount, color: colorGreen},
			{data: data.DownstreamRTXCCount, color: colorBlue},
			{data: data.DownstreamRTXUCCount, color: colorRed},
		})
}

func DrawUpstreamRetransmissionGraph(out io.Writer, data models.ErrorsHistory, params GraphParams) error {
	return drawErrorsGraph(out, data, params,
		[]errorsGraphItem{
			{data: data.UpstreamRTXTXCount, color: colorGreen},
			{data: data.UpstreamRTXCCount, color: colorBlue},
			{data: data.UpstreamRTXUCCount, color: colorRed},
		})
}

func DrawDownstreamErrorsGraph(out io.Writer, data models.ErrorsHistory, params GraphParams) error {
	return drawErrorsGraph(out, data, params,
		[]errorsGraphItem{
			{data: data.DownstreamFECCount, color: colorBlue},
			{data: data.DownstreamCRCCount, color: colorRed},
		})
}

func DrawUpstreamErrorsGraph(out io.Writer, data models.ErrorsHistory, params GraphParams) error {
	return drawErrorsGraph(out, data, params,
		[]errorsGraphItem{
			{data: data.UpstreamFECCount, color: colorBlue},
			{data: data.UpstreamCRCCount, color: colorRed},
		})
}

func DrawDownstreamErroredSecondsGraph(out io.Writer, data models.ErrorsHistory, params GraphParams) error {
	return drawErrorsGraph(out, data, params,
		[]errorsGraphItem{
			{data: data.DownstreamESCount, color: colorBlue},
			{data: data.DownstreamSESCount, color: colorRed},
		})
}

func DrawUpstreamErroredSecondsGraph(out io.Writer, data models.ErrorsHistory, params GraphParams) error {
	return drawErrorsGraph(out, data, params,
		[]errorsGraphItem{
			{data: data.UpstreamESCount, color: colorBlue},
			{data: data.UpstreamSESCount, color: colorRed},
		})
}
