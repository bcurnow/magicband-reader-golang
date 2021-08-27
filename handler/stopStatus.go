package handler

import (
	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/context"
	"github.com/bcurnow/magicband-reader/event"
)

type StopStatus struct{}

func (h *StopStatus) Handle(e event.Event) error {
	log.Trace("Waiting for the auth sound to stop")
	waitForAsync("authSoundPlaying")
	waitForAsync("showStatus")

	context.LEDController.FadeOff(fadeEffectDelay)
	log.Trace("auth sound has stopped")
	return nil
}

func init() {
	if err := context.RegisterHandler(22, &StopStatus{}); err != nil {
		panic(err)
	}
}
