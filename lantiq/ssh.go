// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lantiq

import (
	"3e8.eu/go/dsl/internal/ssh"
	"3e8.eu/go/dsl/models"
)

type SSHClient struct {
	command string
	client  *ssh.Client
	rawData []byte
	status  models.Status
	bins    models.Bins
}

func NewSSHClient(config SSHConfig) (*SSHClient, error) {
	c := SSHClient{}
	c.command = config.Command

	var err error

	c.client, err = ssh.NewClient(config.Host, config.User, config.Password, config.PrivateKeys, config.KnownHosts)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *SSHClient) RawData() []byte {
	return c.rawData
}

func (c *SSHClient) Status() models.Status {
	return c.status
}

func (c *SSHClient) Bins() models.Bins {
	return c.bins
}

func (c *SSHClient) UpdateData() (err error) {
	c.status, c.bins, c.rawData, err = updateData(c.client, c.command)
	return
}

func (c *SSHClient) Close() {
	c.client.Close()
}
