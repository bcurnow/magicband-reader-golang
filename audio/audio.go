package audio

import (
	"os"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	log "github.com/sirupsen/logrus"
)

const (
	defaultBase = 2
)

type Controller interface {
	Load(soundFile string) (beep.Streamer, error)
	Play(streamer beep.Streamer)
}

type controller struct {
	sampleRate beep.SampleRate
	Volume     float64
	Base       float64
}

func NewController(volume float64, base float64) (*controller, error) {
	log.Trace("Creating new audio.Controller")
	c := controller{
		Volume: volume,
		Base:   base,
	}
	c.handleDefaults()
	return &c, nil
}

func (c *controller) Load(soundFile string) (beep.Streamer, error) {
	f, err := os.Open(soundFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

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
		// We're going to choose a default buffer size (5K)  that should be sufficient
		// to stream any of the files.
		speaker.Init(format.SampleRate, 5*1024)
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

func (c *controller) Play(streamer beep.Streamer) {
	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		close(done)
	})))
	<-done
}

func (c *controller) handleDefaults() {
	if c.Base == 0 {
		c.Base = defaultBase
	}
}
