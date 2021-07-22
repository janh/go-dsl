// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package graphs

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type path struct {
	sb          strings.Builder
	roundFactor float64
	openX       float64
	openY       float64
	lastX       float64
	lastY       float64
	lastX2      float64
	lastY2      float64
}

func (p *path) SetPrecision(digits uint) {
	p.roundFactor = math.Pow10(int(digits))
}

func (p *path) roundCoord(val float64) float64 {
	if p.roundFactor == 0 {
		p.roundFactor = 100000
	}
	val = math.Round(val*p.roundFactor) / p.roundFactor
	return val
}

func (p *path) roundCoords(x, y float64) (float64, float64) {
	x = p.roundCoord(x)
	y = p.roundCoord(y)
	return x, y
}

func (p *path) formatCoord(val float64) (valStr string) {
	valStr = strconv.FormatFloat(val, 'f', -1, 64)
	if valStr[0] == '0' && len(valStr) > 1 {
		valStr = valStr[1:]
	} else if valStr[0] == '-' && valStr[1] == '0' && len(valStr) > 2 {
		valStr = "-" + valStr[2:]
	}
	return
}

func (p *path) formatCoords(x, y float64) (xStr, yStr string) {
	xStr = p.formatCoord(x)
	yStr = p.formatCoord(y)
	return
}

func (p *path) buildCommand(cmd string, coords ...string) string {
	out := cmd
	var lastC string
	for i, c := range coords {
		if i == 0 || c[0] == '-' || (c[0] == '.' && strings.Contains(lastC, ".")) {
			out += c
		} else {
			out += " " + c
		}
		lastC = c
	}
	return out
}

func (p *path) MoveTo(x, y float64) {
	x, y = p.roundCoords(x, y)
	xStr, yStr := p.formatCoords(x, y)
	cmd := p.buildCommand("M", xStr, yStr) // move absolute

	dx, dy := p.roundCoords(x-p.lastX, y-p.lastY)
	dxStr, dyStr := p.formatCoords(dx, dy)
	cmdRelative := p.buildCommand("m", dxStr, dyStr) // move relative
	if len(cmdRelative) < len(cmd) {
		cmd = cmdRelative
	}

	fmt.Fprint(&p.sb, cmd)

	p.openX = x
	p.openY = y
	p.lastX = x
	p.lastY = y
	p.lastX2 = x
	p.lastY2 = y
}

func (p *path) LineTo(x, y float64) {
	x, y = p.roundCoords(x, y)
	xStr, yStr := p.formatCoords(x, y)
	cmd := p.buildCommand("L", xStr, yStr) // line absolute

	dx, dy := p.roundCoords(x-p.lastX, y-p.lastY)
	dxStr, dyStr := p.formatCoords(dx, dy)
	if dxStr == "0" && dyStr == "0" {
		return
	} else if dyStr == "0" {
		if len(xStr) <= len(dxStr) {
			cmd = "H" + xStr // horizontal line absolute
		} else {
			cmd = "h" + dxStr // horizontal line relative
		}
	} else if dxStr == "0" {
		if len(yStr) <= len(dyStr) {
			cmd = "V" + yStr // vertical line absolute
		} else {
			cmd = "v" + dyStr // vertical line relative
		}
	} else {
		cmdRelative := p.buildCommand("l", dxStr, dyStr) // line relative
		if len(cmdRelative) < len(cmd) {
			cmd = cmdRelative
		}
	}

	fmt.Fprint(&p.sb, cmd)

	p.lastX = x
	p.lastY = y
	p.lastX2 = x
	p.lastY2 = y
}

func (p *path) BezierCurveTo(x1, y1, x2, y2, x, y float64) {
	x1, y1 = p.roundCoords(x1, y1)
	x1Str, y1Str := p.formatCoords(x1, y1)
	x2, y2 = p.roundCoords(x2, y2)
	x2Str, y2Str := p.formatCoords(x2, y2)
	x, y = p.roundCoords(x, y)
	xStr, yStr := p.formatCoords(x, y)
	cmd := p.buildCommand("C", x1Str, y1Str, x2Str, y2Str, xStr, yStr) // absolute curveto

	smoothCheckLastX2, smoothCheckLastY2 := p.formatCoords(p.lastX-p.lastX2, p.lastY-p.lastY2)
	smoothCheckX1, smoothCheckY1 := p.formatCoords(x1-p.lastX, y1-p.lastY)

	dx1, dy1 := p.roundCoords(x1-p.lastX, y1-p.lastY)
	dx1Str, dy1Str := p.formatCoords(dx1, dy1)
	dx2, dy2 := p.roundCoords(x2-p.lastX, y2-p.lastY)
	dx2Str, dy2Str := p.formatCoords(dx2, dy2)
	dx, dy := p.roundCoords(x-p.lastX, y-p.lastY)
	dxStr, dyStr := p.formatCoords(dx, dy)
	if dxStr == "0" && dyStr == "0" {
		return
	} else if smoothCheckLastX2 == smoothCheckX1 && smoothCheckLastY2 == smoothCheckY1 {
		cmd = p.buildCommand("S", x2Str, y2Str, xStr, yStr)              // absolute smooth curveto
		cmdRelative := p.buildCommand("s", dx2Str, dy2Str, dxStr, dyStr) // relative smooth curveto
		if len(cmdRelative) < len(cmd) {
			cmd = cmdRelative
		}
	} else {
		cmdRelative := p.buildCommand("c", dx1Str, dy1Str, dx2Str, dy2Str, dxStr, dyStr) // relative curveto
		if len(cmdRelative) < len(cmd) {
			cmd = cmdRelative
		}
	}

	fmt.Fprint(&p.sb, cmd)

	p.lastX = x
	p.lastY = y
	p.lastX2 = x2
	p.lastY2 = y2
}

func (p *path) AddPath(otherPath path) {
	fmt.Fprint(&p.sb, otherPath.String())

	p.openX = otherPath.openX
	p.openY = otherPath.openY
	p.lastX = otherPath.lastX
	p.lastY = otherPath.lastY
	p.lastX2 = otherPath.lastX2
	p.lastY2 = otherPath.lastY2
}

func (p *path) Close() {
	fmt.Fprint(&p.sb, "z")

	p.lastX = p.openX
	p.lastY = p.openY
}

func (p *path) IsEmpty() bool {
	return p.sb.Len() == 0
}

func (p path) String() string {
	return p.sb.String()
}
