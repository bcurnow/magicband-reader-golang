package audio

import (
	"github.com/faiface/beep"
	"os"

	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

const (
	defaultBase = 2
)

type Controller struct {
	sampleRate beep.SampleRate
	Volume     float64
	Base       float64
}

func (c *Controller) Load(soundFile string) (beep.Streamer, error) {
	f, err := os.Open(soundFile)
	if err != nil {
		return nil, err
	}

	// For the beep library, they don't recommend closing the file yourself
	// streamer.Close() (below) will take care of it
	// defer f.Close()

	wavStreamer, format, err := wav.Decode(f)
	if err != nil {
		return nil, err
	}
	defer wavStreamer.Close()

	var streamer beep.Streamer = wavStreamer

	if c.sampleRate != 0 {
		// We've already played at least one file
		// This means the speaker has already been initialized
		if c.sampleRate != format.SampleRate {
			streamer = beep.Resample(4, format.SampleRate, c.sampleRate, streamer)
		}
	} else {
		c.sampleRate = format.SampleRate

		// The examples show using the sample rate to determine the size of the buffer
		// Because we need to support multiple, potentially user provided sound files
		// We're going to choose a default buffer size (1K)  that should be sufficient
		// to stream any of the files.
		speaker.Init(format.SampleRate, 1024)
	}

	// Read the file into memory, these are very small files
	buffer := beep.NewBuffer(format)
	buffer.Append(streamer)

	streamer = buffer.Streamer(0, buffer.Len())

	volume := &effects.Volume{
		Streamer: streamer,
		Base:     c.Base,
		Volume:   c.Volume,
		Silent:   false,
	}
	return volume, nil
}

func (c *Controller) Play(streamer beep.Streamer) {
	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))
	<-done
}

func (c *Controller) Init() error {
	c.handleDefaults()
	return nil
}

func (c *Controller) handleDefaults() {
	if c.Base == 0 {
		c.Base = defaultBase
	}
}
