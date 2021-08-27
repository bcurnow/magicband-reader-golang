package handler

import (
	"path/filepath"

	"github.com/faiface/beep"
	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/config"
	"github.com/bcurnow/magicband-reader/context"
	"github.com/bcurnow/magicband-reader/event"
)

type ReadSound struct {
	sound beep.Streamer
}

func (h *ReadSound) Handle(e event.Event) error {
	log.Trace("Playing the read sound")
	return runAsync("readSoundPlaying", func() {
		context.AudioController.Play(h.sound)
	})
}

func init() {
	sound, err := context.AudioController.Load(filepath.Join(config.SoundDir, config.ReadSound))
	if err != nil {
		panic(err)
	}

	if err := context.RegisterHandler(10, &ReadSound{
		sound: sound,
	}); err != nil {
		panic(err)
	}
}
