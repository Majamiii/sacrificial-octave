package main

import (
	"fmt"
	"math"
	"time"

	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // Inicijalizuje drajvere za Windows/Mac/Linux
	// go get gitlab.com/gomidi/midi/v2
	// go get gitlab.com/gomidi/midi/v2/drivers/rtmididrv
)

const A4 = 440.0
const A4MIDI = 69.0

func MidiToFreq(midi uint8) float64 {
	// return 440.0 * math.Pow(2.0, float64(midi-69)/12.0)
	return A4 * math.Pow(2, (float64(midi)-A4MIDI)/12)
}

func main() {
	// 1. Uvek prvo zatvori MIDI drajvere na kraju programa
	defer midi.CloseDriver()

	// 2. Izlistaj sve dostupne MIDI ulaze - sta je sve detektovano
	fmt.Println("Dostupni MIDI ulazi:")
	inPorts := midi.GetInPorts()
	for _, port := range inPorts {
		fmt.Printf("[%d] -> % s\n", port.Number(), port.String())
	}

	if len(inPorts) == 0 {
		fmt.Println("Nije pronađena nijedna klavijatura!")
		return
	}

	// 3. Odaberi prvi slobodan port (obično tvoja klavijatura)
	in := inPorts[0]
	fmt.Printf("\nSlušam na portu: %s\n", in.String())

	fmt.Println("Pritisni dirke na klavijaturi")

	// 4. Pokreni slušanje
	stop, err := midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		var channel, note, velocity uint8

		switch {
		case msg.GetNoteOn(&channel, &note, &velocity):
			fmt.Printf("[Note ON]  Nota: %d | Frekvencija: %f | Jačina: %d | Vreme: %dms\n", note, MidiToFreq(note), velocity, timestampms)

		case msg.GetNoteOff(&channel, &note, &velocity):
			fmt.Printf("[Note OFF] Nota: %d | Vreme: %dms\n", note, timestampms)
		}
	})

	if err != nil {
		fmt.Printf("Greška pri slušanju: %v\n", err)
		return
	}

	// 5. Drži program budnim
	for {
		time.Sleep(time.Second)
	}

	// Pozovi stop() ako ikada budeš htela da prekineš slušanje programski
	_ = stop
}
