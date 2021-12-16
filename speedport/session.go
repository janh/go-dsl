// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package speedport

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"3e8.eu/go/dsl"
)

var regexpChallenge = regexp.MustCompile(`challenge\s?=\s?"([0-9A-Za-z]+)"`)

type responseVar struct {
	Type  string `json:"vartype"`
	ID    string `json:"varid"`
	Value string `json:"varvalue"`
}

type session struct {
	host          string
	client        *http.Client
	clientNoRedir *http.Client
	challengev    string
}

func newSession(host string, passwordCallback dsl.PasswordCallback, tlsSkipVerify bool) (*session, error) {
	s := session{}

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

	err = s.loadChallenge()
	if err != nil {
		return nil, err
	}

	password := passwordCallback()

	err = s.login(password)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (s *session) hashPassword(password string) string {
	data := []byte(s.challengev + ":" + password)
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

func (s *session) loadChallenge() error {
	index, err := s.get("/html/login/index.html")
	if err != nil {
		return err
	}

	challengeMatch := regexpChallenge.FindSubmatch(index)
	if challengeMatch == nil {
		return errors.New("no challenge found")
	}

	s.challengev = string(challengeMatch[1])

	return nil
}

func (s *session) login(password string) error {
	data := url.Values{}
	data.Set("csrf_token", "nulltoken")
	data.Set("password", s.hashPassword(password))
	data.Set("challengev", s.challengev)

	response, err := s.postForm("/data/Login.json", data)
	if err != nil {
		return err
	}

	values, err := s.parseResponse(response)
	if err != nil {
		return err
	}

	if values["login"].Value == "failed" {
		var err error
		var waitTime time.Duration

		if loginLocked, ok := values["login_locked"]; ok {
			loginLockedInt, _ := strconv.Atoi(loginLocked.Value)
			waitTime = time.Duration(loginLockedInt) * time.Second
			err = fmt.Errorf("authentication failed, login locked for %d seconds", loginLockedInt)
		} else {
			err = errors.New("authentication failed")
		}

		return &dsl.AuthenticationError{Err: err, WaitTime: waitTime}
	}

	if values["login"].Value != "success" {
		return errors.New("unexpected response to login request")
	}

	return nil
}

func (s *session) createHTTPClient(tlsSkipVerify bool) error {
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

func (s *session) loadData(path string) ([]byte, map[string]responseVar, error) {
	response, err := s.get(path)
	if err != nil {

		return nil, nil, err
	}

	values, err := s.parseResponse(response)
	if err != nil {
		return nil, nil, err
	}

	return response, values, nil
}

func (s *session) parseResponse(data []byte) (map[string]responseVar, error) {
	var varList []responseVar
	err := json.Unmarshal(data, &varList)
	if err != nil {
		return nil, err
	}

	values := make(map[string]responseVar)
	for _, v := range varList {
		values[v.ID] = v
	}

	return values, nil
}

func (s *session) get(path string) ([]byte, error) {
	resp, err := s.client.Get(s.host + path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err := fmt.Errorf("request for %s failed with status %d", path, resp.StatusCode)

		if resp.StatusCode == 302 {
			return nil, &dsl.ConnectionError{Err: err}
		}
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	return body, err
}

func (s *session) postForm(path string, data url.Values) ([]byte, error) {
	resp, err := s.client.PostForm(s.host+path, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err := fmt.Errorf("request for %s failed with status %d", path, resp.StatusCode)

		if resp.StatusCode == 302 {
			return nil, &dsl.ConnectionError{Err: err}
		}
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	return body, err
}

func (s *session) close() {
	data := url.Values{}
	data.Set("logout", "byby")

	s.postForm("/data/Login.json", data)
}
