package handler

import (
	"github.com/faiface/beep"
	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/context"
	"github.com/bcurnow/magicband-reader/event"
)

type AuthSound struct {
	authSound   *beep.Buffer
	unauthSound *beep.Buffer
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
	authSound, err := context.AudioController.Load(context.AudioController.AuthorizedSound())
	if err != nil {
		panic(err)
	}

	unauthSound, err := context.AudioController.Load(context.AudioController.UnauthorizedSound())
	if err != nil {
		panic(err)
	}

	handler := &AuthSound{
		authSound:   authSound,
		unauthSound: unauthSound,
	}

	if err := context.RegisterHandler(21, handler); err != nil {
		panic(err)
	}
}
