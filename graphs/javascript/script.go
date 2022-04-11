// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package javascript

import (
	_ "embed"
)

//go:embed graphs.js
var script []byte

// Script returns a Javascript library for drawing graphs to an HTML canvas.
func Script() []byte {
	return script
}
