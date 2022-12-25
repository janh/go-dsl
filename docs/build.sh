#!/bin/sh
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at https://mozilla.org/MPL/2.0/.


[ $# -lt 2 ] || { echo "Too many arguments."; exit 1; }
[ $# -eq 1 ] || { echo "Usage: $0 TARGETDIR"; exit 1; }

target="$1"
[ -d "$target" ] || { echo "Target directory does not exist."; exit 1; }

[ -f README.md ] && [ -d docs ] || { echo "This script needs to be run from the root source folder."; exit 1; }


mkdir -p "$target/docs"
for doc in README.md docs/*.md; do
	out="$target/${doc%.*}.html"
	[ "$doc" != "README.md" ] && var="homelink:../README.html"
	pandoc \
		--from=markdown \
		--to=html5 \
		--template=docs/template.html \
		--variable="$var" \
		--lua-filter=docs/filter.lua \
		--output="$out" \
		"$doc"
done
