package synth

import "github.com/lucianthorr/ebsynth/midi"

type Signal struct {
	Freq float64
	Vel  float64
	Gate bool
}

type SignalFunc func() *Signal

func MidiAdapter(handler midi.Handler) SignalFunc {
	note := int64(0)
	velocity := float64(0)
	gate := false
	return func() *Signal {
		events := handler.GetEvents()
		for i := range events {
			if events[i].Status == 0x90 { // NOTE ON
				gate = true
				note = events[i].Data1
				velocity = float64(events[i].Data2) / 128.0
			}
			if events[i].Status == 0x80 { // NOTE OFF
				if events[i].Data1 == note {
					gate = false
					velocity = 0.0
				}
			}
		}
		return &Signal{
			Freq: NOTE_MAP[note],
			Vel:  velocity,
			Gate: gate,
		}
	}
}
