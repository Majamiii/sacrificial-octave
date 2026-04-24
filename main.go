package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hajimehoshi/oto/v2"
	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

type audioReader struct{ synth *Synth }

func (a *audioReader) Read(buf []byte) (int, error) {
	a.synth.Generate(buf)
	return len(buf), nil
}

func main() {
	fmt.Println("┌──────────────────────────────────────────────────────┐")
	fmt.Println("│  Yamaha PSR-E363 → Go MIDI DSP Synth                 │")
	fmt.Println("│                                                      │")
	fmt.Println("│  EFFECT KEYS (highest octave — silent triggers):     │")
	fmt.Println("│    C6  (84) → Low-Pass Filter                        │")
	fmt.Println("│    D6  (86) → Distortion                             │")
	fmt.Println("│    E6  (88) → Ring Modulator                         │")
	fmt.Println("│    F6  (89) → Bypass (no effect)                     │")
	fmt.Println("│  Press the same key again to toggle OFF.             │")
	fmt.Println("└──────────────────────────────────────────────────────┘")
	fmt.Println()

	// ── MIDI ports ────────────────────────────────────────────────────────────
	ins := midi.GetInPorts()
	if len(ins) == 0 {
		log.Fatal("No MIDI input ports found.\n" +
			"  • Check USB cable\n" +
			"  • Run from PowerShell/cmd, not WSL")
	}

	fmt.Println("Available MIDI input ports:")
	for i, p := range ins {
		fmt.Printf("  [%d] %s\n", i, p)
	}

	portIdx := 0
	if v := os.Getenv("MIDI_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &portIdx)
	}
	fmt.Printf("\nUsing port [%d]: %s\n\n", portIdx, ins[portIdx])

	// ── Synth ─────────────────────────────────────────────────────────────────
	synth := NewSynth()

	// ── Audio output ──────────────────────────────────────────────────────────
	ctx, ready, err := oto.NewContext(SampleRate, NumChannels, 2)
	if err != nil {
		log.Fatalf("oto init: %v", err)
	}
	<-ready

	player := ctx.NewPlayer(&audioReader{synth: synth})
	// Small buffer = low latency. 2048 bytes ≈ 11ms at 44100Hz stereo int16.
	// If you hear crackles, bump to 4096. Default is ~500ms which feels broken.
	//player.SetBufferSize(2048)

	if setter, ok := player.(oto.BufferSizeSetter); ok {
		setter.SetBufferSize(4096) // Set to a specific byte size
	}

	player.Play()
	defer player.Close()

	// ── MIDI listener ─────────────────────────────────────────────────────────
	in := ins[portIdx]
	stop, err := midi.ListenTo(in, func(msg midi.Message, _ int32) {
		var ch, key, vel uint8

		switch {
		case msg.GetNoteOn(&ch, &key, &vel):
			if vel == 0 {
				synth.NoteOff(key)
				logNote("OFF", key, vel, false, false)
				return
			}
			isEffectKey := isHighestOctave(key) // any high-octave key is an effect key
			wasHandled := synth.HandleNoteOn(key, vel)
			if !wasHandled {
				logNote("ON ", key, vel, false, false)
			} else {
				logNote("FX ", key, vel, true, synth.activeKey == key)
			}
			_ = isEffectKey

		case msg.GetNoteOff(&ch, &key, &vel):
			synth.NoteOff(key)
			logNote("OFF", key, vel, false, false)
		}
	}, midi.UseSysEx())

	if err != nil {
		log.Fatalf("MIDI listen: %v", err)
	}
	defer stop()

	fmt.Println("Listening… Ctrl+C to quit.")
	fmt.Printf("Active effect: %s\n\n", synth.ActiveEffectName())
	fmt.Printf("%-5s %-6s %-5s %-4s\n", "TYPE", "NOTE", "MIDI", "VEL")
	fmt.Println("─────────────────────────")

	for {
		time.Sleep(time.Hour)
	}
}

var noteNames = [12]string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

func midiNoteName(n uint8) string {
	return fmt.Sprintf("%s%d", noteNames[n%12], int(n/12)-1)
}

func logNote(ev string, note, vel uint8, isFX bool, fxOn bool) {
	tag := ""
	if isFX {
		if fxOn {
			tag = "◀ effect OFF"
		} else {
			tag = "▶ effect ON"
		}
	}
	fmt.Printf("%-5s %-6s %-5d %-4d %s\n", ev, midiNoteName(note), note, vel, tag)
}
