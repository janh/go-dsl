// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package helpers

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// SplitHostPort is a wrapper around net.SplitHostPort which accepts inputs without a port. It also
// enforces brackets around IPv6 addresses.
func SplitHostPort(hostport string) (host string, port uint16, err error) {
	colonCount := strings.Count(hostport, ":")

	if colonCount > 0 {
		if colonCount > 1 {
			if hostport[0] != '[' {
				return "", 0, errors.New("invalid host")
			}

			if hostport[len(hostport)-1] == ']' {
				return hostport[1 : len(hostport)-2], 0, nil
			}
		}

		host, portStr, err := net.SplitHostPort(hostport)
		if err != nil {
			return "", 0, err
		}

		portUint64, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil {
			return "", 0, fmt.Errorf("invalid port: %w", err)
		}

		return host, uint16(portUint64), nil
	}

	return hostport, 0, nil
}
