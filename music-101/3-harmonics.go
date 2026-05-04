package main

import (
	"encoding/binary"
	"math"
	"time"

	"github.com/ebitengine/oto/v3"
)

const sampleRate int = 44100

func main() {
	const sampleRate = 44100
	op := &oto.NewContextOptions{
		SampleRate:   sampleRate,
		ChannelCount: 1,
		Format:       oto.FormatFloat32LE,
	}
	otoCtx, ready, _ := oto.NewContext(op)
	<-ready

	// --- SIMULTANO PUŠTANJE (AKORD) ---

	// Pravimo tri plejera, ali ih ne blokiramo sa Sleep-om odmah
	p1 := createPlayer(otoCtx, 261.63, 0.2) // C4
	p2 := createPlayer(otoCtx, 329.63, 0.2) // E4
	p3 := createPlayer(otoCtx, 392.00, 0.2) // G4

	// Svi počinju da sviraju skoro u istom trenutku
	p1.Play()
	p2.Play()
	p3.Play()

	// Čekamo 2 sekunde dok sva tri plejera rade "u pozadini"
	time.Sleep(2 * time.Second)

	// Gasimo ih sve
	p1.Pause()
	p2.Pause()
	p3.Pause()
}

// Funkcija sada samo PRAVI plejer, ali ga ne pokreće i ne čeka
func createPlayer(ctx *oto.Context, freq float64, volume float64) *oto.Player {
	osc := &sineWave{
		freq:   freq,
		sample: 0,
	}

	p := ctx.NewPlayer(osc)
	p.SetVolume(volume)
	return p
}

// --- Struktura sineWave i Read metoda ostaju iste ---
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
