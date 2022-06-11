package synth

import "math"

type SoundGen func(buf []byte) (int, error) // generates the sineWave and reads it to a buffer

func (sg SoundGen) Read(buf []byte) (int, error) {
	return sg(buf)
}

// WithListener sends a copy of the generated audio data into a channel
func (sg SoundGen) WithListener(dest chan []byte) SoundGen {
	return func(buf []byte) (int, error) {
		select {
		case dest <- buf:
		default:
		}
		return sg(buf)
	}
}

func SineGen(cfg *Config, adapter SignalFunc) SoundGen {
	var lastFreq float64
	var lastVelocity float64
	var lastGate bool
	var pos float64
	return func(buf []byte) (int, error) {
		bytesRead := 0
		bytesPerSample := cfg.BitDepthInBytes * cfg.NumChannels
		numSamples := len(buf) / bytesPerSample
		deltaT := float64(1) / float64(cfg.SampleRate)
		for sampleIdx := 0; sampleIdx < numSamples; sampleIdx++ {
			signal := adapter()

			if signal.Gate && !lastGate {
				pos = 0
			}

			if signal.Freq != lastFreq { // resolve clicking on new notes and between frequency changes
				pos = (lastFreq * pos) / signal.Freq
			}

			if signal.Gate {
				signal.Vel *= 0.8 // scale the volume down a little
			} else {
				signal.Vel = lastVelocity * 0.9995 // decay
			}

			b := int16(math.Sin(2*math.Pi*float64(signal.Freq)*pos) * (math.MaxInt16 - 1) * signal.Vel)

			for channelIdx := 0; channelIdx < cfg.NumChannels; channelIdx++ {
				idx := (bytesPerSample * sampleIdx) + (channelIdx * cfg.BitDepthInBytes)
				buf[idx] = byte(b)
				buf[idx+1] = byte(b >> 8)
				bytesRead = idx + 2
			}

			lastFreq = signal.Freq
			lastVelocity = signal.Vel
			lastGate = signal.Gate
			pos += deltaT
		}
		return bytesRead, nil
	}
}
