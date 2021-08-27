package handler

import (
	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/context"
	"github.com/bcurnow/magicband-reader/event"
)

type StopSpin struct{}

func (h *StopSpin) Handle(e event.Event) error {
	log.Trace("Stopping the spin")
	waitForAsync("readSoundPlaying")

	// Make the spinning stop
	close(context.State["stopSpinning"].(chan bool))
	defer context.ClearState("stopSpinning")

	// Wait for the actual spinning to stop
	waitForAsync("spinning")
	log.Trace("Spinning has stopped")
	return nil
}

func init() {
	if err := context.RegisterHandler(13, &StopSpin{}); err != nil {
		panic(err)
	}
}
