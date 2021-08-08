// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ssh

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"strings"
)

func checkHashedHost(hash, host string) bool {
	split := strings.Split(hash, "|")
	if len(split) != 4 {
		return false
	}

	if split[0] != "" || split[1] != "1" {
		return false
	}

	saltBytes, err := base64.StdEncoding.DecodeString(split[2])
	if err != nil {
		return false
	}

	hmacBytes, err := base64.StdEncoding.DecodeString(split[3])
	if err != nil {
		return false
	}

	mac := hmac.New(sha1.New, saltBytes)
	mac.Write([]byte(host))
	hmacCalculated := mac.Sum(nil)

	return hmac.Equal(hmacBytes, hmacCalculated)
}
