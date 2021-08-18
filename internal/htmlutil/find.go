// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package htmlutil

import (
	"golang.org/x/net/html"
)

func FindFirstNode(node *html.Node, matcher Matcher) *html.Node {
	if matcher(node) {
		return node
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		out := FindFirstNode(c, matcher)
		if out != nil {
			return out
		}
	}
	return nil
}

func FindLastNode(node *html.Node, matcher Matcher) *html.Node {
	if matcher(node) {
		return node
	}
	for c := node.LastChild; c != nil; c = c.PrevSibling {
		out := FindLastNode(c, matcher)
		if out != nil {
			return out
		}
	}
	return nil
}

func findAllNodesInternal(out *[]*html.Node, node *html.Node, matcher Matcher) {
	if matcher(node) {
		*out = append(*out, node)
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		findAllNodesInternal(out, c, matcher)
	}
}

func FindAllNodes(node *html.Node, matcher Matcher) []*html.Node {
	var out []*html.Node
	findAllNodesInternal(&out, node, matcher)
	return out
}
