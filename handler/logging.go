package handler

import (
	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/config"
	"github.com/bcurnow/magicband-reader/event"
)

type Logging struct{}

func (h *Logging) Handle(e event.Event) error {
	log.Debug(e.String())

	switch e.Type() {
	case event.UNKNOWN:
		log.Errorf("Received an unknown event: %v", e.String)
	case event.AUTHORIZED:
		log.Infof("%v was authorized for '%v'", e.UID(), config.Permission)
	case event.UNAUTHORIZED:
		log.Warnf("%v was NOT authorized for '%v'", e.UID(), config.Permission)
	}
	return nil
}

func init() {
	if err := Register(999, &Logging{}); err != nil {
		panic(err)
	}
}
