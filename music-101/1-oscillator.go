// y(t)=sin(2πft)

// phase ide od 0 do 1
// increment = freq / sampleRate

// go get [github.com/ebitengine/oto/v3](https://github.com/ebitengine/oto/v3)
// go get github.com/ebitengine/oto/v3

package main

import (
	"math"
	"time"

	"github.com/ebitengine/oto/v3"
)

func main() {
	const (
		sampleRate = 44100
		frequency  = 440.0 // The pitch (A4)
		duration   = 2 * time.Second
	)

	// 1. Initialize the audio context
	op := &oto.NewContextOptions{
		SampleRate:   sampleRate,
		ChannelCount: 1, // Mono
		Format:       oto.FormatFloat32LE,
		BufferSize:   32,
	}
	otoCtx, ready, _ := oto.NewContext(op)
	<-ready // Wait for the audio hardware to be ready

	// 2. Create a player
	p := otoCtx.NewPlayer(newSineWave(frequency, sampleRate))
	p.Play()

	// 3. Let it play for the duration
	time.Sleep(duration)
}

// sineWave implements the io.Reader interface to stream audio
type sineWave struct {
	freq   float64
	sRate  float64
	sample int
}

func newSineWave(freq, sRate float64) *sineWave {
	return &sineWave{freq, sRate, 0}
}

func (s *sineWave) Read(p []byte) (n int, err error) {
	// We need to fill p with 4-byte float32 values
	for i := 0; i < len(p)/4; i++ {
		// Calculate the sine value
		v := math.Sin(2 * math.Pi * s.freq * float64(s.sample) / s.sRate)
		s.sample++

		// Convert float64 to float32 bits
		bits := math.Float32bits(float32(v))
		p[i*4] = byte(bits)
		p[i*4+1] = byte(bits >> 8)
		p[i*4+2] = byte(bits >> 16)
		p[i*4+3] = byte(bits >> 24)

		// bits := math.Float32bits(float32(v))
		// binary.LittleEndian.PutUint32(p[i*4:], bits)
	}
	return len(p), nil
}
