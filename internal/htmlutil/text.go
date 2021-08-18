// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package htmlutil

import (
	"strings"

	"golang.org/x/net/html"
)

func GetText(node *html.Node) string {
	textNodes := FindAllNodes(node, func(n *html.Node) bool {
		return n.Type == html.TextNode
	})

	out := ""
	for _, n := range textNodes {
		content := strings.TrimSpace(n.Data)
		if len(content) == 0 {
			continue
		}
		if len(out) != 0 {
			out += " "
		}
		out += content
	}

	return out
}
