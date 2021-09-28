package handler

import (
	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/context"
	"github.com/bcurnow/magicband-reader/event"
	"github.com/bcurnow/magicband-reader/led"
)

type Spin struct{}

func (h *Spin) Handle(e event.Event) error {
	log.Trace("Spinning the lights")
	stop := make(chan bool)
	context.State["stopSpinning"] = stop

	runAsync("spinning", func() {
		context.LEDController.Spin(led.WHITE, context.LEDController.Brightness(), reverseSpin, colorChaseWidth, stop)
	})

	return nil
}

func init() {
	if err := context.RegisterHandler(11, &Spin{}); err != nil {
		panic(err)
	}
}
