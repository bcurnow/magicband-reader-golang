package audio

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"

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
	AuthorizedSound() string
	ReadSound() string
	UnauthorizedSound() string
}

type controller struct {
	cache             Cache
	sampleRate        beep.SampleRate
	volume            float64
	base              float64
	authorizedSound   string
	readSound         string
	unauthorizedSound string
}

func NewController(volume float64, base float64, cache Cache, authorizedSound string, readSound string, unauthorizedSound string) (Controller, error) {
	log.Trace("Creating new audio.Controller")

	c := controller{
		cache:             cache,
		volume:            volume,
		base:              base,
		authorizedSound:   authorizedSound,
		readSound:         readSound,
		unauthorizedSound: unauthorizedSound,
	}

	if err := c.validateSoundConfig(cache.CacheDir()); err != nil {
		return nil, err
	}

	c.handleDefaults()
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

func (c *controller) AuthorizedSound() string {
	return c.authorizedSound
}

func (c *controller) ReadSound() string {
	return c.readSound
}

func (c *controller) UnauthorizedSound() string {
	return c.unauthorizedSound
}

func (c *controller) handleDefaults() {
	if c.base == 0 {
		c.base = defaultBase
	}
}

func (c *controller) validateSoundConfig(soundDir string) error {
	if err := validateFileExists(soundDir, "sound-dir"); err != nil {
		return err
	}

	if err := validateFileExists(path.Join(soundDir, c.authorizedSound), "authorized-sound"); err != nil {
		return err
	}

	if err := validateFileExists(path.Join(soundDir, c.readSound), "read-sound"); err != nil {
		return err
	}

	if err := validateFileExists(path.Join(soundDir, c.unauthorizedSound), "unauthorized-sound"); err != nil {
		return err
	}
	return nil
}

func validateFileExists(file string, name string) error {
	if _, err := os.Stat(file); errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("Invalid value for %v: '%v'. %v", name, file, err)
	}
	return nil
}
