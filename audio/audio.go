package audio

import (
	"errors"
	"fmt"
	"io/fs"
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
	Load(soundFile string) (*beep.Buffer, error)
	Play(buffer *beep.Buffer)
	AuthorizedSound() *beep.Buffer
	ReadSound() *beep.Buffer
	UnauthorizedSound() *beep.Buffer
}

type controller struct {
	cache             Cache
	sampleRate        beep.SampleRate
	volume            float64
	base              float64
	authorizedSound   *beep.Buffer
	readSound         *beep.Buffer
	unauthorizedSound *beep.Buffer
}

func NewController(volume float64, base float64, cache Cache, authorizedSound string, readSound string, unauthorizedSound string) (Controller, error) {
	log.Trace("Creating new audio.Controller")

	c := controller{
		cache:  cache,
		volume: volume,
		base:   base,
	}

	if err := c.validateSoundConfig(cache.CacheDir()); err != nil {
		return nil, err
	}

	c.handleDefaults()

	// Pre-load the default sounds
	sound, err := c.Load(authorizedSound)
	if err != nil {
		return nil, err
	}
	c.authorizedSound = sound

	sound, err = c.Load(readSound)
	if err != nil {
		return nil, err
	}
	c.readSound = sound

	sound, err = c.Load(unauthorizedSound)
	if err != nil {
		return nil, err
	}
	c.unauthorizedSound = sound

	return &c, nil
}

func (c *controller) Load(soundFile string) (*beep.Buffer, error) {
	f, err := c.cache.Load(soundFile)
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
	return buffer, nil
}

func (c *controller) Play(buffer *beep.Buffer) {
	streamer := buffer.Streamer(0, buffer.Len())

	volume := &effects.Volume{
		Streamer: streamer,
		Base:     c.base,
		Volume:   c.volume,
		Silent:   false,
	}

	done := make(chan bool)
	speaker.Play(beep.Seq(volume, beep.Callback(func() {
		close(done)
	})))
	<-done
}

func (c *controller) AuthorizedSound() *beep.Buffer {
	return c.authorizedSound
}

func (c *controller) ReadSound() *beep.Buffer {
	return c.readSound
}

func (c *controller) UnauthorizedSound() *beep.Buffer {
	return c.unauthorizedSound
}

func (c *controller) handleDefaults() {
	if c.base == 0 {
		c.base = defaultBase
	}
}

func (c *controller) validateSoundConfig(soundDir string) error {
	if _, err := os.Stat(soundDir); errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("Invalid value sound-dir: '%v'. %v", soundDir, err)
	}

	return nil
}
