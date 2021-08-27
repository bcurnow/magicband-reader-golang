package handler

import (
	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/context"
	"github.com/bcurnow/magicband-reader/event"
)

type WaitForSpin struct{}

func (h *WaitForSpin) Handle(e event.Event) error {
	log.Trace("Waiting for spin to stop")
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
	if err := Register(20, &WaitForSpin{}); err != nil {
		panic(err)
	}
}
