// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package draytek

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"3e8.eu/go/dsl"
)

type webSession struct {
	host    string
	client  *http.Client
	authStr string
}

func newWebSession(host, username string, passwordCallback dsl.PasswordCallback, tlsSkipVerify bool) (*webSession, error) {
	s := webSession{}

	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "http://" + host
	}
	if host[len(host)-1] == '/' {
		host = host[:len(host)-1]
	}
	if strings.Count(host, "/") != 2 {
		return nil, errors.New("invalid host")
	}
	s.host = host

	err := s.createHTTPClient(tlsSkipVerify)
	if err != nil {
		return nil, err
	}

	err = s.generateAuthStr()
	if err != nil {
		return nil, err
	}

	password, err := passwordCallback()
	if err != nil {
		return nil, &dsl.AuthenticationError{Err: err}
	}

	err = s.login(username, password)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (s *webSession) createHTTPClient(tlsSkipVerify bool) error {
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: tlsSkipVerify},
	}

	s.client = &http.Client{
		Jar:       cookieJar,
		Timeout:   10 * time.Second,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return nil
}

func (s *webSession) generateAuthStr() error {
	characters := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	randMax := big.NewInt(int64(len(characters)))

	bytes := make([]byte, 15)

	for i := range bytes {
		index, err := rand.Int(rand.Reader, randMax)
		if err != nil {
			return err
		}

		bytes[i] = characters[index.Int64()]
	}

	s.authStr = string(bytes)

	return nil
}

func (s *webSession) login(username, password string) error {
	data := url.Values{}
	data.Add("aa", base64.StdEncoding.EncodeToString([]byte(username)))
	data.Add("ab", base64.StdEncoding.EncodeToString([]byte(password)))
	data.Add("sFormAuthStr", s.authStr)

	resp, err := s.client.PostForm(s.host+"/cgi-bin/wlogin.cgi", data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 302 {
		return fmt.Errorf("received unexpected status code %d for login", resp.StatusCode)
	}

	location := resp.Header.Get("Location")

	if location == "/weblogin.htm" {
		err = errors.New("authentication failed")
		return &dsl.AuthenticationError{Err: err}
	}

	if location != "/" {
		return fmt.Errorf("received unexpected location on login: %s", location)
	}

	return nil
}

func (s *webSession) Execute(command string) (output string, err error) {
	data := url.Values{}
	data.Add("sFormAuthStr", s.authStr)
	data.Add("fid", "310")
	data.Add("cmd", command)

	resp, err := s.client.PostForm(s.host+"/cgi-bin/v2x00.cgi", data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("command request failed with status %d", resp.StatusCode)

		if resp.StatusCode == 302 {
			return "", &dsl.ConnectionError{Err: err}
		}

		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	return string(body), err
}

func (s *webSession) close() {
	params := url.Values{}
	params.Add("sFormAuthStr", s.authStr)

	resp, err := s.client.Get(s.host + "/cgi-bin/wlogout.cgi?" + params.Encode())
	if err != nil {
		return
	}

	resp.Body.Close()
}
