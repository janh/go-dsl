// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package common

import (
	"errors"
	"time"

	"3e8.eu/go/dsl"
	"3e8.eu/go/dsl/history"
	"3e8.eu/go/dsl/models"
)

type State string

const (
	StateReady                        State = "ready"
	StatePasswordRequired             State = "password"
	StatePassphraseRequired           State = "passphrase"
	StateEncryptionPassphraseRequired State = "encryption-passphrase"
	StateLoading                      State = "loading"
	StateError                        State = "error"
)

const (
	intervalDefault time.Duration = 30 * time.Second
	intervalShort   time.Duration = 10 * time.Second
)

type StateChange struct {
	State State

	Time        time.Time
	RawData     []byte
	Status      models.Status
	Bins        models.Bins
	BinsHistory models.BinsHistory

	Fingerprint string

	Err error
}

type Client struct {
	setPassword             chan string
	setPassphrase           chan string
	setEncryptionPassphrase chan string
	changeState             chan StateChange
	registerReceiver        chan chan StateChange
	unregisterReceiver      chan chan StateChange
	stopDistribute          chan bool

	receivers       map[chan StateChange]bool
	lastStateChange StateChange

	client dsl.Client

	interval        time.Duration
	intervalChanged chan bool

	canceled bool
	cancel   chan bool
	done     chan bool

	errCount int

	config               dsl.Config
	password             string
	passphrase           map[string]string
	encryptionPassphrase string
}

func NewClient(config dsl.Config) *Client {
	c := &Client{
		setPassword:             make(chan string),
		setPassphrase:           make(chan string),
		setEncryptionPassphrase: make(chan string),
		changeState:             make(chan StateChange),
		registerReceiver:        make(chan chan StateChange),
		unregisterReceiver:      make(chan chan StateChange),
		stopDistribute:          make(chan bool),
		receivers:               make(map[chan StateChange]bool),
		lastStateChange:         StateChange{State: StateLoading},
		interval:                intervalDefault,
		intervalChanged:         make(chan bool),
		cancel:                  make(chan bool),
		done:                    make(chan bool),
		config:                  config,
		passphrase:              make(map[string]string),
	}

	go c.distribute()
	go c.update()

	return c
}

func (c *Client) State() StateChange {
	receiver := make(chan StateChange, 1)
	c.registerReceiver <- receiver
	change := <-receiver
	c.unregisterReceiver <- receiver
	return change
}

func (c *Client) SetPassword(password string) error {
	select {
	case c.setPassword <- password:
		return nil
	default:
		return errors.New("no password required")
	}
}

func (c *Client) SetPassphrase(passphrase string) error {
	select {
	case c.setPassphrase <- passphrase:
		return nil
	default:
		return errors.New("no passphrase required")
	}
}

func (c *Client) SetEncryptionPassphrase(encryptionPassphrase string) error {
	select {
	case c.setEncryptionPassphrase <- encryptionPassphrase:
		return nil
	default:
		return errors.New("no encryption passphrase required")
	}
}

func (c *Client) RegisterReceiver(receiver chan StateChange) {
	c.registerReceiver <- receiver
}

func (c *Client) UnregisterReceiver(receiver chan StateChange) {
	c.unregisterReceiver <- receiver
}

func (c *Client) distribute() {
	for {
		select {

		case change := <-c.changeState:
			c.lastStateChange = change
			for receiver := range c.receivers {
				select {
				case receiver <- change:
				default:
					delete(c.receivers, receiver)
				}
			}

		case receiver := <-c.registerReceiver:
			select {
			case receiver <- c.lastStateChange:
			default:
				return
			}
			c.receivers[receiver] = true

			if len(c.receivers) == 1 {
				c.interval = intervalShort
				select {
				case c.intervalChanged <- true:
				default:
				}
			}

		case receiver := <-c.unregisterReceiver:
			delete(c.receivers, receiver)

			if len(c.receivers) == 0 {
				c.interval = intervalDefault
				select {
				case c.intervalChanged <- true:
				default:
				}
			}

		case <-c.stopDistribute:
			return

		}
	}
}

func (c *Client) connect() {
	var err error
	var interval = 2 * time.Second

	for {
		c.client, err = dsl.NewClient(c.config)
		if err == nil {
			c.errCount = 0
			return
		}
		if c.canceled {
			return
		}

		var authErr *dsl.AuthenticationError
		if errors.As(err, &authErr) {
			c.password = ""
			c.passphrase = make(map[string]string)
			c.encryptionPassphrase = ""

			if authErr.WaitTime > interval {
				interval = authErr.WaitTime
			}
		}

		c.changeState <- StateChange{State: StateError, Err: err}

		select {
		case <-c.cancel:
			c.canceled = true
			return
		case <-time.After(interval):
		}

		interval *= 2
		if interval > 30*time.Second {
			interval = 30 * time.Second
		}
	}
}

func (c *Client) update() {
	clientDesc := c.config.Type.ClientDesc()

	if clientDesc.SupportedAuthTypes&dsl.AuthTypePassword != 0 && c.config.AuthPassword == nil {
		c.config.AuthPassword = func() (string, error) {
			if c.password == "" {
				c.changeState <- StateChange{State: StatePasswordRequired}

				select {
				case <-c.cancel:
					c.canceled = true
					return "", errors.New("canceled")
				case password := <-c.setPassword:
					c.password = password
				}

				c.changeState <- StateChange{State: StateLoading}
			}

			return c.password, nil
		}
	}

	if clientDesc.SupportedAuthTypes&dsl.AuthTypePrivateKeys != 0 && c.config.AuthPrivateKeys.Passphrase == nil {
		c.config.AuthPrivateKeys.Passphrase = func(fingerprint string) (string, error) {
			if c.passphrase[fingerprint] == "" {
				c.changeState <- StateChange{State: StatePassphraseRequired, Fingerprint: fingerprint}

				select {
				case <-c.cancel:
					c.canceled = true
					return "", errors.New("canceled")
				case passphrase := <-c.setPassphrase:
					c.passphrase[fingerprint] = passphrase
				}

				c.changeState <- StateChange{State: StateLoading}
			}

			return c.passphrase[fingerprint], nil
		}
	}

	if clientDesc.SupportsEncryptionPassphrase && c.config.EncryptionPassphrase == nil {
		c.config.EncryptionPassphrase = func() (string, error) {
			if c.encryptionPassphrase == "" {
				c.changeState <- StateChange{State: StateEncryptionPassphraseRequired}

				select {
				case <-c.cancel:
					c.canceled = true
					return "", errors.New("canceled")
				case encryptionPassphrase := <-c.setEncryptionPassphrase:
					c.encryptionPassphrase = encryptionPassphrase
				}

				c.changeState <- StateChange{State: StateLoading}
			}

			return c.encryptionPassphrase, nil
		}
	}

	history, err := history.NewBins(history.DefaultBinsConfig)
	if err != nil {
		panic(err)
	}

mainloop:
	for {
		for i := 0; i < 2; i++ {
			if c.client == nil {
				c.connect()
				if c.canceled {
					break mainloop
				}
			}

			err := c.client.UpdateData()

			if err == nil {

				now := time.Now()

				history.Update(c.client.Status(), c.client.Bins(), now)

				c.changeState <- StateChange{
					State:       StateReady,
					Time:        now,
					RawData:     c.client.RawData(),
					Status:      c.client.Status(),
					Bins:        c.client.Bins(),
					BinsHistory: history.Data(),
				}

				c.errCount = 0

				break

			} else {

				c.changeState <- StateChange{State: StateError, Err: err}

				c.errCount++

				var connErr *dsl.ConnectionError
				if errors.As(err, &connErr) || c.errCount == 10 {
					c.client.Close()
					c.client = nil
				} else {
					break
				}

			}
		}

	waitloop:
		for {
			now := time.Now()
			nextUpdate := now.Truncate(c.interval).Add(c.interval)

			select {
			case <-c.cancel:
				break mainloop
			case <-time.After(nextUpdate.Sub(now)):
				break waitloop
			case <-c.intervalChanged:
				continue
			}
		}
	}

	if c.client != nil {
		c.client.Close()
	}

	c.stopDistribute <- true
	c.done <- true
}

func (c *Client) Close() {
	c.cancel <- true
	<-c.done
}
