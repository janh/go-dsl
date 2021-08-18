// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package htmlutil

import (
	"strings"

	"golang.org/x/net/html"
)

type Matcher func(*html.Node) bool

func MatcherTagName(tagName string) Matcher {
	return func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == tagName
	}
}

func MatcherTagNameAndClass(tagName string, class string) Matcher {
	return func(n *html.Node) bool {
		if n.Type != html.ElementNode || n.Data != "tr" {
			return false
		}
		for _, attr := range n.Attr {
			if attr.Key == "class" {
				classes := strings.Fields(attr.Val)
				for _, c := range classes {
					if c == class {
						return true
					}
				}
				break
			}
		}
		return false
	}
}
