// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package httpdigest

import (
	"crypto"
	_ "crypto/md5"
	"crypto/rand"
	_ "crypto/sha256"
	_ "crypto/sha512"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

func sliceContains(items []string, str string) bool {
	for _, item := range items {
		if item == str {
			return true
		}
	}
	return false
}

func generateRandomValue(length int) (string, error) {
	val := make([]byte, length)

	_, err := rand.Read(val)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", val), nil
}

func hashHelper(hash crypto.Hash, data string) string {
	h := hash.New()
	h.Write([]byte(data))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func parseWWWAuthenticationParams(params string) (realm, nonce, opaque, algorithm string, qop []string) {
	data := parseParams(params)

	realm = data["realm"]
	nonce = data["nonce"]
	opaque = data["opaque"]
	algorithm = data["algorithm"]

	qopStr := data["qop"]
	if qopStr != "" {
		qop = strings.Split(qopStr, ",")
	}

	return
}

func GetAuthorization(resp *http.Response, username, password string) (string, error) {
	wwwAuthenticate := resp.Header.Get("WWW-Authenticate")
	if !strings.HasPrefix(wwwAuthenticate, "Digest") {
		return "", errors.New("httpdigest: unexpected WWW-Authenticate prefix")
	}

	realm, nonce, opaque, algorithm, qop := parseWWWAuthenticationParams(wwwAuthenticate[6:])

	if !sliceContains(qop, "auth") {
		return "", errors.New("httpdigest: unsupported quality of protection")
	}

	var hashType crypto.Hash
	switch algorithm {
	case "SHA-512":
		hashType = crypto.SHA512_256
	case "SHA-256":
		hashType = crypto.SHA256
	case "MD5":
		hashType = crypto.MD5
	default:
		return "", errors.New("httpdigest: unsupported algorithm")
	}

	authorization := "Digest realm=" + quoteString(realm)
	authorization += ",uri=" + quoteString(resp.Request.URL.Path)
	authorization += ",nonce=" + quoteString(nonce)
	authorization += ",opaque=" + quoteString(opaque)
	authorization += ",algorithm=" + algorithm
	authorization += ",qop=auth"

	cnonce, err := generateRandomValue(8)
	if err != nil {
		return "", err
	}
	authorization += ",cnonce=" + quoteString(cnonce)

	nc := "00000001"
	authorization += ",nc=" + nc

	ha1 := hashHelper(hashType, unq(username)+":"+unq(realm)+":"+password)
	ha2 := hashHelper(hashType, resp.Request.Method+":"+resp.Request.URL.Path)
	response := hashHelper(hashType, ha1+":"+unq(nonce)+":"+nc+":"+unq(cnonce)+":auth:"+ha2)

	authorization += ",response=" + quoteString(response)
	authorization += ",username=" + quoteString(username)

	return authorization, nil
}
