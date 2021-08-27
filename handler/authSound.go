package handler

import (
	"path/filepath"

	"github.com/faiface/beep"
	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/config"
	"github.com/bcurnow/magicband-reader/context"
	"github.com/bcurnow/magicband-reader/event"
)

type AuthSound struct {
	authSound   beep.Streamer
	unauthSound beep.Streamer
}

func (h *AuthSound) Handle(e event.Event) error {
	log.Trace("Playing the auth sound")
	switch e.Type() {
	case event.AUTHORIZED:
		runAsync("authSoundPlaying", func() {
			context.AudioController.Play(h.authSound)
		})
	case event.UNAUTHORIZED:
		runAsync("authSoundPlaying", func() {
			context.AudioController.Play(h.unauthSound)
		})
	}
	return nil
}

func init() {
	authSound, err := context.AudioController.Load(filepath.Join(config.SoundDir, config.AuthorizedSound))
	if err != nil {
		panic(err)
	}

	unauthSound, err := context.AudioController.Load(filepath.Join(config.SoundDir, config.UnauthorizedSound))
	if err != nil {
		panic(err)
	}

	if err := context.RegisterHandler(21, &AuthSound{
		authSound:   authSound,
		unauthSound: unauthSound,
	}); err != nil {
		panic(err)
	}
}
