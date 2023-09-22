// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package javascript

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"

	"3e8.eu/go/dsl/models"
)

// List encoding format:
// command (single character) followed by number without sign
// r: repeat previous value N times
// P: absolute value, positive
// Q: absolute value, positive, only fractional part
// N: absolute value, negative
// O: absolute value, negative, only fractional part
// p: relative value, positive
// q: relative value, positive, only fractional part
// n: relative value, negative
// o: relative value, negative, only fractional part
// e: invalid value (null)
// t: boolean true
// f: boolean false

func formatListValFloat64(prefixPositive, prefixNegative byte, val int64) (valStr string) {
	var prefix byte
	if val >= 0 {
		prefix = prefixPositive
	} else {
		prefix = prefixNegative
		val = -val
	}

	num := val / 10
	frac := val % 10

	if num == 0 {
		valStr = fmt.Sprintf("%s%d", string(prefix+1), frac)
	} else if frac == 0 {
		valStr = fmt.Sprintf("%s%d", string(prefix), num)
	} else {
		valStr = fmt.Sprintf("%s%d.%d", string(prefix), num, frac)
	}

	return
}

func encodeListFloat64(list []float64) json.RawMessage {
	var buf bytes.Buffer

	buf.WriteByte('"')

	var lastVal int64
	var count int

	for _, floatVal := range list {
		val := int64(math.Round(floatVal * 10))

		if val == lastVal {
			count++
			continue
		}
		if count > 0 {
			buf.WriteByte('r')
			fmt.Fprintf(&buf, "%d", count)
			count = 0
		}

		abs := formatListValFloat64('P', 'N', val)

		diff := val - lastVal
		rel := formatListValFloat64('p', 'n', diff)

		if len(abs) <= len(rel) {
			fmt.Fprint(&buf, abs)
		} else {
			fmt.Fprint(&buf, rel)
		}

		lastVal = val
	}

	if count > 0 {
		buf.WriteByte('r')
		fmt.Fprintf(&buf, "%d", count)
	}

	buf.WriteByte('"')

	return json.RawMessage(buf.Bytes())
}

func formatListValInt8(prefixPositive, prefixNegative byte, val int8) (valStr string) {
	var prefix byte
	if val >= 0 {
		prefix = prefixPositive
	} else {
		prefix = prefixNegative
		val = -val
	}

	valStr = fmt.Sprintf("%s%d", string(prefix), val)

	return
}

func encodeListInt8(list []int8) json.RawMessage {
	var buf bytes.Buffer

	buf.WriteByte('"')

	var lastVal int8
	var count int

	for _, val := range list {
		if val == lastVal {
			count++
			continue
		}
		if count > 0 {
			buf.WriteByte('r')
			fmt.Fprintf(&buf, "%d", count)
			count = 0
		}

		abs := formatListValInt8('P', 'N', val)

		diff := val - lastVal
		rel := formatListValInt8('p', 'n', diff)

		if len(abs) <= len(rel) {
			fmt.Fprint(&buf, abs)
		} else {
			fmt.Fprint(&buf, rel)
		}

		lastVal = val
	}

	if count > 0 {
		buf.WriteByte('r')
		fmt.Fprintf(&buf, "%d", count)
	}

	buf.WriteByte('"')

	return json.RawMessage(buf.Bytes())
}

func formatListValInt64(prefixPositive, prefixNegative byte, val int64) (valStr string) {
	var prefix byte
	if val >= 0 {
		prefix = prefixPositive
	} else {
		prefix = prefixNegative
		val = -val
	}

	valStr = fmt.Sprintf("%s%d", string(prefix), val)

	return
}

func encodeListIntValue(list []models.IntValue) json.RawMessage {
	var buf bytes.Buffer

	buf.WriteByte('"')

	var lastVal = models.IntValue{Valid: true, Int: 0}
	var count int

	for _, val := range list {
		if !val.Valid {
			val.Int = 0
		}
		if val == lastVal {
			count++
			continue
		}
		if count > 0 {
			buf.WriteByte('r')
			fmt.Fprintf(&buf, "%d", count)
			count = 0
		}

		if !val.Valid {
			buf.WriteByte('e')
		} else {
			abs := formatListValInt64('P', 'N', val.Int)

			diff := val.Int - lastVal.Int
			rel := formatListValInt64('p', 'n', diff)

			if !lastVal.Valid || len(abs) <= len(rel) {
				fmt.Fprint(&buf, abs)
			} else {
				fmt.Fprint(&buf, rel)
			}
		}

		lastVal = val
	}

	if count > 0 {
		buf.WriteByte('r')
		fmt.Fprintf(&buf, "%d", count)
	}

	buf.WriteByte('"')

	return json.RawMessage(buf.Bytes())
}

func encodeListBoolValue(list []models.BoolValue) json.RawMessage {
	var buf bytes.Buffer

	buf.WriteByte('"')

	var lastVal = models.BoolValue{Valid: true, Bool: false}
	var count int

	for i, val := range list {
		if !val.Valid {
			val.Bool = false
		}
		if i != 0 && val == lastVal {
			count++
			continue
		}
		if count > 0 {
			buf.WriteByte('r')
			fmt.Fprintf(&buf, "%d", count)
			count = 0
		}

		if !val.Valid {
			buf.WriteByte('e')
		} else {
			if val.Bool {
				buf.WriteByte('t')
			} else {
				buf.WriteByte('f')
			}
		}

		lastVal = val
	}

	if count > 0 {
		buf.WriteByte('r')
		fmt.Fprintf(&buf, "%d", count)
	}

	buf.WriteByte('"')

	return json.RawMessage(buf.Bytes())
}
