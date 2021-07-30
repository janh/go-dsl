// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ssh

import (
	"bufio"
	"errors"
	"regexp"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

var regexpPort = regexp.MustCompile(`:[0-9]+$`)

type Client struct {
	client *ssh.Client
}

func NewClient(host, username, password, privateKey, knownHosts string) (*Client, error) {
	c := Client{}

	err := c.connect(host, username, password, privateKey, knownHosts)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *Client) connect(host, username, password, privateKey, knownHosts string) error {
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

	if knownHosts == "" {
		return errors.New("missing SSH host key")
	} else if knownHosts == "IGNORE" {
		config.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	} else {
		var hostKey ssh.PublicKey

		hostNormalized := knownhosts.Normalize(host)

		scanner := bufio.NewScanner(strings.NewReader(knownHosts))
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)

			if len(line) == 0 {
				continue
			}

			_, hosts, key, _, _, err := ssh.ParseKnownHosts([]byte(line))
			if err != nil {
				return err
			}

			for _, h := range hosts {
				if h == hostNormalized {
					hostKey = key
					break
				}
			}
		}

		if hostKey == nil {
			return errors.New("no matching SSH host key found")
		}

		config.HostKeyAlgorithms = []string{hostKey.Type()}
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
