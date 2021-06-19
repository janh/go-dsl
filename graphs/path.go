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

type Path struct {
	sb          strings.Builder
	roundFactor float64
	openX       float64
	openY       float64
	lastX       float64
	lastY       float64
}

func (p *Path) SetPrecision(digits uint) {
	p.roundFactor = math.Pow10(int(digits))
}

func (p *Path) roundCoord(val float64) float64 {
	if p.roundFactor == 0 {
		p.roundFactor = 100000
	}
	val = math.Round(val*p.roundFactor) / p.roundFactor
	return val
}

func (p *Path) roundCoords(x, y float64) (float64, float64) {
	x = p.roundCoord(x)
	y = p.roundCoord(y)
	return x, y
}

func (p *Path) formatCoord(val float64) (valStr string) {
	valStr = strconv.FormatFloat(val, 'f', -1, 64)
	if valStr[0] == '0' && len(valStr) > 1 {
		valStr = valStr[1:]
	} else if valStr[0] == '-' && valStr[1] == '0' && len(valStr) > 2 {
		valStr = "-" + valStr[2:]
	}
	return
}

func (p *Path) formatCoords(x, y float64) (xStr, yStr string) {
	xStr = p.formatCoord(x)
	yStr = p.formatCoord(y)
	return
}

func (p *Path) buildCommand(cmd, xStr, yStr string) string {
	if yStr[0] == '-' || (yStr[0] == '.' && strings.Contains(xStr, ".")) {
		return cmd + xStr + yStr
	}
	return cmd + xStr + " " + yStr
}

func (p *Path) MoveTo(x, y float64) {
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
}

func (p *Path) LineTo(x, y float64) {
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
}

func (p *Path) AddPath(otherPath Path) {
	fmt.Fprint(&p.sb, otherPath.String())

	p.openX = otherPath.openX
	p.openY = otherPath.openY
	p.lastX = otherPath.lastX
	p.lastY = otherPath.lastY
}

func (p *Path) Close() {
	fmt.Fprint(&p.sb, "z")

	p.lastX = p.openX
	p.lastY = p.openY
}

func (p *Path) IsEmpty() bool {
	return p.sb.Len() == 0
}

func (p *Path) String() string {
	return p.sb.String()
}
