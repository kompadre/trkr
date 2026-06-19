package audio

import "math"
import . "trkr"

type ZeroCrossing struct {
	Index float64
	RMS   float32 // Root-Mean-Square (volume) at this neighborhood
}

func FindOptimalLoopPoints(buffer []float32) (start float64, end float64) {
	if len(buffer) < 100 {
		return 0, 0
	}

	var candidates []ZeroCrossing
	var zeroThreshold float32 = 0.005 // How close to zero we demand

	// 1. Scan the buffer to find ALL valid zero-crossings moving UPWARD
	startWindow := 4410
	for i := startWindow; i < len(buffer)-1; i++ {
		prev := buffer[i-1]
		curr := buffer[i]

		// Check if it crosses 0 moving from negative to positive
		if prev <= 0 && curr > 0 && max(curr, -curr) < zeroThreshold {

			// Calculate local amplitude (RMS) over a tiny 16-sample window
			// to ensure the audio is "still going strong" and not silent tail
			var sumSq float32
			window := 16
			if i+window < len(buffer) {
				for w := 0; w < window; w++ {
					sumSq += buffer[i+w] * buffer[i+w]
				}
				rms := float32(math.Sqrt(float64(sumSq / float32(window))))

				candidates = append(candidates, ZeroCrossing{
					Index: float64(i),
					RMS:   rms,
				})
			}
		}
	}

	// 2. Pair them up: Look for a 'Start' early on, and an 'End' later
	// where the amplitude profile matches or is strong.
	if len(candidates) < 2 {
		return 0, 0
	}

	// Simple heuristic:
	// Pick a start point in the first 25% of the sample (sustain entry)
	// Pick an end point in the middle/late section where RMS is still high

	bestStart := candidates[0].Index
	bestEnd := float64(len(buffer) - 1)
	minDistance := float64(999999) // Look for the TIGHTEST loop, not the biggest

	for _, cStart := range candidates {
		if cStart.Index > float64(len(buffer))*0.2 {
			break
		}

		for _, cEnd := range candidates {
			if cEnd.Index <= cStart.Index {
				continue
			}

			distance := cEnd.Index - cStart.Index

			// FORCE a tight loop: e.g., between 100 and 2000 samples long
			// This corresponds to a localized waveform cycle rather than a massive chunk
			if distance > 100 && distance < 2000 {
				if cEnd.RMS > cStart.RMS*0.9 { // Demands very tight volume matching
					if distance < minDistance {
						minDistance = distance
						bestStart = cStart.Index
						bestEnd = cEnd.Index
					}
				}
			}
		}
	}

	Logf("BestStart: %f, BestEnd: %f.\n", bestStart, bestEnd)
	return bestStart, bestEnd
}

func CalibrateSustainLoop(buffer []float32, roughStart int, targetLength float64) (float64, float64) {
	var loopStart float64 = float64(roughStart)
	var loopEnd float64 = float64(roughStart) + targetLength

	// 1. Find the exact upward zero-crossing for the START
	for i := roughStart; i < len(buffer)-1; i++ {
		if buffer[i-1] <= 0 && buffer[i] > 0 {
			loopStart = float64(i)
			break
		}
	}

	// 2. Calculate ideal end point based on the newly locked start
	idealEnd := loopStart + targetLength
	idealEndIdx := int(idealEnd)

	// 3. Scan a tight neighborhood (+/- 15 samples) around idealEnd
	// to find the matching upward zero-crossing
	minDelta := 999.0
	searchRadius := 15

	for i := idealEndIdx - searchRadius; i <= idealEndIdx+searchRadius; i++ {
		if i > 0 && i < len(buffer)-1 {
			// Is it an upward crossing?
			if buffer[i-1] <= 0 && buffer[i] > 0 {
				delta := math.Abs(float64(i) - idealEnd)
				if delta < minDelta {
					minDelta = delta
					loopEnd = float64(i)
				}
			}
		}
	}

	return loopStart, loopEnd
}

func SmoothSampleLoop(buffer []float32, start, end int) {
	// A tiny crossfade window (e.g., 32 samples out of a 1468-sample loop)
	fadeWindow := 32
	if (end - start) < fadeWindow*2 {
		return // Loop too short to crossfade safely
	}

	for i := 0; i < fadeWindow; i++ {
		// Calculate the fade weights
		t := float32(i) / float32(fadeWindow)

		// Mirror the audio data right before the loop end
		// into the audio data right at the loop start
		endTargetIdx := end - fadeWindow + i
		startTargetIdx := start + i

		if endTargetIdx < len(buffer) && startTargetIdx < len(buffer) {
			// Smoothly blend the waveform tail back into the head
			buffer[startTargetIdx] = (1.0-t)*buffer[startTargetIdx] + t*buffer[endTargetIdx]
		}
	}

	// Force the exact loop boundary samples to match perfectly to 0 to prevent clicks
	buffer[start] = 0.0
	buffer[end] = 0.0
}
