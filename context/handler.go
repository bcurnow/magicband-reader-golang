package context

import (
	"fmt"
	"sort"

	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/event"
)

type Handler interface {
	Handle(event event.Event) error
}

var (
	Handlers map[int]Handler
)

func init() {
	log.Debug("Initializing Handlers")
	Handlers = make(map[int]Handler)
}

func RegisterHandler(priority int, handler Handler) error {
	existingHandler, exists := Handlers[priority]

	if exists {
		return fmt.Errorf("Handler '%T' already registered with priority %v", existingHandler, priority)
	}
	Handlers[priority] = handler
	log.Debugf("Handler '%T' registered with priority %v", handler, priority)
	return nil
}

func SortedHandlers() []Handler {
	keys := make([]int, 0, len(Handlers))

	for key := range Handlers {
		keys = append(keys, key)
	}

	sort.Ints(keys)

	sorted := make([]Handler, 0, len(Handlers))
	for _, key := range keys {
		sorted = append(sorted, Handlers[key])
	}
	return sorted
}
