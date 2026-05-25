package mixer

import (
	"fmt"
)

var delayBufferSize uint
var delayBuffer []float32 //= make([]float32, delayBufferSize)
var delayReadIndex uint
var delayWriteIndex uint
var delayBufferIsFull bool
var empty []int32

//func Attach() {
//	rl.AttachAudioMixedProcessor(ProcessAudio)
//}
//
//func Detach() {
//	rl.DetachAudioMixedProcessor(ProcessAudio)
//}

// ProcessAudio is the audio processing function
func ProcessAudio(buffer []float32, frames int) {

	if len(delayBuffer) == 0 {
		fmt.Println("Initializing")
		delayBufferSize = uint(frames * 200)
		delayReadIndex = 0
		delayWriteIndex = uint(frames)
		delayBuffer = make([]float32, delayBufferSize)
		delayStepBuffer = make([]float32, frames)
		empty = make([]int32, frames)
		fmt.Printf("Empty len %d\n", len(empty))
	}
	copy(delayBuffer[delayWriteIndex:int(delayWriteIndex)+frames], buffer)
	// copy(buffer, delayBuffer[delayReadIndex:int(delayReadIndex)+len(buffer)])
	if delayBufferIsFull {
		//for i := 0; i < frames; i += 2 {
		//	delayStepBuffer[i] = 0.5*buffer[i] + 0.5*delayBuffer[int(delayReadIndex)+i]
		//}
		////		fmt.Printf("delayWriteIndex:%d, frames: %d\n", delayWriteIndex, frames)
		//copy(buffer, delayStepBuffer)

		//copied := copy(buffer, []float32(empty))
		//fmt.Printf("Empty len %d Copied %d Frames %d Buffer len %d\n", len(empty), copied, frames, len(buffer))
		return
	}

	delayWriteIndex += uint(frames)
	if delayWriteIndex+uint(len(buffer)) > delayBufferSize {
		delayWriteIndex = 0
		delayBufferIsFull = true
	}

	delayReadIndex += uint(frames)
	if delayReadIndex+uint(len(buffer)) > delayBufferSize {
		delayReadIndex = 0
	}

	//delayReadIndex += uint(len(buffer))
	//if delayReadIndex+uint(len(buffer)) > uint(len(delayBuffer)) {
	//	delayReadIndex = 0
	//}

	//	ExposedDelayBuffer = delayBuffer[delayReadIndex : int(delayReadIndex)+maxFrame]
}
