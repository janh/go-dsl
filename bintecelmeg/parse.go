// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package bintecelmeg

import (
	"bufio"
	"regexp"
	"strings"
	"unicode"
)

var regexpFilterCharacters = regexp.MustCompile(`[^a-zA-Z0-9]+`)

var regexpStatisticsItem = regexp.MustCompile(`^([a-zA-Z ]+)\s+([\-0-9].*)$`)

func normalizeLabel(label string) string {
	return strings.ToLower(regexpFilterCharacters.ReplaceAllString(label, ""))
}

func isSectionHeader(line string) bool {
	if len(line) < 1 || line[len(line)-1] != ':' {
		return false
	}
	line = line[0 : len(line)-1]

	for _, r := range line {
		if !unicode.IsUpper(r) && !unicode.IsSpace(r) {
			return false
		}
	}

	return true
}

func parseSections(data string) map[string][]string {
	sections := make(map[string][]string)
	currentSection := ""

	scanner := bufio.NewScanner(strings.NewReader(data))

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if isSectionHeader(line) {
			currentSection = normalizeLabel(line)
		} else if currentSection != "" {
			sections[currentSection] = append(sections[currentSection], line)
		}
	}

	return sections
}

func parseKeyValueItems(lines []string) map[string]string {
	result := make(map[string]string)

	for _, line := range lines {
		split := strings.SplitN(line, ": ", 2)
		if len(split) != 2 {
			continue
		}

		key := normalizeLabel(split[0])
		val := split[1]

		result[key] = val
	}

	return result
}

func parseStatisticsItems(lines []string) map[string]string {
	result := make(map[string]string)

	currentBinSection := ""

	for _, line := range lines {
		if len(line) > 1 && line[len(line)-1] == ':' {
			currentBinSection = normalizeLabel(line[0 : len(line)-1])
			continue
		}

		if currentBinSection != "" && strings.ContainsRune(line, ':') {
			if _, ok := result[currentBinSection]; ok {
				result[currentBinSection] += "\n" + line
			} else {
				result[currentBinSection] = line
			}
		} else {
			currentBinSection = ""

			if matches := regexpStatisticsItem.FindStringSubmatch(line); len(matches) > 0 {
				key := normalizeLabel(matches[1])
				val := matches[2]

				result[key] = val
			}
		}
	}

	return result
}
