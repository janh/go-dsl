// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ssh

import (
	"errors"
	"regexp"

	"golang.org/x/crypto/ssh"
)

var regexpPort = regexp.MustCompile(`:[0-9]+$`)

type Client struct {
	client *ssh.Client
}

func NewClient(host, username, password, privateKey, knownHost string) (*Client, error) {
	c := Client{}

	err := c.connect(host, username, password, privateKey, knownHost)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *Client) connect(host, username, password, privateKey, knownHost string) error {
	if !regexpPort.MatchString(host) {
		host += ":22"
	}

	config := &ssh.ClientConfig{User: username}

	if privateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(privateKey))
		if err != nil {
			return err
		}

		config.Auth = append(config.Auth, ssh.PublicKeys(signer))
	}

	if password != "" {
		config.Auth = append(config.Auth, ssh.Password(password))
	}

	if knownHost == "" {
		return errors.New("missing SSH host key")
	} else if knownHost == "IGNORE" {
		config.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	} else {
		hostKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(knownHost))
		if err != nil {
			return err
		}

		config.HostKeyCallback = ssh.FixedHostKey(hostKey)
	}

	var err error
	c.client, err = ssh.Dial("tcp", host, config)
	return err
}

func (c *Client) Execute(command string) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func (c *Client) Close() error {
	return c.client.Close()
}
