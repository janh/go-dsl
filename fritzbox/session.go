// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package fritzbox

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"

	"golang.org/x/crypto/pbkdf2"

	"3e8.eu/go/dsl"
	"3e8.eu/go/dsl/internal/httpdigest"
)

type session struct {
	host          string
	username      string
	password      string
	client        *http.Client
	clientSupport *http.Client
	sid           string
}

type sessionInfo struct {
	XMLName   xml.Name          `xml:"SessionInfo"`
	SID       string            `xml:"SID"`
	Challenge string            `xml:"Challenge"`
	BlockTime int               `xml:"BlockTime"`
	Users     []sessionInfoUser `xml:"Users>User"`
}

type sessionInfoUser struct {
	XMLName xml.Name `xml:"User"`
	Name    string   `xml:",chardata"`
}

var regexpDefaultUser = regexp.MustCompile(`fritz[0-9]{4}`)

func newSession(host, username string, passwordCallback dsl.PasswordCallback, tlsSkipVerify bool) (*session, error) {
	s := session{}
	s.username = username

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

	s.createHTTPClient(tlsSkipVerify)

	enforceAuthentication := strings.HasPrefix(host, "https://")

	// load session info
	sessionInfoRaw, err := s.get("/login_sid.lua?version=2")
	if err != nil {
		return nil, err
	}

	var info sessionInfo
	err = xml.Unmarshal(sessionInfoRaw, &info)
	if err != nil {
		return nil, err
	}

	// check for valid session, then no login required
	if info.SID != "0000000000000000" && !enforceAuthentication {
		s.sid = info.SID
		return &s, nil
	}

	// check if login is blocked at the moment
	if info.BlockTime > 0 {
		err := fmt.Errorf("login blocked for %d seconds", info.BlockTime)
		return nil, &dsl.AuthenticationError{Err: err, WaitTime: time.Duration(info.BlockTime) * time.Second}
	}

	// support login without username on firmware >= 7.25
	if s.username == "" {
		if len(info.Users) == 1 {
			s.username = info.Users[0].Name
		} else if len(info.Users) > 1 {
			for _, u := range info.Users {
				if regexpDefaultUser.MatchString(u.Name) {
					s.username = u.Name
					break
				}
			}
		}
	}

	// get password
	if passwordCallback != nil {
		s.password, err = passwordCallback()
		if err != nil {
			return nil, &dsl.AuthenticationError{Err: err}
		}
	}
	if enforceAuthentication && s.password == "" {
		err := errors.New("password authentication is required when TLS is used")
		return nil, &dsl.AuthenticationError{Err: err}
	}

	// calculate challenge response and try to get session
	response := getChallengeResponse(info.Challenge, s.password)

	data := url.Values{}
	data.Add("username", s.username)
	data.Add("response", response)

	sessionInfoRaw, err = s.postForm("/login_sid.lua?version=2", data)
	if err != nil {
		return nil, err
	}

	err = xml.Unmarshal(sessionInfoRaw, &info)
	if err != nil {
		return nil, err
	}

	// check if session is valid
	if info.SID == "0000000000000000" {
		err := errors.New("authentication failed")
		return nil, &dsl.AuthenticationError{Err: err}
	}

	s.sid = info.SID

	return &s, nil
}

func (s *session) createHTTPClient(tlsSkipVerify bool) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: tlsSkipVerify},
	}
	s.client = &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	s.clientSupport = &http.Client{
		Timeout:   60 * time.Second,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func getChallengeResponse(challenge, password string) string {
	if strings.HasPrefix(challenge, "2$") {

		challengeSplit := strings.Split(challenge, "$")
		if len(challengeSplit) != 5 {
			return ""
		}

		iter1, _ := strconv.Atoi(challengeSplit[1])
		salt1, _ := hex.DecodeString(challengeSplit[2])
		iter2, _ := strconv.Atoi(challengeSplit[3])
		salt2, _ := hex.DecodeString(challengeSplit[4])

		hash1 := pbkdf2.Key([]byte(password), salt1, iter1, 32, sha256.New)
		hash2 := pbkdf2.Key(hash1, salt2, iter2, 32, sha256.New)
		response := fmt.Sprintf("%x$%x", salt2, hash2)

		return response

	} else {

		password = strings.Map(func(r rune) rune {
			if r > 255 {
				return '.'
			}
			return r
		}, password)

		data := challenge + "-" + password
		dataUTF16 := utf16.Encode([]rune(data))
		dataUTF16LE := make([]byte, 2*len(dataUTF16))

		bytes := make([]byte, 2, 2)
		for i, val := range dataUTF16 {
			binary.LittleEndian.PutUint16(bytes, val)
			dataUTF16LE[i*2] = bytes[0]
			dataUTF16LE[i*2+1] = bytes[1]
		}

		hash := md5.Sum(dataUTF16LE)
		response := fmt.Sprintf("%s-%x", challenge, hash)

		return response

	}
}

func (s *session) get(path string) ([]byte, error) {
	resp, err := s.client.Get(s.host + path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("request for %s failed with status %d", path, resp.StatusCode)

		if resp.StatusCode == 303 {
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
		err = fmt.Errorf("request for %s failed with status %d", path, resp.StatusCode)

		if resp.StatusCode == 303 {
			return nil, &dsl.ConnectionError{Err: err}
		}
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	return body, err
}

func (s *session) loadGet(path string, data url.Values) (string, error) {
	data.Add("sid", s.sid)
	path = path + "?" + data.Encode()
	body, err := s.get(path)
	return string(body), err
}

func (s *session) loadPost(path string, data url.Values) (string, error) {
	data.Add("sid", s.sid)
	body, err := s.postForm(path, data)
	return string(body), err
}

func (s *session) loadSupportData() (string, error) {
	// this needs to use multipart/form-data and the order of the fields is important
	var body bytes.Buffer
	mpart := multipart.NewWriter(&body)
	mpart.WriteField("sid", s.sid)
	mpart.WriteField("DiagnosisData", "")

	req, err := http.NewRequest(http.MethodPost, s.host+"/cgi-bin/firmwarecfg", &body)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", mpart.FormDataContentType())

	resp, err := s.clientSupport.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("request for support data failed with status %d", resp.StatusCode)

		if resp.StatusCode == 303 {
			return "", &dsl.ConnectionError{Err: err}
		}
		return "", err
	}

	var b strings.Builder
	var foundBeginSection bool

	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := scanner.Text()

		if !foundBeginSection && strings.HasPrefix(line, "#### BEGIN SECTION DSLManager_port") {
			foundBeginSection = true
		}

		if foundBeginSection {
			fmt.Fprintln(&b, line)
		}

		if foundBeginSection && strings.HasPrefix(line, "#### END SECTION DSLManager_port") {
			resp.Body.Close()
			break
		}
	}

	return b.String(), scanner.Err()
}

func (s *session) getHostWithoutPort() string {
	bracketIndex := strings.LastIndexByte(s.host, ']')
	if bracketIndex != -1 {
		return s.host[:bracketIndex+1]
	}

	colonIndex := strings.LastIndexByte(s.host, ':')
	if colonIndex > 5 {
		return s.host[:colonIndex]
	}

	return s.host
}

func (s *session) loadTR064(path, serviceType, action string) (string, error) {
	var url string
	if strings.HasPrefix(s.host, "https://") {
		url = s.host + "/tr064" + path
	} else {
		url = s.getHostWithoutPort() + ":49000" + path
	}

	soapAction := serviceType + "#" + action
	soapRequest := fmt.Sprintf(
		`<?xml version="1.0"?>`+
			`<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">`+
			`<s:Body><u:%[2]s xmlns:u="%[1]s"></u:%[2]s></s:Body>`+
			`</s:Envelope>`,
		serviceType, action)

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(soapRequest))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", soapAction)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		user := s.username
		if user == "" {
			// authentication seems to fail when username is empty string
			user = " "
		}

		authorization, err := httpdigest.GetAuthorization(resp, user, s.password)
		if err != nil {
			return "", err
		}

		req, err = http.NewRequest(http.MethodPost, url, strings.NewReader(soapRequest))
		if err != nil {
			return "", err
		}

		req.Header.Set("Content-Type", "text/xml; charset=utf-8")
		req.Header.Set("SOAPAction", soapAction)
		req.Header.Set("Authorization", authorization)

		resp, err = s.client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("request for %s failed with status %d", soapAction, resp.StatusCode)

		if resp.StatusCode == 401 {
			return "", &dsl.ConnectionError{Err: err}
		}
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	return string(body), err
}

func (s *session) close() {
	if s.sid == "" {
		return
	}

	data := url.Values{}
	data.Add("logout", "")
	data.Add("sid", s.sid)

	s.postForm("/login_sid.lua?version=2", data)
	s.sid = ""
}
