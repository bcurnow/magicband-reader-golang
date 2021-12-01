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
	if mediaConfig, err := context.RFIDSecuritySvc.Authorized(e, h.permission); err != nil {
		e.SetType(event.UNAUTHORIZED)
	} else {
		e.SetType(event.AUTHORIZED)
		context.State["mediaConfig"] = mediaConfig
		log.Tracef("%+v", mediaConfig)
	}
	return nil
}

func init() {
	if err := context.RegisterHandler(12, &Authorize{permission: context.Permission}); err != nil {
		panic(err)
	}
}
