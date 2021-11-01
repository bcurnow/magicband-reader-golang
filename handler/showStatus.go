package handler

import (
	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/context"
	"github.com/bcurnow/magicband-reader/event"
	"github.com/bcurnow/magicband-reader/led"
	"github.com/bcurnow/magicband-reader/rfidsecuritysvc"
)

const (
	defaultAuthColor = led.GREEN
)

type ShowStatus struct{}

func (h *ShowStatus) Handle(e event.Event) error {
	log.Trace("Showing status")
	switch e.Type() {
	case event.AUTHORIZED:
		runAsync("showStatus", func() {
			context.LEDController.FadeOn(resolveColor(), fadeEffectDelay)
		})
	case event.UNAUTHORIZED:
		runAsync("showStatus", func() {
			context.LEDController.FadeOn(led.BLUE, fadeEffectDelay)
		})
	}
	return nil
}

func resolveColor() led.Color {
	// Not sure how this would happen but we don't have a MediaConfig object in state
	if context.State["mediaConfig"] == nil {
		log.Errorf("Unable to find mediaConfig in State, using default authorized color")
		return defaultAuthColor
	}

	mediaConfig := context.State["mediaConfig"].(*rfidsecuritysvc.MediaConfig)
	if mediaConfig.Color == nil {
		log.Infof("No color configured in mediaConfig, using default authorized color")
		return defaultAuthColor
	}
	return led.Color(uint32(mediaConfig.Color.Int))
}

func init() {
	if err := context.RegisterHandler(20, &ShowStatus{}); err != nil {
		panic(err)
	}
}
