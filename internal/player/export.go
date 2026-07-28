package player

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	. "trkr"
	"trkr/internal/audio"
)

func ExportWav(filename string) error {
	fmt.Printf("Exporting to %s...\n", filename)
	IsExporting = true
	defer func() { IsExporting = false }()

	if IsPlaying {
		Stop()
	}

	// Prepare audio engine for offline rendering
	audio.StopFm()
	audio.CutAllVoices()
	audio.ResetSaturator()
	audio.PlayFmPickup()

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Placeholder for header
	placeholder := make([]byte, 44)
	f.Write(placeholder)

	sampleRate := uint32(48000)
	bpm := BeatsPerMinute
	// Samples per tick with high precision
	samplesPerTickFloat := (60.0 / float64(bpm) / 4.0) * float64(sampleRate)

	totalSamplesWritten := 0
	
	// Internal render block size (MUST be consistent for limiter behavior)
	renderBlockSize := 2048
	renderBuffer := make([]float32, renderBlockSize)
	pcmBuffer := make([]int16, renderBlockSize)
	
	// Overflow buffer for partial ticks
	overflowBuffer := make([]float32, 0)

	accumulatedSamplesGoal := 0.0

	for sectionIdx := 0; sectionIdx < len(CurrentProject.Sections); sectionIdx++ {
		fmt.Printf("Rendering section %d/%d (%s)...\n", sectionIdx+1, len(CurrentProject.Sections), CurrentProject.Sections[sectionIdx].Name)
		Head = NewPlayhead(uint8(sectionIdx))
		sectionRows := int(CurrentProject.Sections[sectionIdx].Rows)

		for row := 0; row < sectionRows; row++ {
			Tick(int64(row))
			audio.MasterQueueCommands()
			
			accumulatedSamplesGoal += samplesPerTickFloat
			targetSamples := int(accumulatedSamplesGoal)
			samplesNeededForThisTick := targetSamples - totalSamplesWritten

			// While we haven't satisfied the sample count for this tick
			for samplesNeededForThisTick > 0 {
				// 1. Process block if overflow is empty
				if len(overflowBuffer) == 0 {
					audio.Mix(renderBuffer)
					overflowBuffer = append(overflowBuffer, renderBuffer...)
				}

				// 2. Determine how much to take from overflow
				take := samplesNeededForThisTick
				if take > len(overflowBuffer) {
					take = len(overflowBuffer)
				}

				// 3. Convert current slice to PCM
				for i := 0; i < take; i++ {
					s := overflowBuffer[i]
					if s > 1.0 { s = 1.0 } else if s < -1.0 { s = -1.0 }
					pcmBuffer[i] = int16(s * 32767.0)
				}

				// 4. Write to file
				binary.Write(f, binary.LittleEndian, pcmBuffer[:take])
				
				// 5. Update counters
				totalSamplesWritten += take
				samplesNeededForThisTick -= take
				
				// 6. Consume from overflow
				overflowBuffer = overflowBuffer[take:]
			}
		}
	}

	// Finalize header
	f.Seek(0, 0)
	var headerBuf bytes.Buffer
	audio.WaveHeader(&headerBuf, uint32(totalSamplesWritten), sampleRate)
	f.Write(headerBuf.Bytes())

	fmt.Printf("Export complete: %d samples (%.2f seconds).\n", totalSamplesWritten, float64(totalSamplesWritten)/float64(sampleRate))
	return nil
}
