package main

import (
	"fmt"
	"math"
	"sync"
)

const (
	SampleRate  = 44100
	NumChannels = 2
	BitDepth    = 16

	// Yamaha PSR-E363 — 61 keys: C2 (MIDI 36) to C7 (MIDI 96)
	KeyboardLowest  = 36
	KeyboardHighest = 96

	// Highest octave: C6 (84) to C7 (96) — used as effect trigger keys, not played
	HighestOctaveLow  = 84
	HighestOctaveHigh = 96

	MaxVoices = 8
)

// effectKeyMap maps specific high-octave MIDI notes to a Filter factory.
// Press the key → activate that effect. Press again → back to bypass.
var effectKeyMap = map[uint8]func() Filter{
	84: func() Filter { return NewLowPassFilter(800, SampleRate) }, // C6
	86: func() Filter { return NewDistortionFilter(6) },            // D6
	88: func() Filter { return NewRingMod(110, SampleRate) },       // E6
	96: func() Filter { return &BypassFilter{} },                   // F6  (explicit bypass)
}

var effectKeyNames = map[uint8]string{
	84: "Low-Pass Filter",
	86: "Distortion",
	88: "Ring Modulator",
	96: "Bypass",
}

func isHighestOctave(note uint8) bool {
	return note >= HighestOctaveLow && note <= HighestOctaveHigh
}

func midiNoteToFreq(note uint8) float64 {
	return 440.0 * math.Pow(2, (float64(note)-69)/12)
}

// ─── ADSR ────────────────────────────────────────────────────────────────────

type EnvState int

const (
	Idle EnvState = iota
	Attack
	Decay
	Sustain
	Release
)

type ADSR struct {
	state                   EnvState
	value                   float64
	attack, decay, release  float64
	sustain                 float64
	sampleRate              float64
	releaseStartVal         float64
}

func NewADSR(sr float64) *ADSR {
	return &ADSR{attack: 0.01, decay: 0.1, sustain: 0.7, release: 0.3, sampleRate: sr}
}

func (e *ADSR) NoteOn() { e.state = Attack }
func (e *ADSR) NoteOff() {
	e.releaseStartVal = e.value
	e.state = Release
}
func (e *ADSR) IsIdle() bool { return e.state == Idle }

func (e *ADSR) Tick() float64 {
	switch e.state {
	case Attack:
		e.value += 1 / (e.attack * e.sampleRate)
		if e.value >= 1 {
			e.value = 1
			e.state = Decay
		}
	case Decay:
		e.value -= (1 - e.sustain) / (e.decay * e.sampleRate)
		if e.value <= e.sustain {
			e.value = e.sustain
			e.state = Sustain
		}
	case Sustain:
		e.value = e.sustain
	case Release:
		e.value -= e.releaseStartVal / (e.release * e.sampleRate)
		if e.value <= 0 {
			e.value = 0
			e.state = Idle
		}
	}
	return e.value
}

// ─── Oscillator ───────────────────────────────────────────────────────────────

type Oscillator struct {
	freq, sampleRate, phase float64
}

func NewOscillator(freq, sr float64) *Oscillator {
	return &Oscillator{freq: freq, sampleRate: sr}
}

func (o *Oscillator) Tick() float64 {
	// Additive: fundamental + harmonics → piano-ish timbre
	s := math.Sin(2*math.Pi*o.phase)*1.0 +
		math.Sin(4*math.Pi*o.phase)*0.5 +
		math.Sin(6*math.Pi*o.phase)*0.25 +
		math.Sin(8*math.Pi*o.phase)*0.12 +
		math.Sin(10*math.Pi*o.phase)*0.06
	s /= 1.93
	o.phase += o.freq / o.sampleRate
	if o.phase >= 1 {
		o.phase -= 1
	}
	return s
}

// ─── Voice ────────────────────────────────────────────────────────────────────

type Voice struct {
	note   uint8
	osc    *Oscillator
	env    *ADSR
	active bool
	vel    float64
}

func NewVoice(sr float64) *Voice { return &Voice{env: NewADSR(sr)} }

func (v *Voice) NoteOn(note, velocity uint8) {
	v.note = note
	v.vel = float64(velocity) / 127.0
	v.osc = NewOscillator(midiNoteToFreq(note), SampleRate)
	v.env.NoteOn()
	v.active = true
}

func (v *Voice) NoteOff() { v.env.NoteOff() }

func (v *Voice) Tick() float64 {
	if !v.active {
		return 0
	}
	env := v.env.Tick()
	if v.env.IsIdle() {
		v.active = false
		return 0
	}
	return v.osc.Tick() * env * v.vel
}

// ─── Synth ────────────────────────────────────────────────────────────────────

type Synth struct {
	mu            sync.Mutex
	voices        [MaxVoices]*Voice
	activeFilter  Filter
	activeKey     uint8  // which effect key is currently on (0 = none / bypass)
}

func NewSynth() *Synth {
	s := &Synth{activeFilter: &BypassFilter{}}
	for i := range s.voices {
		s.voices[i] = NewVoice(SampleRate)
	}
	return s
}

// HandleNoteOn processes a note-on. Returns true if the note was an effect key.
func (s *Synth) HandleNoteOn(note, velocity uint8) bool {
	if factory, ok := effectKeyMap[note]; ok {
		s.mu.Lock()
		defer s.mu.Unlock()

		if s.activeKey == note {
			// Same key pressed again → toggle off (bypass)
			s.activeFilter = &BypassFilter{}
			s.activeKey = 0
			fmt.Printf("\n→ Effect OFF (Bypass)\n\n")
		} else {
			s.activeFilter = factory()
			s.activeKey = note
			fmt.Printf("\n→ Effect ON: %s\n\n", effectKeyNames[note])
		}
		return true
	}
	// Regular note in lower octaves
	s.mu.Lock()
	defer s.mu.Unlock()
	target := s.voices[0]
	for _, v := range s.voices {
		if !v.active {
			target = v
			break
		}
	}
	target.NoteOn(note, velocity)
	return false
}

func (s *Synth) NoteOff(note uint8) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, v := range s.voices {
		if v.active && v.note == note {
			v.NoteOff()
		}
	}
}

func (s *Synth) ActiveEffectName() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.activeFilter.Name()
}

// Generate fills buf with interleaved int16 stereo PCM.
func (s *Synth) Generate(buf []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	frames := len(buf) / (NumChannels * 2)
	for i := 0; i < frames; i++ {
		var raw float64
		for _, v := range s.voices {
			if v.active {
				raw += v.Tick()
			}
		}

		// Apply the active DSP filter to the mix
		raw = s.activeFilter.Process(raw)

		// Drive hard into tanh so the limiter is audible and loud
		raw = math.Tanh(raw * 4.0)

		// Float [-1, 1] → int16, near full scale
		sample16 := int16(raw * 31000)
		lo := byte(uint16(sample16))
		hi := byte(uint16(sample16) >> 8)

		base := i * 4
		buf[base+0], buf[base+1] = lo, hi // L
		buf[base+2], buf[base+3] = lo, hi // R
	}
}
