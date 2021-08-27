package handler

import (
	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/context"
	"github.com/bcurnow/magicband-reader/event"
)

type Authorize struct{}

func (h *Authorize) Handle(e event.Event) error {
	log.Tracef("Authenticating '%v'", e.UID())
	if authorized := context.AuthController.Authorized(e); authorized {
		e.SetType(event.AUTHORIZED)
	} else {
		e.SetType(event.UNAUTHORIZED)
	}
	return nil
}

func init() {
	if err := context.RegisterHandler(12, &Authorize{}); err != nil {
		panic(err)
	}
}
