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

type Client struct {
	config ClientConfig
	conn   *telnet.Conn
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

	err = c.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		return err
	}

	for {
		prompts := []string{c.config.PromptAccount, c.config.PromptPassword, c.config.PromptCommand}
		index, err := c.conn.SkipUntilIndex(prompts...)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				return errors.New("no prompt detected")
			}
			return err
		}
		prompt := prompts[index]

		switch prompt {

		case c.config.PromptAccount:
			if triedUsername {
				return errors.New("invalid username/password")
			}
			triedUsername = true

			c.conn.Write([]byte(username + "\r\n"))

		case c.config.PromptPassword:
			if triedPassword {
				return errors.New("invalid username/password")
			}
			triedPassword = true

			var password string
			if passwordCallback != nil {
				password = passwordCallback()

				err = c.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
				if err != nil {
					return err
				}
			}

			c.conn.Write([]byte(password + "\r\n"))

		case c.config.PromptCommand:
			return nil

		}
	}
}

func (c *Client) Execute(command string) (string, error) {
	c.conn.Write([]byte(command + "\r\n"))

	err := c.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		return "", err
	}

	data, err := c.conn.ReadUntil(c.config.PromptCommand)
	if err != nil {
		return "", err
	}
	if len(data) >= 2 && data[0] == '\r' && data[1] != '\n' {
		// found carriage return: we likely read the same prompt again, continue reading
		data, err = c.conn.ReadUntil(c.config.PromptCommand)
		if err != nil {
			return "", err
		}
	}
	str := string(data)

	if strings.HasPrefix(str, command+"\r\n") {
		str = str[len(command)+2:]
	}
	if index := strings.LastIndexAny(str, "\r\n"); index >= 0 {
		for index >= 1 && (str[index-1] == '\r' || str[index-1] == '\n') {
			index -= 1
		}
		str = str[0:index]
	}

	return str, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
