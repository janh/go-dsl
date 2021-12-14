// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"errors"
	"time"

	"3e8.eu/go/dsl"
	"3e8.eu/go/dsl/models"
)

type State string

const (
	StateReady              State = "ready"
	StatePasswordRequired   State = "password"
	StatePassphraseRequired State = "passphrase"
	StateLoading            State = "loading"
	StateError              State = "error"
)

type stateChange struct {
	State State

	Time    time.Time
	RawData []byte
	Status  models.Status
	Bins    models.Bins

	Fingerprint string

	Err error
}

type client struct {
	setPassword        chan string
	setPassphrase      chan string
	changeState        chan stateChange
	registerReceiver   chan chan stateChange
	unregisterReceiver chan chan stateChange

	receivers       map[chan stateChange]bool
	lastStateChange stateChange

	client dsl.Client

	cancel chan bool
	done   chan bool

	errCount int

	config     dsl.Config
	password   string
	passphrase map[string]string
}

func newClient(config dsl.Config) *client {
	c := &client{
		setPassword:        make(chan string),
		setPassphrase:      make(chan string),
		changeState:        make(chan stateChange),
		registerReceiver:   make(chan chan stateChange),
		unregisterReceiver: make(chan chan stateChange),
		receivers:          make(map[chan stateChange]bool),
		lastStateChange:    stateChange{State: StateLoading},
		cancel:             make(chan bool),
		done:               make(chan bool),
		config:             config,
		passphrase:         make(map[string]string),
	}

	go c.distribute()
	go c.update()

	return c
}

func (c *client) State() stateChange {
	receiver := make(chan stateChange, 1)
	c.registerReceiver <- receiver
	change := <-receiver
	c.unregisterReceiver <- receiver
	return change
}

func (c *client) SetPassword(password string) error {
	select {
	case c.setPassword <- password:
		return nil
	default:
		return errors.New("no password required")
	}
}

func (c *client) SetPassphrase(passphrase string) error {
	select {
	case c.setPassphrase <- passphrase:
		return nil
	default:
		return errors.New("no password required")
	}
}

func (c *client) RegisterReceiver(receiver chan stateChange) {
	c.registerReceiver <- receiver
}

func (c *client) UnregisterReceiver(receiver chan stateChange) {
	c.unregisterReceiver <- receiver
}

func (c *client) distribute() {
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

		case receiver := <-c.unregisterReceiver:
			delete(c.receivers, receiver)

		}
	}
}

func (c *client) connect() (canceled bool) {
	var err error
	var interval = 2 * time.Second

	for {
		c.client, err = dsl.NewClient(c.config)
		if err == nil {
			c.errCount = 0
			return
		}

		var authErr *dsl.AuthenticationError
		if errors.As(err, &authErr) {
			c.password = ""
			c.passphrase = make(map[string]string)

			if authErr.WaitTime > interval {
				interval = authErr.WaitTime
			}
		}

		c.changeState <- stateChange{State: StateError, Err: err}

		select {
		case <-c.cancel:
			canceled = true
			return
		case <-time.After(interval):
		}

		interval *= 2
		if interval > 30*time.Second {
			interval = 30 * time.Second
		}
	}
}

func (c *client) update() {
	clientDesc := c.config.Type.ClientDesc()

	if clientDesc.SupportedAuthTypes&dsl.AuthTypePassword != 0 {
		c.config.AuthPassword = func() string {
			if c.password == "" {
				c.changeState <- stateChange{State: StatePasswordRequired}
				c.password = <-c.setPassword
				c.changeState <- stateChange{State: StateLoading}
			}
			return c.password
		}
	}

	if clientDesc.SupportedAuthTypes&dsl.AuthTypePrivateKeys != 0 {
		c.config.AuthPrivateKeys.Passphrase = func(fingerprint string) string {
			if c.passphrase[fingerprint] == "" {
				c.changeState <- stateChange{State: StatePasswordRequired, Fingerprint: fingerprint}
				c.passphrase[fingerprint] = <-c.setPassphrase
				c.changeState <- stateChange{State: StateLoading}
			}
			return c.passphrase[fingerprint]
		}
	}

mainloop:
	for {
		for i := 0; i < 2; i++ {
			if c.client == nil {
				canceled := c.connect()
				if canceled {
					break mainloop
				}
			}

			err := c.client.UpdateData()

			if err == nil {

				c.changeState <- stateChange{
					State:   StateReady,
					Time:    time.Now(),
					RawData: c.client.RawData(),
					Status:  c.client.Status(),
					Bins:    c.client.Bins(),
				}

				c.errCount = 0

				break

			} else {

				c.changeState <- stateChange{State: StateError, Err: err}

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

		now := time.Now()
		nextUpdate := now.Truncate(30 * time.Second).Add(30 * time.Second)

		select {
		case <-c.cancel:
			break mainloop
		case <-time.After(nextUpdate.Sub(now)):
		}
	}

	if c.client != nil {
		c.client.Close()
	}

	c.done <- true
}

func (c *client) close() {
	c.cancel <- true
	<-c.done
}
