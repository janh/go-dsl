// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package exec

import (
	"errors"
	"regexp"

	"golang.org/x/crypto/ssh"
)

var regexpCommandNotFound = regexp.MustCompile("(.*)sh: (.+): (command |)not found")

func IsCommandNotFound(output string, err error) bool {
	// Check SSH exit status first
	if err != nil {
		var exitErr *ssh.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.ExitStatus() == 127 {
				return true
			}
		}
	}

	truncatedOutput := output
	if len(truncatedOutput) > 100 {
		truncatedOutput = truncatedOutput[0:100]
	}

	// Check if output matches typical error of common shells
	if regexpCommandNotFound.MatchString(truncatedOutput) {
		return true
	}

	return false
}
