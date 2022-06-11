package synth

import (
	"io"
	"runtime"

	"github.com/hajimehoshi/oto/v2"
	"github.com/pkg/errors"
)

type Closer func()

type Config struct {
	SampleRate      int
	NumChannels     int
	BitDepthInBytes int
}

func NewPlayer(cfg *Config, r io.Reader) (oto.Player, Closer, error) {
	ctx, ready, err := oto.NewContext(cfg.SampleRate, cfg.NumChannels, cfg.BitDepthInBytes)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error creating player")
	}
	<-ready

	p := ctx.NewPlayer(r)
	p.(oto.BufferSizeSetter).SetBufferSize(1024 * cfg.NumChannels * cfg.BitDepthInBytes) // 2048
	closer := func() {
		runtime.KeepAlive(p)
	}
	return p, closer, nil
}
