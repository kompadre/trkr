package effects

// Delay implements a feedback delay audio effect using a circular buffer.
type Delay struct {
	DelayTime  float64 // Delay time in seconds
	Feedback   float64 // Feedback gain coefficient (0.0 to < 1.0)
	Mix        float64 // Dry/Wet mix ratio (0.0 = all dry, 1.0 = all wet)
	SampleRate float64 // Sample rate in Hz

	buffer   []float32
	writePos int
}

const (
	defaultMaxDelayTime = 5.0 // Maximum supported delay time in seconds
)

// NewDelay creates and initializes a new Delay effect.
func NewDelay(delayTime float64, feedback float64, mix float64, sampleRate float64) *Delay {
	if sampleRate <= 0 {
		sampleRate = 48000.0
	}
	d := &Delay{
		DelayTime:  delayTime,
		Feedback:   feedback,
		Mix:        mix,
		SampleRate: sampleRate,
	}
	d.ensureBuffer()
	return d
}

// ensureBuffer allocates or expands the delay circular buffer if needed.
func (d *Delay) ensureBuffer() {
	maxTime := defaultMaxDelayTime
	if d.DelayTime > maxTime {
		maxTime = d.DelayTime
	}
	minLen := int(maxTime * d.SampleRate)
	if minLen < 1 {
		minLen = 1
	}
	if len(d.buffer) < minLen {
		d.buffer = make([]float32, minLen)
		d.writePos = 0
	}
}

// SetDelayTime updates the delay time in seconds and ensures buffer capacity.
func (d *Delay) SetDelayTime(delayTime float64) {
	if d.DelayTime != delayTime {
		d.DelayTime = delayTime
		d.ensureBuffer()
	}
}

// SetFeedback updates the feedback gain (clamped between 0.0 and 0.99).
func (d *Delay) SetFeedback(feedback float64) {
	if feedback < 0.0 {
		feedback = 0.0
	} else if feedback >= 1.0 {
		feedback = 0.99
	}
	d.Feedback = feedback
}

// SetMix updates the dry/wet mix ratio (clamped between 0.0 and 1.0).
func (d *Delay) SetMix(mix float64) {
	if mix < 0.0 {
		mix = 0.0
	} else if mix > 1.0 {
		mix = 1.0
	}
	d.Mix = mix
}

// SetSampleRate updates the sample rate and re-allocates buffer if required.
func (d *Delay) SetSampleRate(sampleRate float64) {
	if sampleRate <= 0 {
		sampleRate = 48000.0
	}
	if d.SampleRate != sampleRate {
		d.SampleRate = sampleRate
		d.ensureBuffer()
	}
}

// Reset clears the delay buffer and resets buffer positions.
func (d *Delay) Reset() {
	for i := range d.buffer {
		d.buffer[i] = 0.0
	}
	d.writePos = 0
}

// ProcessSample processes a single sample through the delay effect.
func (d *Delay) ProcessSample(sample float32) float32 {
	if len(d.buffer) == 0 {
		return sample
	}

	fb := d.Feedback
	if fb < 0.0 {
		fb = 0.0
	} else if fb >= 1.0 {
		fb = 0.99
	}

	mix := d.Mix
	if mix < 0.0 {
		mix = 0.0
	} else if mix > 1.0 {
		mix = 1.0
	}

	delayTime := d.DelayTime
	if delayTime < 0.0 {
		delayTime = 0.0
	}
	delaySamples := int(delayTime * d.SampleRate)
	if delaySamples < 1 {
		delaySamples = 1
	} else if delaySamples > len(d.buffer) {
		delaySamples = len(d.buffer)
	}

	bufLen := len(d.buffer)
	readPos := (d.writePos - delaySamples + bufLen) % bufLen
	delayed := d.buffer[readPos]

	d.buffer[d.writePos] = sample + float32(fb)*delayed

	d.writePos = (d.writePos + 1) % bufLen

	return (1.0-float32(mix))*sample + float32(mix)*delayed
}

// Process processes a slice of audio samples in place.
func (d *Delay) Process(samples []float32) {
	for i, sample := range samples {
		samples[i] = d.ProcessSample(sample)
	}
}
