package midi

import (
	"fmt"
	"log"

	"github.com/pkg/errors"
	"github.com/rakyll/portmidi"
)

type Driver struct {
	isInitialized bool
	isOpen        bool

	stream *portmidi.Stream
}

func (d Driver) Close() error {
	if d.isOpen {
		d.stream.Close()
	}
	return portmidi.Terminate()
}

func New() (*Driver, error) {
	if err := portmidi.Initialize(); err != nil {
		return nil, errors.Wrap(err, "Error initializing midi")
	}
	d := Driver{
		isInitialized: true,
	}

	return &d, nil
}

func (d *Driver) ListInputs() error {
	if d.isInitialized {
		for i := 0; i < portmidi.CountDevices(); i++ {
			info := portmidi.Info(portmidi.DeviceID(i))
			if info.IsInputAvailable {
				fmt.Printf("%d: %s\n", i, info.Name)
			}
		}
	} else {
		return fmt.Errorf("Driver must first be initialized")
	}
	return nil
}

func (d *Driver) Start(idx int) (err error) {
	d.stream, err = portmidi.NewInputStream(portmidi.DeviceID(idx), 64)
	if err != nil {
		log.Fatal(fmt.Errorf("Error creating stream: %s", err.Error()))
	}

	d.stream.Listen()
	return nil
}
