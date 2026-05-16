package main

import (
	"fmt"
	"math"
	"time"

	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

var midiFreqTable [128]float64

func BuildMidiFreqTable() {
	for i := 0; i < 128; i++ {
		midiFreqTable[i] = 440.0 * math.Pow(2.0, float64(i-69)/12.0)
	}
}

func main() {
	defer midi.CloseDriver()

	// napravimo tabelu koja mapira midi notu sa frekkvencijom
	BuildMidiFreqTable()

	inPorts := midi.GetInPorts()

	if len(inPorts) == 0 {
		fmt.Println("Nije pronađena nijedna klavijatura!")
		return
	}

	in := inPorts[0]

	fmt.Println("Pritisni dirke na klavijaturi")

	stop, err := midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		var channel, note, velocity uint8

		switch {
		case msg.GetNoteOn(&channel, &note, &velocity):
			fmt.Printf("[Note ON]  Nota: %d | Frekvencija: %f | Jačina: %d | Vreme: %dms\n", note, midiFreqTable[note], velocity, timestampms)

		case msg.GetNoteOff(&channel, &note, &velocity):
			fmt.Printf("[Note OFF] Nota: %d | Vreme: %dms\n", note, timestampms)
		}
	})

	if err != nil {
		fmt.Printf("Greška pri slušanju: %v\n", err)
		return
	}

	for {
		time.Sleep(time.Second)
	}

	_ = stop
}
