// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package httpdigest

import (
	"strings"
)

var quoteEscaper = strings.NewReplacer(`\`, `\\`, `"`, `\"`)

func parseParams(data string) map[string]string {
	values := make(map[string]string)

	for {
		indexSeparator := strings.IndexByte(data, '=')
		if indexSeparator == -1 {
			break
		}

		key := strings.ToLower(strings.TrimSpace(data[:indexSeparator]))
		data = data[indexSeparator+1:]

		if len(data) == 0 {
			break
		}

		if data[0] == '"' {

			values[key] = parseQuotedString(&data)

			indexComma := strings.IndexByte(data, ',')
			if indexComma == -1 {
				break
			}
			data = data[indexComma+1:]

		} else {

			indexComma := strings.IndexByte(data, ',')
			if indexComma == -1 {
				values[key] = data
				break
			}

			values[key] = data[:indexComma]

			data = data[indexComma+1:]

		}
	}

	return values
}

func parseQuotedString(data *string) string {
	if (*data)[0] != '"' {
		return ""
	}
	*data = (*data)[1:]

	var b strings.Builder

	escaped := false
	i := 0
	for ; i < len(*data); i++ {
		char := (*data)[i]
		if char == '\\' {
			escaped = true
			continue
		} else if char == '"' && !escaped {
			break
		}
		b.WriteByte(char)
		escaped = false
	}
	(*data) = (*data)[i+1:]

	return b.String()
}

func unq(str string) string {
	return quoteEscaper.Replace(str)
}

func quoteString(str string) string {
	return `"` + unq(str) + `"`
}
