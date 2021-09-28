package handler

import (
	"github.com/faiface/beep"
	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/context"
	"github.com/bcurnow/magicband-reader/event"
)

type ReadSound struct {
	sound *beep.Buffer
}

func (h *ReadSound) Handle(e event.Event) error {
	log.Trace("Playing the read sound")
	return runAsync("readSoundPlaying", func() {
		context.AudioController.Play(h.sound)
	})
}

func init() {
	sound, err := context.AudioController.Load(context.AudioController.ReadSound())
	if err != nil {
		panic(err)
	}

	handler := &ReadSound{
		sound: sound,
	}

	if err := context.RegisterHandler(10, handler); err != nil {
		panic(err)
	}
}
