// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package models

type Inventory struct {
	Vendor  string
	Version string
}

func (i Inventory) String() string {
	var output string
	if i.Vendor != "" {
		output += i.Vendor
	} else {
		output += "Unknown"
	}
	if i.Version != "" {
		output += " " + i.Version
	}
	return output
}
