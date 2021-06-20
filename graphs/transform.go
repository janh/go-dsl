// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package graphs

import (
	"math"
	"strconv"
	"strings"
)

type Transform struct {
	funcs []string
}

func (t *Transform) formatCoord(val float64) (valStr string) {
	val = math.Round(val*100000) / 100000
	valStr = strconv.FormatFloat(val, 'f', -1, 64)
	if valStr[0] == '0' && len(valStr) > 1 {
		valStr = valStr[1:]
	}
	return
}

func (t *Transform) addFunction(funcName string, coords ...float64) {
	f := funcName + "("
	for i, val := range coords {
		if i != 0 {
			f += " "
		}
		f += t.formatCoord(val)
	}
	f += ")"

	t.funcs = append(t.funcs, f)
}

func (t *Transform) Scale(x, y float64) {
	if x == y {
		t.addFunction("scale", x)
	} else {
		t.addFunction("scale", x, y)
	}
}

func (t *Transform) Translate(x, y float64) {
	if y != 0 {
		t.addFunction("translate", x, y)
	} else {
		t.addFunction("translate", x)
	}
}

func (t Transform) String() string {
	var sb strings.Builder

	count := len(t.funcs)
	for i := 0; i < count; i++ {
		if i != 0 {
			sb.WriteRune(' ')
		}
		sb.WriteString(t.funcs[count-i-1])
	}

	return sb.String()
}
