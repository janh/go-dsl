// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lantiq

import (
	"3e8.eu/go/dsl"
	"3e8.eu/go/dsl/internal/ssh"
	"3e8.eu/go/dsl/models"
)

type sshClient struct {
	command string
	client  *ssh.Client
	rawData []byte
	status  models.Status
	bins    models.Bins
}

func NewSSHClient(config SSHConfig) (dsl.Client, error) {
	c := sshClient{}
	c.command = config.Command

	var err error

	c.client, err = ssh.NewClient(config.Host, config.User, config.Password, config.PrivateKeys, config.KnownHosts, config.InsecureAlgorithms)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *sshClient) RawData() []byte {
	return c.rawData
}

func (c *sshClient) Status() models.Status {
	return c.status
}

func (c *sshClient) Bins() models.Bins {
	return c.bins
}

func (c *sshClient) UpdateData() (err error) {
	c.status, c.bins, c.rawData, err = updateData(c.client, c.command)
	return
}

func (c *sshClient) Close() {
	c.client.Close()
}
