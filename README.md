# 🎹 Sacrificial Octave
## PSR-E363 Go MIDI DSP Synth

A real-time MIDI synthesizer written in Go that reads input from a Yamaha PSR-E363 keyboard, synthesizes audio, and applies DSP effects triggered by keys in the highest octave.

Built as a project for [Petnica Science Center](https://petnica.rs), summer 2026.

---

## How it works

The keyboard sends MIDI messages over USB. Go reads them, synthesizes audio using a polyphonic oscillator engine, runs the signal through a DSP filter, and outputs it to your speakers — all in real time.

```
Keyboard (USB MIDI)
       ↓
  MIDI Listener        reads NoteOn / NoteOff events
       ↓
  Synth Engine         8 voices, ADSR envelopes, additive oscillator
       ↓
  DSP Filter           Low-Pass / Distortion / Ring Modulator / Bypass
       ↓
  Audio Output         int16 PCM → sound card via oto
```

The highest octave (C6–C7, MIDI 84–96) is "sacrificed" — those keys don't play notes. Instead they silently toggle DSP effects that apply to everything played in the lower octaves.

---

## Effect keys

| Key | MIDI | Effect |
|-----|------|--------|
| C6  | 84   | Low-Pass Filter — warm, muffled tone (Butterworth, 800 Hz cutoff) |
| D6  | 86   | Distortion — soft-clip overdrive (tanh waveshaper, drive = 6) |
| E6  | 88   | Ring Modulator — metallic, alien sound (110 Hz carrier) |
| F6  | 89   | Bypass — no effect |

Press a key to activate its effect. Press the same key again to toggle it off (back to bypass). Only one effect is active at a time.

---

## Project structure

```
.
├── main.go          # Entry point: MIDI listener + audio output glue
├── synth.go         # Polyphonic synth engine (voices, ADSR, oscillator)
├── filters.go       # DSP effects (Low-Pass, Distortion, Ring Mod, Bypass)
└── go.mod           # Dependencies
```

### `filters.go`
Defines a `Filter` interface and four implementations. Each filter receives one audio sample at a time and returns a transformed sample. The low-pass filter uses a 2nd-order Butterworth biquad IIR. Distortion uses `tanh()` waveshaping. The ring modulator multiplies the signal by a 110 Hz sine carrier.

### `synth.go`
The audio engine. Manages 8 polyphonic voices, each with an additive oscillator (fundamental + 4 harmonics for a piano-like timbre) and an ADSR amplitude envelope. The `Generate` function is called ~90 times per second by the audio system, filling PCM buffers which are sent to the sound card.

### `main.go`
Glues everything together. Opens the MIDI port, starts the oto audio context, and connects the MIDI callback to the synth. Audio buffer is set to 2048 bytes (~11ms) for low latency.

---

## Requirements

### Hardware
- Yamaha PSR-E363 (or any class-compliant USB MIDI keyboard)
- Connected via USB to Windows

### Software
- Go 1.21+
- A C++ compiler for CGO (needed by `rtmididrv`)

The easiest way to get a compiler on Windows is [MSYS2](https://www.msys2.org):

```bash
# In the MSYS2 terminal:
pacman -S mingw-w64-x86_64-gcc
```

Then add `C:\msys64\mingw64\bin` to your Windows PATH.

> **Important:** Run the program from **PowerShell or cmd.exe**, not from WSL. The keyboard is a Windows USB device and WSL cannot see it.

---

## Build & run

```powershell
# Clone / navigate to the project folder
cd "D:\your\project\folder"

# Download dependencies
go mod tidy

# Run
go run .
```

The program will print available MIDI ports, connect to the first one, and start listening. Play notes in the lower octaves and use the effect keys to switch DSP processing on the fly.

To select a specific MIDI port if you have more than one:

```powershell
$env:MIDI_PORT=1; go run .
```

---

## Adding a new effect

1. Open `filters.go` and implement the `Filter` interface:

```go
type MyEffect struct {
    // your state here
}

func (f *MyEffect) Process(sample float64) float64 {
    // transform sample and return it
    return sample
}
func (f *MyEffect) Reset() {}
func (f *MyEffect) Name() string { return "My Effect" }
```

2. Add it to the `effectKeyMap` in `synth.go`:

```go
var effectKeyMap = map[uint8]func() Filter{
    84: func() Filter { return NewLowPassFilter(800, SampleRate) }, // C6
    86: func() Filter { return NewDistortionFilter(6) },            // D6
    88: func() Filter { return NewRingMod(110, SampleRate) },       // E6
    89: func() Filter { return &BypassFilter{} },                   // F6
    91: func() Filter { return &MyEffect{} },                       // G6 ← new
}
```

---

## Latency tuning

The audio buffer size is set in `main.go`:

```go
player.SetBufferSize(2048) // ~11ms at 44100 Hz
```

| Buffer (bytes) | Latency | Risk |
|----------------|---------|------|
| 1024 | ~6ms | may crackle on slow machines |
| 2048 | ~11ms | good default |
| 4096 | ~23ms | very stable, slightly sluggish feel |

---

## Dependencies

| Package | Purpose |
|---------|---------|
| [`gitlab.com/gomidi/midi/v2`](https://gitlab.com/gomidi/midi) | MIDI message parsing |
| [`gitlab.com/gomidi/midi/v2/drivers/rtmididrv`](https://gitlab.com/gomidi/midi) | MIDI I/O via rtmidi (uses WinMM on Windows) |
| [`github.com/hajimehoshi/oto/v2`](https://github.com/hajimehoshi/oto) | Cross-platform audio output |