// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package helpers

func GuessSNRGroupSize(maxValidSNRIndex, maxValidBitsIndex, binCount int) int {
	if binCount <= 512 || maxValidSNRIndex == 0 {
		return 1
	}

	maxGroupSize := binCount / maxValidSNRIndex

	var groupSize int
	for groupSize = 1; groupSize < maxGroupSize; groupSize *= 2 {
		// after applying groupSize, maxValidSNRIndex should be at most 10% lower than maxValidBitsIndex, because:
		// - maxValidSNRIndex > maxValidBitsIndex is common when SNR is too low to allocate bits
		// - maxValidSNRIndex < maxValidBitsIndex unlikely, as SNR value needed to allocate bins
		if float64(maxValidSNRIndex*groupSize)/float64(maxValidBitsIndex) > 0.9 {
			break
		}
	}

	return groupSize
}
