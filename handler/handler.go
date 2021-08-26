package handler

import (
	"errors"
	"fmt"
	"sort"

	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/event"
)

type Handler interface {
	Handle(event event.Event) error
}

var Handlers map[int]Handler

func init() {
	Handlers = make(map[int]Handler)
}

func Register(priority int, handler Handler) error {
	existingHandler, exists := Handlers[priority]

	if exists {
		return errors.New(fmt.Sprintf("Handler '%v' already registered with priority %v", existingHandler, priority))
	}
	Handlers[priority] = handler
	log.Debug(fmt.Sprintf("Handler '%v' registered with priority %v", handler, priority))
	return nil
}

func Sorted() []Handler {
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
