// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package sagemcom

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha512"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"3e8.eu/go/dsl"
)

var regexpConfigurationSHA512 = regexp.MustCompile(`GUI_ACTIVATE_SHA512ENCODE_OPT:\s?([0-9]+)`)
var regexpConfigurationSalt = regexp.MustCompile(`GUI_PASSWORD_SALT:\s?("(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*')`)

type session struct {
	host           string
	username       string
	password       string
	client         *http.Client
	clientLogin    *http.Client
	useSHA512      bool
	salt           string
	requestCounter uint32
	sessionID      string
	serverNonce    string
}

type xmoRequestWrapper struct {
	Request xmoRequest `json:"request"`
}

type xmoRequest struct {
	ID        uint32             `json:"id"`
	SessionID string             `json:"session-id"`
	Priority  bool               `json:"priority"`
	Actions   []xmoRequestAction `json:"actions"`
	CNonce    uint32             `json:"cnonce"`
	AuthKey   string             `json:"auth-key"`
}

type xmoRequestAction struct {
	ID         uint32      `json:"id"`
	Method     string      `json:"method"`
	XPath      string      `json:"xpath,omitempty"`
	Parameters interface{} `json:"parameters,omitempty"`
}

type xmoLoginRequestParameters struct {
	User           string            `json:"user"`
	Persistent     string            `json:"persistent"`
	SessionOptions xmoSessionOptions `json:"session-options"`
}

type xmoSessionOptions struct {
	NSS             []xmoNSS           `json:"nss"`
	Language        string             `json:"language"`
	ContextFlags    xmoContextFlags    `json:"context-flags"`
	CapabilityDepth int                `json:"capability-depth"`
	CapabilityFlags xmoCapabilityFlags `json:"capability-flags"`
	TimeFormat      string             `json:"time-format"`
}

type xmoNSS struct {
	Name string `json:"name"`
	URI  string `json:"uri"`
}

type xmoContextFlags struct {
	GetContentName bool `json:"get-content-name"`
	LocalTime      bool `json:"local-time"`
}

type xmoCapabilityFlags struct {
	Name         bool `json:"name"`
	DefaultValue bool `json:"default-value"`
	Restriction  bool `json:"restriction"`
	Description  bool `json:"description"`
}

type xmoReplyWrapper struct {
	Reply xmoReply `json:"reply"`
}

type xmoReply struct {
	UID     uint32           `json:"uid"`
	ID      uint32           `json:"id"`
	Error   xmoError         `json:"error"`
	Actions []xmoReplyAction `json:"actions"`
}

type xmoReplyAction struct {
	UID       uint32        `json:"uid"`
	ID        uint32        `json:"id"`
	Error     xmoError      `json:"error"`
	Callbacks []xmoCallback `json:"callbacks"`
}

type xmoCallback struct {
	UID        uint32          `json:"uid"`
	Result     xmoError        `json:"result"`
	XPath      string          `json:"xpath"`
	Parameters json.RawMessage `json:"parameters"`
}

type xmoError struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

type xmoSessionID string

func (id *xmoSessionID) UnmarshalJSON(data []byte) (err error) {
	if len(data) > 0 && data[0] == '"' {
		var s string
		err = json.Unmarshal(data, &s)
		*id = xmoSessionID(s)
		return err
	}

	var i uint32
	err = json.Unmarshal(data, &i)
	*id = xmoSessionID(strconv.FormatUint(uint64(i), 10))
	return
}

type xmoLoginReplyParameters struct {
	ID    xmoSessionID `json:"id"`
	Nonce string       `json:"nonce"`
}

type xmoValueReplyParameters struct {
	Value json.RawMessage `json:"value"`
}

const (
	xmoInvalidSessionError = "XMO_INVALID_SESSION_ERR"
	xmoRequestNoError      = "XMO_REQUEST_NO_ERR"
	xmoNoError             = "XMO_NO_ERR"
)

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

	err := s.loadConfiguration()
	if err != nil {
		return nil, err
	}

	s.password, err = passwordCallback()
	if err != nil {
		return nil, &dsl.AuthenticationError{Err: err}
	}

	err = s.login()
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (s *session) createHTTPClient(tlsSkipVerify bool) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: tlsSkipVerify},
	}
	s.client = &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}
	s.clientLogin = &http.Client{
		Timeout:   40 * time.Second, // required due to delay of up to 32 seconds after failed login retries
		Transport: transport,
	}
}

func (s *session) unquote(str string) string {
	str = str[1 : len(str)-1]

	var escaped bool
	str = strings.Map(func(r rune) rune {
		if escaped {
			escaped = false
		} else if r == '\\' {
			escaped = true
			return -1
		}
		return r
	}, str)

	return str
}

func (s *session) loadConfiguration() error {
	guiCoreJS, err := s.get("/js/gui-core.js")
	if err != nil {
		return err
	}

	configSHA512Match := regexpConfigurationSHA512.FindSubmatch(guiCoreJS)
	if configSHA512Match != nil {
		s.useSHA512 = string(configSHA512Match[1]) == "1"
	}

	configSaltMatch := regexpConfigurationSalt.FindSubmatch(guiCoreJS)
	if configSaltMatch != nil {
		s.salt = s.unquote(string(configSaltMatch[1]))
	}

	return nil
}

func (s *session) login() error {
	actions := []xmoRequestAction{
		xmoRequestAction{
			ID:     0,
			Method: "logIn",
			Parameters: xmoLoginRequestParameters{
				User:       s.username,
				Persistent: "true",
				SessionOptions: xmoSessionOptions{
					NSS: []xmoNSS{
						xmoNSS{
							Name: "gtw",
							URI:  "http://sagemcom.com/gateway-data",
						},
					},
					Language: "ident",
					ContextFlags: xmoContextFlags{
						GetContentName: true,
						LocalTime:      true,
					},
					CapabilityDepth: 2,
					CapabilityFlags: xmoCapabilityFlags{
						Name:         true,
						DefaultValue: false,
						Restriction:  true,
						Description:  false,
					},
					TimeFormat: "ISO_8601",
				},
			},
		},
	}

	reply, err := s.doRequest(actions, true)
	if err != nil {
		return err
	}

	if len(reply.Actions) != 1 {
		return fmt.Errorf("unexpected login reply: %d actions", len(reply.Actions))
	}
	action := &reply.Actions[0]

	if len(action.Callbacks) != 1 {
		return fmt.Errorf("unexpected login reply: %d callbacks", len(action.Callbacks))
	}
	callback := &action.Callbacks[0]

	if callback.Result.Description != xmoNoError {
		err := fmt.Errorf("login failed: %s", callback.Result.Description)
		return &dsl.AuthenticationError{Err: err}
	}

	var parameters xmoLoginReplyParameters
	err = json.Unmarshal(callback.Parameters, &parameters)
	if err != nil {
		return err
	}

	s.sessionID = string(parameters.ID)
	s.serverNonce = parameters.Nonce

	return nil
}

func (s *session) loadValue(xpath string) ([]byte, error) {
	actions := []xmoRequestAction{
		xmoRequestAction{
			ID:     0,
			Method: "getValue",
			XPath:  xpath,
		},
	}

	reply, err := s.doRequest(actions, false)
	if err != nil {
		return nil, err
	}

	if reply.Error.Description != xmoRequestNoError {
		err = fmt.Errorf("loading value failed: %s", reply.Error.Description)

		if reply.Error.Description == xmoInvalidSessionError {
			return nil, &dsl.ConnectionError{Err: err}
		}
		return nil, err
	}

	if len(reply.Actions) != 1 {
		return nil, fmt.Errorf("unexpected value reply: %d actions", len(reply.Actions))
	}
	action := &reply.Actions[0]

	if action.Error.Description != xmoNoError {
		return nil, fmt.Errorf("loading value failed: %s", action.Error.Description)
	}

	if len(action.Callbacks) != 1 {
		return nil, fmt.Errorf("unexpected value reply: %d callbacks", len(action.Callbacks))
	}
	callback := &action.Callbacks[0]

	if callback.Result.Description != xmoNoError {
		return nil, fmt.Errorf("loading value failed: %s", callback.Result.Description)
	}

	var parameters xmoValueReplyParameters
	err = json.Unmarshal(callback.Parameters, &parameters)
	if err != nil {
		return nil, err
	}

	return parameters.Value, nil
}

func (s *session) doRequest(actions []xmoRequestAction, isLogin bool) (*xmoReply, error) {
	request, err := s.buildRequest(actions)
	if err != nil {
		return nil, err
	}

	data := url.Values{}
	data.Add("req", request)

	response, err := s.postForm("/cgi/json-req", data, isLogin)
	if err != nil {
		return nil, err
	}

	reply, err := s.parseReply(response)
	return reply, err
}

func (s *session) generateNonce() (uint32, error) {
	nonceBytes := make([]byte, 4)

	_, err := rand.Read(nonceBytes)
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint32(nonceBytes), nil
}

func (s *session) hash(data []byte) []byte {
	if s.useSHA512 {
		h := sha512.Sum512(data)
		return h[:]
	}

	h := md5.Sum(data)
	return h[:]
}

func (s *session) hashPassword() []byte {
	if s.salt != "" {
		return s.hash([]byte(s.password + ":" + s.salt))
	}
	return s.hash([]byte(s.password))
}

func (s *session) buildRequest(actions []xmoRequestAction) (string, error) {
	requestID := s.requestCounter
	s.requestCounter++

	nonce, err := s.generateNonce()
	if err != nil {
		return "", err
	}

	passwordHash := s.hashPassword()

	ha1Str := fmt.Sprintf("%s:%s:%x", s.username, s.serverNonce, passwordHash)
	ha1 := s.hash([]byte(ha1Str))

	authKeyStr := fmt.Sprintf("%x:%d:%d:JSON:/cgi/json-req", ha1, requestID, nonce)
	authKey := s.hash([]byte(authKeyStr))

	request := xmoRequestWrapper{
		Request: xmoRequest{
			ID:        requestID,
			SessionID: s.sessionID,
			Priority:  true,
			Actions:   actions,
			CNonce:    nonce,
			AuthKey:   fmt.Sprintf("%x", authKey),
		},
	}

	requestJSON, err := json.Marshal(request)
	return string(requestJSON), err
}

func (s *session) parseReply(data []byte) (*xmoReply, error) {
	var replyWrapper xmoReplyWrapper
	err := json.Unmarshal(data, &replyWrapper)
	return &replyWrapper.Reply, err
}

func (s *session) get(path string) ([]byte, error) {
	resp, err := s.client.Get(s.host + path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("request for %s failed with status %d", path, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	return body, err
}

func (s *session) postForm(path string, data url.Values, isLogin bool) ([]byte, error) {
	client := s.client
	if isLogin {
		client = s.clientLogin
	}

	resp, err := client.PostForm(s.host+path, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("request for %s failed with status %d", path, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	return body, err
}

func (s *session) close() {
	if s.sessionID == "" {
		return
	}

	actions := []xmoRequestAction{
		xmoRequestAction{
			ID:     0,
			Method: "logOut",
		},
	}
	s.doRequest(actions, false)

	s.sessionID = ""
}
