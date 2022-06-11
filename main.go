package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"

	"os"
	"os/signal"
	"syscall"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lucianthorr/ebsynth/midi"
	"github.com/lucianthorr/ebsynth/scope"
	"github.com/lucianthorr/ebsynth/synth"
	"github.com/rakyll/portmidi"
)

const (
	screenWidth  = 640
	screenHeight = 640
)

var (
	listFlag    = flag.Bool("ls", false, "list available input devices")
	monitorFlag = flag.Bool("m", false, "run a simple midi monitor")
	deviceFlag  = flag.Int("d", -1, "device to listen")
)

func main() {
	flag.Parse()

	driver, err := midi.New()
	if err != nil {
		log.Fatal(err)
	}
	defer driver.Close()
	if *listFlag {
		if err := driver.ListInputs(); err != nil {
			log.Fatal(err)
		}
		return
	}
	if 0 <= *deviceFlag && *deviceFlag < portmidi.CountDevices()-1 {
		if err := driver.Start(*deviceFlag); err != nil {
			log.Fatal(err)
		}
		midiHandler := midi.GetHandler(driver)
		midiAdapter := synth.MidiAdapter(midiHandler)
		if *monitorFlag {
			midiHandler.Monitor()
		}

		cfg := &synth.Config{
			SampleRate:      48000,
			NumChannels:     2,
			BitDepthInBytes: 2, // 16-bit
		}

		dest := make(chan []byte, 1)
		gen := synth.SineGen(cfg, midiAdapter).WithListener(dest)

		p, closer, err := synth.NewPlayer(cfg, gen)
		if err != nil {
			log.Fatal(err)
		}
		defer closer()
		p.Play()

		ebiten.SetWindowSize(screenWidth, screenHeight)
		ebiten.SetWindowTitle("Static Sine Wave Demo")
		scope := scope.New(&scope.Config{
			X:               5,
			Y:               5,
			Width:           screenWidth - 10,
			Height:          screenHeight - 10,
			Color:           color.White,
			Weight:          5,
			Src:             dest,
			ScreenWidth:     screenWidth,
			ScreenHeight:    screenHeight,
			NumChannels:     2,
			BitDepthInBytes: 2,
		})

		if err := ebiten.RunGame(scope); err != nil {
			log.Fatal(err)
		}

		wait := make(chan os.Signal, 1)
		signal.Notify(wait, os.Interrupt, syscall.SIGTERM)
		<-wait

	} else if *monitorFlag {
		if err := driver.ListInputs(); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Specify an input device to monitor")
	}
}
