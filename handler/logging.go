package handler

import (
	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/event"
)

type Logging struct{}

func (h *Logging) Handle(event event.Event) error {
	log.Debug(event.String())
	return nil
}

func init() {
	Register(999, &Logging{})
}
