// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package history

import (
	"errors"
	"fmt"
	"io"
)

type storageMainHeader struct {
	Version      uint32
	CreationTime int64
}

func checkEndOfFile(r io.Reader) error {
	b := make([]byte, 1)
	_, err := r.Read(b)

	if err == nil {
		return errors.New("unexpected trailing data")
	} else if err == io.EOF {
		return nil
	} else {
		return fmt.Errorf("read error: %w", err)
	}
}
