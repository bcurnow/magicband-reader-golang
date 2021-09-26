package rfidsecuritysvc

import (
	"fmt"
	"net/url"

	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/event"
)

const (
	permissionUrlFormat = "authorized/%v/%v"
)

func (s *service) Authorized(event event.Event, permission string) bool {
	url := fmt.Sprintf(permissionUrlFormat, url.PathEscape(event.UID()), url.PathEscape(permission))
	if err := s.Get(url, 200, nil); err != nil {
		log.Debugf("Error calling '%v': %v", url, err)
		return false
	}
	return true
}
