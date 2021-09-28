package handler

import (
	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/context"
	"github.com/bcurnow/magicband-reader/event"
)

const (
	// TODO this should really be math.MaxInt but that constant wasn't introduced until golang 1.17
	maxInt = int(^uint(0) >> 1)
)

type Logging struct {
	permission string
}

func (h *Logging) Handle(e event.Event) error {
	log.Debug(e.String())

	switch e.Type() {
	case event.UNKNOWN:
		log.Errorf("Received an unknown event: %v", e.String)
	case event.AUTHORIZED:
		log.Infof("%v was authorized for '%v'", e.UID(), h.permission)
	case event.UNAUTHORIZED:
		log.Warnf("%v was NOT authorized for '%v'", e.UID(), h.permission)
	}
	return nil
}

func init() {
	if err := context.RegisterHandler(maxInt, &Logging{permission: context.Permission}); err != nil {
		panic(err)
	}
}
