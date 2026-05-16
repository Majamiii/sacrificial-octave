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

	p1 := createPlayer(otoCtx, 261.63, 0.5, "sine")   // Čist ton C4
	p2 := createPlayer(otoCtx, 261.63, 0.2, "square") // 8-bit retro zvuk
	p3 := createPlayer(otoCtx, 261.63, 0.5, "dist")   // Agresivan, pojačan ton

	p1.Play()
	time.Sleep(2 * time.Second)
	p1.Pause()
	p2.Play()
	time.Sleep(2 * time.Second)
	p2.Pause()

	p3.Play()
	time.Sleep(2 * time.Second)
	p3.Pause()
}

func createPlayer(ctx *oto.Context, freq float64, volume float64, waveType string) *oto.Player {
	osc := &oscillator{
		freq:     freq,
		sample:   0,
		waveType: waveType,
	}

	p := ctx.NewPlayer(osc)
	p.SetVolume(volume)
	return p
}

type oscillator struct {
	freq     float64
	sample   int
	waveType string // "sine", "square", "dist"
}

func (o *oscillator) Read(p []byte) (n int, err error) {
	for i := 0; i < len(p)/4; i++ {
		// 1. Prvo izračunamo osnovni sinusni talas kao bazu
		val := math.Sin(2 * math.Pi * o.freq * float64(o.sample) / float64(sampleRate))
		o.sample++

		// 2. Switch-case menja taj talas pre slanja na zvučnike
		var finalOut float64

		switch o.waveType {
		case "square":
			// Square wave: ako je sinus > 0 vrati 1, inače vrati -1
			if val >= 0 {
				finalOut = 1.0
			} else {
				finalOut = -1.0
			}

		case "dist":
			// Distorzija: Pomnožimo signal i "isečemo" ga na 0.7
			// Ovo simulira preopterećenje pojačala
			const gain = 2.0
			const limit = 0.7
			finalOut = val * gain
			if finalOut > limit {
				finalOut = limit
			} else if finalOut < -limit {
				finalOut = -limit
			}

		case "sine":
			fallthrough
		default:
			// Čist sinusni talas
			finalOut = val
		}

		bits := math.Float32bits(float32(finalOut))
		binary.LittleEndian.PutUint32(p[i*4:], bits)
	}
	return len(p), nil
}
