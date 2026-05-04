package main

import (
	"encoding/binary"
	"math"
	"time"

	"github.com/ebitengine/oto/v3"
)

const sampleRate int = 44100

func main() {
	op := &oto.NewContextOptions{
		SampleRate:   sampleRate,
		ChannelCount: 1,
		Format:       oto.FormatFloat32LE,
	}
	otoCtx, ready, _ := oto.NewContext(op)
	<-ready

	// Redosled tonova: Frekvencija, Glasnoća (0.0 do 1.0), Trajanje
	playTone(otoCtx, 440.0, 0.1, 3*time.Second)        // Tih A4
	playTone(otoCtx, 880.0, 0.5, 500*time.Millisecond) // Glasniji A5
}

func playTone(ctx *oto.Context, freq float64, volume float64, duration time.Duration) {
	// Sada koristimo ctx.SampleRate() - to je metoda, ne polje
	osc := &sineWave{
		freq:   freq,
		sample: 0,
	}

	p := ctx.NewPlayer(osc)

	// Koristimo ugrađenu metodu za glasnoću
	p.SetVolume(volume)

	p.Play()

	time.Sleep(duration)
	p.Pause()
}

type sineWave struct {
	freq   float64
	sample int
}

func (s *sineWave) Read(p []byte) (n int, err error) {
	for i := 0; i < len(p)/4; i++ {
		v := math.Sin(2 * math.Pi * s.freq * float64(s.sample) / float64(sampleRate))
		s.sample++

		bits := math.Float32bits(float32(v))
		binary.LittleEndian.PutUint32(p[i*4:], bits)
	}
	return len(p), nil
}
