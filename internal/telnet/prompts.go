// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package telnet

import (
	"errors"
)

type promptType int

const (
	promptTypeAccount promptType = 1 << iota
	promptTypePassword
	promptTypeCommand
)

func (c *Client) getPromptList(types promptType) []string {
	prompts := []string{}

	for _, promptData := range c.config.Prompts {
		if types&promptTypeAccount != 0 {
			prompts = append(prompts, promptData.Account)
		}
		if types&promptTypePassword != 0 {
			prompts = append(prompts, promptData.Password)
		}
		if types&promptTypeCommand != 0 {
			prompts = append(prompts, promptData.Command)
		}
	}

	return prompts
}

func (c *Client) handleReceivedPrompt(prompt string) (typeVal promptType, err error) {
	i := 0

	for _, promptData := range c.config.Prompts {
		var t promptType

		switch prompt {
		case promptData.Account:
			t = promptTypeAccount
		case promptData.Password:
			t = promptTypePassword
		case promptData.Command:
			t = promptTypeCommand
		}

		if t != 0 {
			if typeVal == 0 {
				typeVal = t
			} else if typeVal != t {
				// this should only happen if there is an error in the configuration
				err = errors.New("ambiguous prompt detected")
				return
			}

			c.config.Prompts[i] = promptData
			i++
		}
	}

	c.config.Prompts = c.config.Prompts[:i]

	if typeVal == 0 {
		err = errors.New("unrecognized prompt")
		return
	}

	return
}
