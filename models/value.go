// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package models

import (
	"fmt"
)

type Value interface {
	String() string
	Value() string
	Unit() string
}

type IntValue struct {
	Valid bool
	Int   int64
}

func (v IntValue) String() string {
	return v.Value()
}

func (v IntValue) Value() string {
	if v.Valid {
		return fmt.Sprintf("%d", v.Int)
	}
	return "-"
}

func (v IntValue) Unit() string {
	return ""
}

type FloatValue struct {
	Valid bool
	Float float64
}

func (v FloatValue) String() string {
	return v.Value()
}

func (v FloatValue) Value() string {
	if v.Valid {
		return fmt.Sprintf("%.1f", v.Float)
	}
	return "-"
}

func (v FloatValue) Unit() string {
	return ""
}

type ValueBandwidth struct {
	IntValue
}

func (v ValueBandwidth) String() string {
	return v.Value() + " " + v.Unit()
}

func (v ValueBandwidth) Unit() string {
	return "kbit/s"
}

type ValueDecibel struct {
	FloatValue
}

func (v ValueDecibel) String() string {
	return v.FloatValue.String() + " " + v.Unit()
}

func (v ValueDecibel) Unit() string {
	return "dB"
}

type ValuePower struct {
	FloatValue
}

func (v ValuePower) String() string {
	return v.FloatValue.String() + " " + v.Unit()
}

func (v ValuePower) Unit() string {
	return "dBm"
}
