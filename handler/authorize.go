package handler

import (
	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/context"
	"github.com/bcurnow/magicband-reader/event"
)

type Authorize struct {
	permission string
}

func (h *Authorize) Handle(e event.Event) error {
	log.Tracef("Authenticating '%v'", e.UID())
	if authorized := context.RFIDSecuritySvc.Authorized(e, h.permission); authorized {
		e.SetType(event.AUTHORIZED)
	} else {
		e.SetType(event.UNAUTHORIZED)
	}
	return nil
}

func init() {
	if err := context.RegisterHandler(12, &Authorize{permission: context.Permission}); err != nil {
		panic(err)
	}
}
