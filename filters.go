package main

import "math"

// Filter is the interface all DSP processors implement.
type Filter interface {
	Process(sample float64) float64
	Reset()
	Name() string
}

// ─── Low-Pass Filter (biquad, Butterworth) ───────────────────────────────────

type LowPassFilter struct {
	cutoff     float64 // normalized: 0–1  (1 = Nyquist)
	sampleRate float64

	// biquad coefficients
	b0, b1, b2 float64
	a1, a2     float64

	// delay lines
	x1, x2, y1, y2 float64
}

func NewLowPassFilter(cutoffHz, sampleRate float64) *LowPassFilter {
	f := &LowPassFilter{cutoff: cutoffHz, sampleRate: sampleRate}
	f.calcCoeffs()
	return f
}

func (f *LowPassFilter) calcCoeffs() {
	// Butterworth 2nd-order low-pass
	w0 := 2 * math.Pi * f.cutoff / f.sampleRate
	cosW := math.Cos(w0)
	sinW := math.Sin(w0)
	alpha := sinW / (2 * 0.7071) // Q = 1/√2 (maximally flat)

	b0 := (1 - cosW) / 2
	b1 := 1 - cosW
	b2 := (1 - cosW) / 2
	a0 := 1 + alpha
	a1 := -2 * cosW
	a2 := 1 - alpha

	f.b0 = b0 / a0
	f.b1 = b1 / a0
	f.b2 = b2 / a0
	f.a1 = a1 / a0
	f.a2 = a2 / a0
}

func (f *LowPassFilter) Process(x float64) float64 {
	y := f.b0*x + f.b1*f.x1 + f.b2*f.x2 - f.a1*f.y1 - f.a2*f.y2
	f.x2 = f.x1
	f.x1 = x
	f.y2 = f.y1
	f.y1 = y
	return y
}

func (f *LowPassFilter) Reset() { f.x1, f.x2, f.y1, f.y2 = 0, 0, 0, 0 }
func (f *LowPassFilter) Name() string { return "Low-Pass (Butterworth, 800 Hz)" }

// ─── Distortion (soft-clip waveshaper) ───────────────────────────────────────

type DistortionFilter struct {
	drive float64 // 1.0 = mild, 10.0 = heavy
}

func NewDistortionFilter(drive float64) *DistortionFilter {
	return &DistortionFilter{drive: drive}
}

// softClip uses tanh waveshaping — sounds warm, no harsh aliasing
func softClip(x, drive float64) float64 {
	return math.Tanh(x * drive)
}

func (d *DistortionFilter) Process(sample float64) float64 {
	return softClip(sample, d.drive) / math.Tanh(d.drive)
}
func (d *DistortionFilter) Reset() {}
func (d *DistortionFilter) Name() string { return "Distortion (tanh soft-clip, drive=6)" }

// ─── Ring Modulator ───────────────────────────────────────────────────────────

type RingMod struct {
	modFreq    float64
	sampleRate float64
	phase      float64
}

func NewRingMod(modFreq, sampleRate float64) *RingMod {
	return &RingMod{modFreq: modFreq, sampleRate: sampleRate}
}

func (r *RingMod) Process(sample float64) float64 {
	out := sample * math.Sin(2*math.Pi*r.phase)
	r.phase += r.modFreq / r.sampleRate
	if r.phase >= 1 {
		r.phase -= 1
	}
	return out * 2.0 // makeup gain: AM modulation halves RMS amplitude
}
func (r *RingMod) Reset() { r.phase = 0 }
func (r *RingMod) Name() string { return "Ring Modulator (110 Hz carrier)" }

// ─── Bypass (no effect) ──────────────────────────────────────────────────────

type BypassFilter struct{}

func (b *BypassFilter) Process(sample float64) float64 { return sample }
func (b *BypassFilter) Reset()                         {}
func (b *BypassFilter) Name() string                   { return "Bypass (no effect)" }
