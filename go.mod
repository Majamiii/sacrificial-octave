module midi-dsp

go 1.22.2

require golang.org/x/sys v0.7.0 // indirect

require (
	github.com/hajimehoshi/oto/v2 v2.4.2
	gitlab.com/gomidi/midi/v2 v2.2.10
)

require github.com/ebitengine/purego v0.4.1 // indirect

// winmididrv uses Windows Multimedia (WinMM) — no extra system libs needed.
// rtmididrv (ALSA) will NOT work in WSL2 since it has no kernel sound subsystem.
