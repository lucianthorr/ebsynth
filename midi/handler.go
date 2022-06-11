package midi

import (
	"fmt"
	"log"

	"github.com/rakyll/portmidi"
)

type Handler interface {
	GetEvents() []portmidi.Event
}

type HandlerFunc func() []portmidi.Event // pulls and returns a list of midi events

func (hf HandlerFunc) GetEvents() []portmidi.Event {
	return hf()
}

func (hf HandlerFunc) Monitor() {
	for _, e := range hf.GetEvents() {
		fmt.Printf("ts: %d\tstatus: %d\tdata1: %d\tdata2: %d\n", e.Timestamp, e.Status, e.Data1, e.Data2)
	}
}

// builds a function to poll midi events
func GetHandler(d *Driver) HandlerFunc {
	return func() []portmidi.Event {
		res, err := d.stream.Poll()
		if err != nil {
			log.Fatal(fmt.Errorf("Error polling: %s", err.Error()))
		}
		filteredEvents := []portmidi.Event{}
		if res {
			events, err := d.stream.Read(1024)
			if err != nil {
				log.Fatal(fmt.Errorf("Error reading: %s", err.Error()))
			}
			for i := range events {
				if 0x08 <= events[i].Status&0xF0 && events[i].Status&0xF0 < 0xF0 {
					// filters out sysex and system real time messages
					filteredEvents = append(filteredEvents, events[i])
				}
			}
		}
		return filteredEvents
	}
}
