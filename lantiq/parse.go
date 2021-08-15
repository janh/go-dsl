// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lantiq

import (
	"strings"
	"unicode"
)

func parseValues(data string) (values map[string]string) {
	values = make(map[string]string)

	for {
		separatorIndex := strings.IndexRune(data, '=')
		if separatorIndex == -1 {
			break
		}

		key := strings.TrimSpace(data[0:separatorIndex])
		data = data[separatorIndex+1:]

		var endIndex, nextIndex int
		if len(data) >= 1 && data[0] == '"' {
			data = data[1:]
			endIndex = strings.IndexRune(data, '"')
		} else if len(data) >= 1 && data[0] == '(' {
			var bracketCount int
			endIndex = -1
			for i := 0; i < len(data); i++ {
				if data[i] == '(' {
					bracketCount++
				} else if data[i] == ')' {
					bracketCount--
				}
				if bracketCount == 0 {
					endIndex = i + 1
					break
				}
			}
		} else {
			endIndex = strings.IndexFunc(data, unicode.IsSpace)
		}

		if endIndex == -1 || endIndex == len(data) {
			endIndex = len(data)
			nextIndex = len(data)
		} else {
			nextIndex = endIndex + 1
		}

		value := strings.TrimSpace(data[0:endIndex])
		values[key] = value

		data = data[nextIndex:]
	}

	if nReturn, ok := values["nReturn"]; !ok || nReturn != "0" {
		return nil
	}

	return
}
