package scope

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	emptyImage    = ebiten.NewImage(3, 3)
	emptySubImage = emptyImage.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)
)

type Config struct {
	X            int
	Y            int
	Height       float64
	Width        float64
	Weight       float64
	Color        color.Color
	Src          chan []byte
	ScreenWidth  int
	ScreenHeight int

	BitDepthInBytes int
	NumChannels     int
}

type Scope struct {
	X            int
	Y            int
	Height       float64
	Width        float64
	Weight       float64
	Color        color.Color
	ScreenWidth  int
	ScreenHeight int

	srcFunc func() []float64
	data    []float64
	r       float32
	g       float32
	b       float32
	a       float32
}

func New(c *Config) *Scope {
	s := Scope{
		X:            c.X,
		Y:            c.Y,
		Width:        c.Width,
		Height:       c.Height,
		Weight:       c.Weight,
		Color:        c.Color,
		ScreenWidth:  c.ScreenWidth,
		ScreenHeight: c.ScreenHeight,
	}

	s.srcFunc = func() []float64 {
		select {
		case buf := <-c.Src:
			return transformAudioData(c, buf)
		default:
			return nil
		}
	}

	r, g, b, a := c.Color.RGBA()
	s.r = float32(r)
	s.g = float32(g)
	s.b = float32(b)
	s.a = float32(a)
	s.data = []float64{1, 1, 1, 1, 1}
	emptyImage.Fill(color.White)
	return &s
}

func transformAudioData(cfg *Config, buf []byte) []float64 {
	result := []float64{}
	bytesPerSample := cfg.BitDepthInBytes * cfg.NumChannels
	numSamples := len(buf) / bytesPerSample
	var sampleBuf []byte
	var sample int16
	for sampleIdx := 0; sampleIdx < numSamples; sampleIdx++ {
		for channelIdx := 0; channelIdx < cfg.NumChannels; channelIdx += 2 {
			idx := (bytesPerSample * sampleIdx) + (channelIdx * cfg.BitDepthInBytes)
			sampleBuf = []byte{buf[idx], buf[idx+1]}

			buf := bytes.NewReader(sampleBuf)
			err := binary.Read(buf, binary.LittleEndian, &sample)
			if err != nil {
				fmt.Println("binary.Read failed:", err)
			}
			result = append(result, float64(sample)/32767.0)
		}
	}
	return result
}

func (e *Scope) Update() error {
	return nil
}

func (e *Scope) Draw(screen *ebiten.Image) {
	var path vector.Path
	yOrigin := float32(e.Y) + float32(e.Height/2.0)

	newData := e.srcFunc()
	if len(newData) > 0 {
		e.data = newData
	}
	sampleRate := float64(len(e.data)) / float64(e.Width)
	yScale := (float32(e.Height / 2.0)) - float32(e.Weight)

	rects := []*rect{}
	for i := 0; i < int(e.Width); i++ {
		idx := int(math.Floor(float64(i) * sampleRate))
		idx2 := int(math.Floor(float64(i+1) * sampleRate))
		if i+1 >= int(e.Width) {
			break
		}
		srcX := float32(i + e.X)
		srcY := yOrigin + (float32(e.data[idx]) * float32(yScale))

		destX := float32(i + 1 + e.X)
		destY := yOrigin + float32(e.data[idx2])*float32(yScale)
		r := rectFromLine(srcX, srcY, destX, destY, yOrigin, e.Weight)
		// Draw the tops of each rectangle as we iterate from left-to-right
		if i == 0 {
			path.MoveTo(r.tL.x, r.tL.y)
		} else {
			path.LineTo(r.tL.x, r.tL.y)
		}
		path.LineTo(r.tR.x, r.tR.y)
		rects = append(rects, r)
	}
	// Then circle back and draw the bottoms of each rectangle from right-to-left to create one polygon
	for i := len(rects) - 1; i >= 0; i-- {
		r := rects[i]
		path.LineTo(r.bR.x, r.bR.y)
		path.LineTo(r.bL.x, r.bL.y)
	}

	op := &ebiten.DrawTrianglesOptions{
		FillRule: ebiten.EvenOdd,
	}
	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = e.r
		vs[i].ColorG = e.g
		vs[i].ColorB = e.b
	}

	screen.DrawTriangles(vs, is, emptySubImage, op)
}

func rectFromLine(srcX, srcY, destX, destY, yOrigin float32, weight float64) *rect {
	var ax float32
	var ay float32
	var m float32
	if srcX == destX {
		// vertical line
		ax = float32(weight)
		ay = 0
	} else if srcY == destY {
		// horizontal
		ax = 0
		ay = float32(weight)
	} else {
		m = float32(destY-srcY) / float32(destX-srcX)
		am := 1 / m
		theta := math.Atan(float64(am))
		ax = float32(math.Cos(theta) * weight)
		ay = float32(math.Sin(theta) * weight)
	}
	// source is always the left because we are drawing left-to-right
	return &rect{
		bL: newPt(srcX-(0.5*ax), srcY+(0.5*ay)),
		tL: newPt(srcX+(0.5*ax), srcY-(0.5*ay)),
		bR: newPt(destX-(0.5*ax), destY+(0.5*ay)),
		tR: newPt(destX+(0.5*ax), destY-(0.5*ay)),
	}
}

func (e *Scope) Layout(outsideWidth, outsideHeight int) (int, int) {
	return e.ScreenWidth, e.ScreenHeight
}
