// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package snmp

import (
	"errors"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"
)

type Value struct {
	OID  string
	Type byte
	Val  interface{}
}

func (v Value) ValBytes() ([]byte, error) {
	switch x := v.Val.(type) {
	case []byte:
		return x, nil
	case string:
		if v.Type == byte(gosnmp.IPAddress) {
			b := net.ParseIP(x)
			if b != nil {
				return b, nil
			}
			return nil, errors.New("invalid IP address")
		}
		return []byte(x), nil
	}
	return nil, fmt.Errorf("cannot convert type %T to bytes", v.Val)
}

func (v Value) ValString() (string, error) {
	switch x := v.Val.(type) {
	case []byte:
		return string(x), nil
	case string:
		return x, nil
	}
	return "", fmt.Errorf("cannot convert type %T to string", v.Val)
}

func (v Value) ValUint64() (uint64, error) {
	switch x := v.Val.(type) {
	case uint:
		return uint64(x), nil
	case uint32:
		return uint64(x), nil
	case uint64:
		return x, nil
	case int:
		if x >= 0 {
			return uint64(x), nil
		}
		return 0, errors.New("cannot convert negative value to uint64")
	case []byte:
		if len(x) <= 8 {
			var out uint64
			for _, b := range x {
				out = (out << 8) | uint64(b)
			}
			return out, nil
		}
		return 0, errors.New("too many bytes to convert to uint64")
	}
	return 0, fmt.Errorf("cannot convert type %T to uint64", v.Val)
}

func (v Value) ValInt64() (int64, error) {
	switch x := v.Val.(type) {
	case int:
		return int64(x), nil
	case uint32:
		return int64(x), nil
	case uint:
		if uint64(x) < uint64(math.MaxInt64) {
			return int64(x), nil
		}
		return 0, errors.New("value to large for int64")
	case uint64:
		if x <= uint64(math.MaxInt64) {
			return int64(x), nil
		}
		return 0, errors.New("value to large for int64")
	}
	return 0, fmt.Errorf("cannot convert type %T to int64", v.Val)
}

func (v Value) ValFloat64() (float64, error) {
	switch x := v.Val.(type) {
	case float32:
		return float64(x), nil
	case float64:
		return x, nil
	case int:
		return float64(x), nil
	case uint:
		return float64(x), nil
	case uint32:
		return float64(x), nil
	case uint64:
		return float64(x), nil
	case []byte:
		return strconv.ParseFloat(string(x), 64)
	case string:
		return strconv.ParseFloat(x, 64)
	}
	return 0, fmt.Errorf("cannot convert type %T to float64", v.Val)
}

func (v Value) ValDuration() (time.Duration, error) {
	switch x := v.Val.(type) {
	case uint32:
		if v.Type == byte(gosnmp.TimeTicks) {
			return time.Duration(x) * 10 * time.Millisecond, nil
		}
		return 0, errors.New("cannot convert non-TimeTicks value to duration")
	}
	return 0, fmt.Errorf("cannot convert type %T to duration", v.Val)
}

type Values struct {
	l []Value
	m map[string]*Value
}

func (v *Values) init() {
	v.l = make([]Value, 0)
	v.m = make(map[string]*Value)
}

func (v *Values) add(val Value) {
	v.l = append(v.l, val)
	v.m[val.OID] = &v.l[len(v.l)-1]
}

func (v *Values) Get(oid string) *Value {
	return v.m[oid]
}

func (v *Values) GetBytes(oid string) ([]byte, error) {
	if val := v.Get(oid); val != nil {
		return val.ValBytes()
	}
	return nil, errors.New("cannot find value")
}

func (v *Values) GetString(oid string) (string, error) {
	if val := v.Get(oid); val != nil {
		return val.ValString()
	}
	return "", errors.New("cannot find value")
}

func (v *Values) GetUint64(oid string) (uint64, error) {
	if val := v.Get(oid); val != nil {
		return val.ValUint64()
	}
	return 0, errors.New("cannot find value")
}

func (v *Values) GetInt64(oid string) (int64, error) {
	if val := v.Get(oid); val != nil {
		return val.ValInt64()
	}
	return 0, errors.New("cannot find value")
}

func (v *Values) GetFloat64(oid string) (float64, error) {
	if val := v.Get(oid); val != nil {
		return val.ValFloat64()
	}
	return 0, errors.New("cannot find value")
}

func (v *Values) Walk(callback func(Value)) {
	for _, val := range v.l {
		callback(val)
	}
}

func (v *Values) String() string {
	var b strings.Builder

	v.Walk(func(val Value) {
		typeStr := gosnmp.Asn1BER(val.Type).String()
		fmt.Fprintf(&b, "%s %s ", val.OID, typeStr)

		switch x := val.Val.(type) {
		case []byte, string:
			fmt.Fprintf(&b, "%x\n", x)
		case int, uint, uint32, uint64:
			fmt.Fprintf(&b, "%d\n", x)
		case float32, float64:
			fmt.Fprintf(&b, "%f\n", x)
		default:
			fmt.Fprintf(&b, "?\n")
		}
	})

	return b.String()
}
