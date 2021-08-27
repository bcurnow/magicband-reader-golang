package handler

import (
	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/context"
	"github.com/bcurnow/magicband-reader/event"
)

type WaitForAuthSound struct{}

func (h *WaitForAuthSound) Handle(e event.Event) error {
	log.Trace("Waiting for the auth sound to stop")
	waitForAsync("authSoundPlaying")
	waitForAsync("showStatus")

	context.LEDController.FadeOff(fadeEffectDuration)
	log.Trace("auth sound has stopped")
	return nil
}

func init() {
	if err := Register(32, &WaitForAuthSound{}); err != nil {
		panic(err)
	}
}
