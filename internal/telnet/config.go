// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package telnet

type Prompts struct {
	Account  string
	Password string
	Command  string
}

type ClientConfig struct {
	Prompts []Prompts

	ExpectRepeatedPromptCRLF bool
}
