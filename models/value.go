// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package models

import (
	"fmt"
	"math"
	"strconv"
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

func (v *IntValue) UnmarshalJSON(data []byte) error {
	val, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		v.Valid = false
		return err
	}

	v.Int = val
	v.Valid = true

	return nil
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

func (v *FloatValue) UnmarshalJSON(data []byte) error {
	val, err := strconv.ParseFloat(string(data), 64)
	if err != nil {
		v.Valid = false
		return err
	}

	v.Float = val
	v.Valid = true

	return nil
}

type BoolValue struct {
	Valid bool
	Bool  bool
}

func (v BoolValue) String() string {
	return v.Value()
}

func (v BoolValue) Value() string {
	if v.Valid {
		if v.Bool {
			return "on"
		} else {
			return "off"
		}
	}
	return "-"
}

func (v BoolValue) Unit() string {
	return ""
}

func (v *BoolValue) UnmarshalJSON(data []byte) error {
	val, err := strconv.ParseBool(string(data))
	if err != nil {
		v.Valid = false
		return err
	}

	v.Bool = val
	v.Valid = true

	return nil
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

type ValueMilliseconds struct {
	FloatValue
}

func (v ValueMilliseconds) String() string {
	return v.Value() + " " + v.Unit()
}

func (v ValueMilliseconds) Value() string {
	if v.Valid {
		if math.Abs(v.Float-math.Round(v.Float)) <= 0.005 {
			return fmt.Sprintf("%.0f", v.Float)
		}
		return fmt.Sprintf("%.2f", v.Float)
	}
	return "-"
}

func (v ValueMilliseconds) Unit() string {
	return "ms"
}

type ValueSymbols struct {
	FloatValue
}

func (v ValueSymbols) String() string {
	return v.FloatValue.String() + " " + v.Unit()
}

func (v ValueSymbols) Unit() string {
	return "symbols"
}

type VectoringValue struct {
	Valid bool
	State VectoringState
}

func (v VectoringValue) String() string {
	return v.Value()
}

func (v VectoringValue) Value() string {
	if v.Valid {
		return v.State.String()
	}
	return "-"
}

func (v VectoringValue) Unit() string {
	return ""
}
