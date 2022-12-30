// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package graphs

import (
	"fmt"
	"io"
	"math"
	"time"

	"3e8.eu/go/dsl/models"
)

type errorsGraphItem struct {
	data  []models.IntValue
	color Color
}

func formatLegendXLabelErrors(val, step, start, end int) string {
	switch {
	case step%(60*24) == 0:
		return fmt.Sprintf("%d\u202Fd", val/(60*24))
	case step%(60*12) == 0:
		return fmt.Sprintf("%.1f\u202Fd", float64(val)/(60*24))
	case step%60 == 0:
		return fmt.Sprintf("%d\u202Fh", val/60)
	case step < 30:
		return fmt.Sprintf("%d\u202Fmin", val)
	case step < 60:
		return fmt.Sprintf("%.1f\u202Fh", float64(val)/60)
	default:
		return "?"
	}
}

func formatLegendYLabelErrors(val, step, start, end int) string {
	if val == 0 {
		return "0"
	}

	if end >= 1_000_000 {
		if val%1_000_000 == 0 {
			return fmt.Sprintf("%d\u202FM", val/1_000_000)
		} else if val%100_000 == 0 {
			return fmt.Sprintf("%.1f\u202FM", float64(val)/1_000_000.0)
		} else {
			return fmt.Sprintf("%.2f\u202FM", float64(val)/1_000_000.0)
		}
	}

	if end >= 1_000 {
		if val%1_000 == 0 {
			return fmt.Sprintf("%d\u202Fk", val/1_000)
		} else if val%100 == 0 {
			return fmt.Sprintf("%.1f\u202Fk", float64(val)/1_000.0)
		} else {
			return fmt.Sprintf("%.2f\u202Fk", float64(val)/1_000.0)
		}
	}

	return fmt.Sprintf("%d", val)
}

func getErrorsHistoryLegendX(data models.ErrorsHistory) (max float64, steps []int) {
	var totalDuration time.Duration
	if data.PeriodCount != 0 {
		totalDuration = time.Duration(data.PeriodCount) * data.PeriodLength
	} else {
		totalDuration = 24 * time.Hour
	}

	max = totalDuration.Seconds() / time.Minute.Seconds()

	steps = []int{1, 2, 5, 10, 20, 30, 1 * 60, 2 * 60, 3 * 60, 6 * 60, 12 * 60, 24 * 60}

	return
}

func getErrorsHistoryLegendY(items []errorsGraphItem) (max float64, end int, steps []int) {
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

	steps = []int{1, 2, 4, 10, 20, 50}

	for i := 2; i < 8; i++ {
		factor := math.Pow10(i)
		for _, j := range []float64{1, 2.5, 5} {
			val := j * factor
			steps = append(steps, int(val))
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

	maxX, stepsX := getErrorsHistoryLegendX(data)
	maxY, endY, stepsY := getErrorsHistoryLegendY(items)

	params.normalize()

	spec := graphSpec{
		Width:                  params.Width,
		Height:                 params.Height,
		ScaleFactor:            params.ScaleFactor,
		FontSize:               params.FontSize,
		ColorBackground:        params.ColorBackground,
		ColorForeground:        params.ColorForeground,
		LegendXMin:             maxX,
		LegendXMax:             0,
		LegendXLabelStart:      0,
		LegendXLabelEnd:        int(maxX),
		LegendXLabelSteps:      stepsX,
		LegendXLabelFormatFunc: formatLegendXLabelErrors,
		LegendXLabelDigits:     5.5,
		LegendYBottom:          0,
		LegendYTop:             maxY,
		LegendYLabelStart:      0,
		LegendYLabelEnd:        endY,
		LegendYLabelSteps:      stepsY,
		LegendYLabelFormatFunc: formatLegendYLabelErrors,
		LegendYLabelDigits:     5.0,
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

func DrawDownstreamErrorSecondsGraph(out io.Writer, data models.ErrorsHistory, params GraphParams) error {
	return drawErrorsGraph(out, data, params,
		[]errorsGraphItem{
			{data: data.DownstreamESCount, color: colorBlue},
			{data: data.DownstreamSESCount, color: colorRed},
		})
}

func DrawUpstreamErrorSecondsGraph(out io.Writer, data models.ErrorsHistory, params GraphParams) error {
	return drawErrorsGraph(out, data, params,
		[]errorsGraphItem{
			{data: data.UpstreamESCount, color: colorBlue},
			{data: data.UpstreamSESCount, color: colorRed},
		})
}
