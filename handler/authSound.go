package handler

import (
	"github.com/faiface/beep"
	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/context"
	"github.com/bcurnow/magicband-reader/event"
	"github.com/bcurnow/magicband-reader/rfidsecuritysvc"
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
			context.AudioController.Play(h.resolveSound())
		})
	case event.UNAUTHORIZED:
		runAsync("authSoundPlaying", func() {
			context.AudioController.Play(h.unauthSound)
		})
	}
	return nil
}

func (h *AuthSound) resolveSound() *beep.Buffer {
	// Not sure how this would happen but we don't have a MediaConfig object in state
	if context.State["mediaConfig"] == nil {
		log.Errorf("Unable to find mediaConfig in State, using default authorized sound")
		return h.authSound
	}

	mediaConfig := context.State["mediaConfig"].(*rfidsecuritysvc.MediaConfig)
	if mediaConfig.Sound == nil {
		log.Infof("No sound configured in mediaConfig, using default authorized sound")
		return h.authSound
	}

	soundBuffer, err := context.AudioController.Load(mediaConfig.Sound)
	if err != nil {
		log.Errorf("Unable to load %v from mediaConfig, using default authorized sound", mediaConfig.Sound.Name)
		return h.authSound
	}

	return soundBuffer
}

func init() {
	handler := &AuthSound{
		authSound:   context.AudioController.AuthorizedSound(),
		unauthSound: context.AudioController.UnauthorizedSound(),
	}

	if err := context.RegisterHandler(21, handler); err != nil {
		panic(err)
	}
}
