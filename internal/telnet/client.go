// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package telnet

import (
	"errors"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/ziutek/telnet"

	"3e8.eu/go/dsl"
)

var regexpPort = regexp.MustCompile(`:[0-9]+$`)

// see ECMA-48, sections 5.3 and 5.4
var regexpANSIEscapeSequence = regexp.MustCompile("\x1b" + `(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])`)

type Client struct {
	config          ClientConfig
	conn            *telnet.Conn
	lastWrittenLine string
	lastPromptLine  string
}

func NewClient(config ClientConfig, host, username string, password dsl.PasswordCallback) (*Client, error) {
	c := Client{
		config: config,
	}

	err := c.connect(host, username, password)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *Client) writeLine(data string, isSensitive bool) error {
	if strings.ContainsAny(data, "\r\n") {
		return errors.New("only input without newline character is supported for Telnet")
	}

	if !isSensitive {
		c.lastWrittenLine = data
	} else {
		c.lastWrittenLine = strings.Repeat("*", len(data))
	}

	_, err := c.conn.Write([]byte(data + "\r\n"))
	return err
}

func (c *Client) readUntilPromptRaw(prompts ...string) (data, prompt string, err error) {
	dataBytes, index, err := c.conn.ReadUntilIndex(prompts...)
	if err != nil {
		return
	}

	// ANSI escape sequences may interfere with parsing
	dataBytes = regexpANSIEscapeSequence.ReplaceAll(dataBytes, nil)

	prompt = prompts[index]
	data = string(dataBytes)

	return
}

func (c *Client) readUntilPrompt(prompts ...string) (data string, promptType promptType, err error) {
	data, prompt, err := c.readUntilPromptRaw(prompts...)
	if err != nil {
		return
	}

	// When it is not clear whether the prompt actually contains a trailing space,
	// it may be omitted in the configuration. If the server sends one, it will be
	// at the start of the read data.
	if c.lastPromptLine != "" && c.lastPromptLine[len(c.lastPromptLine)-1] != ' ' && data[0] == ' ' {
		data = data[1:]
	}

	// Sometimes the entire prompt may be resent by the server instead of just echoing the input
	hasRepeatedPromptCR := strings.HasPrefix(data, "\r"+c.lastPromptLine)
	hasRepeatedPromptCRLF := c.config.ExpectRepeatedPromptCRLF && strings.HasPrefix(data, "\r\n"+c.lastPromptLine)

	if c.lastPromptLine != "" && (hasRepeatedPromptCR || hasRepeatedPromptCRLF) {
		if hasRepeatedPromptCRLF {
			data = data[len(c.lastPromptLine)+2:]
		} else {
			data = data[len(c.lastPromptLine)+1:]
		}

		// We just read until the repeated prompt, continue until the actual prompt
		if data == "" {
			data, prompt, err = c.readUntilPromptRaw(prompts...)
			if err != nil {
				return
			}
		}

		// Remove space as described above
		if c.lastPromptLine[len(c.lastPromptLine)-1] != ' ' && data[0] == ' ' {
			data = data[1:]
		}
	}

	// Remove echo of previously written line
	data = strings.TrimPrefix(data, c.lastWrittenLine+"\r\n")

	// The last received line is assumed to be the prompt
	index := strings.LastIndexAny(data, "\r\n")
	c.lastPromptLine = data[index+1:]

	// Remove prompt (last received line)
	if index != -1 {
		for index >= 1 && (data[index-1] == '\r' || data[index-1] == '\n') {
			index -= 1
		}
		data = data[0:index]
	} else {
		data = ""
	}

	// Detect prompt type and remove non-matching prompt sets from config
	promptType, err = c.handleReceivedPrompt(prompt)
	if err != nil {
		return
	}

	return
}

func (c *Client) connect(host, username string, passwordCallback dsl.PasswordCallback) error {
	if !regexpPort.MatchString(host) {
		host += ":23"
	}

	var err error
	c.conn, err = telnet.DialTimeout("tcp", host, 10*time.Second)
	if err != nil {
		return err
	}

	triedUsername := false
	triedPassword := false

	err = c.conn.SetDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		return err
	}

	for {
		prompts := c.getPromptList(promptTypeAccount | promptTypePassword | promptTypeCommand)
		_, prompt, err := c.readUntilPrompt(prompts...)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				return errors.New("no prompt detected")
			}
			return err
		}

		switch prompt {

		case promptTypeAccount:
			if triedUsername {
				return &dsl.AuthenticationError{Err: errors.New("invalid username/password")}
			}
			triedUsername = true

			err := c.writeLine(username, false)
			if err != nil {
				return err
			}

		case promptTypePassword:
			if triedPassword {
				return &dsl.AuthenticationError{Err: errors.New("invalid username/password")}
			}
			triedPassword = true

			var password string
			if passwordCallback != nil {
				password, err = passwordCallback()
				if err != nil {
					return &dsl.AuthenticationError{Err: err}
				}

				err = c.conn.SetDeadline(time.Now().Add(10 * time.Second))
				if err != nil {
					return err
				}
			}

			err := c.writeLine(password, true)
			if err != nil {
				return err
			}

		case promptTypeCommand:
			return nil

		}
	}
}

func (c *Client) Execute(command string) (string, error) {
	err := c.conn.SetDeadline(time.Now().Add(30 * time.Second))
	if err != nil {
		return "", &dsl.ConnectionError{Err: err}
	}

	err = c.writeLine(command, false)
	if err != nil {
		return "", &dsl.ConnectionError{Err: err}
	}

	prompts := c.getPromptList(promptTypeCommand)
	data, _, err := c.readUntilPrompt(prompts...)
	if err != nil {
		return "", &dsl.ConnectionError{Err: err}
	}

	return data, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
