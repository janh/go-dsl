// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package helpers

import (
	"sort"

	"3e8.eu/go/dsl/models"
)

type BandDecider struct {
	maxIndex     []int
	isDownstream []bool
}

func (d *BandDecider) IsDownstream(num int) bool {
	for i, max := range d.maxIndex {
		if num <= max {
			return d.isDownstream[i]
		}
	}
	return d.isDownstream[len(d.isDownstream)-1]
}

func NewBandDecider(bands models.BandsDownUp) *BandDecider {
	d := BandDecider{}

	bandsDown := bands.Downstream
	bandsUp := bands.Upstream

	d.maxIndex = make([]int, len(bandsDown)+len(bandsUp)-1)
	d.isDownstream = make([]bool, len(bandsDown)+len(bandsUp))

	sort.Slice(bandsDown, func(i, j int) bool { return bandsDown[i].Start < bandsDown[j].Start })
	sort.Slice(bandsUp, func(i, j int) bool { return bandsUp[i].Start < bandsUp[j].Start })

	var band, lastBand models.Band
	for i := 0; i < len(d.isDownstream); i++ {
		if len(bandsUp) == 0 || bandsDown[0].Start < bandsUp[0].Start {
			band = bandsDown[0]
			d.isDownstream[i] = true
			bandsDown = bandsDown[1:]
		} else {
			band = bandsUp[0]
			bandsUp = bandsUp[1:]
		}

		if i != 0 {
			d.maxIndex[i-1] = (lastBand.End + band.Start) / 2
		}

		lastBand = band
	}

	return &d
}
