// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package graphs

import (
	_ "embed"
	"io"
	"text/template"
)

//go:embed templates/base.tmpl
var templateBase string

//go:embed templates/bits.tmpl
var templateBits string

//go:embed templates/snr.tmpl
var templateSNR string

//go:embed templates/qln.tmpl
var templateQLN string

//go:embed templates/hlog.tmpl
var templateHlog string

//go:embed templates/errors.tmpl
var templateErrors string

func writeTemplate(w io.Writer, data interface{}, templates ...string) error {
	t := template.New("")
	for _, tpl := range templates {
		t = template.Must(t.Parse(tpl))
	}

	return t.Execute(w, data)
}
