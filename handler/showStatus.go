package handler

import (
	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/context"
	"github.com/bcurnow/magicband-reader/event"
	"github.com/bcurnow/magicband-reader/led"
)

type ShowStatus struct{}

func (h *ShowStatus) Handle(e event.Event) error {
	log.Trace("Showing status")
	switch e.Type() {
	case event.AUTHORIZED:
		runAsync("showStatus", func() {
			context.LEDController.FadeOn(led.GREEN, fadeEffectDelay)
		})
	case event.UNAUTHORIZED:
		runAsync("showStatus", func() {
			context.LEDController.FadeOn(led.BLUE, fadeEffectDelay)
		})
	}
	return nil
}

func init() {
	if err := context.RegisterHandler(20, &ShowStatus{}); err != nil {
		panic(err)
	}
}
