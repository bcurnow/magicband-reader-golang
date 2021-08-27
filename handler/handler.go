package handler

import (
	"errors"
	"fmt"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/context"
	"github.com/bcurnow/magicband-reader/event"
)

const (
	colorChaseWidth    = 8
	reverseSpin        = true
	ledBrightness      = 150
	fadeEffectDuration = 10 * time.Millisecond
)

type Handler interface {
	Handle(event event.Event) error
}

var Handlers map[int]Handler

func init() {
	initMap()
}

func initMap() {
	//TODO, this is dumb and only started happening when I added a handler that starts with a
	//Golang runs them in file name order within a package <grrrr>
	if Handlers == nil {
		Handlers = make(map[int]Handler)
	}
}

func Register(priority int, handler Handler) error {
	initMap()
	existingHandler, exists := Handlers[priority]

	if exists {
		return errors.New(fmt.Sprintf("Handler '%T' already registered with priority %v", existingHandler, priority))
	}
	Handlers[priority] = handler
	log.Debugf("Handler '%T' registered with priority %v", handler, priority)
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

func runAsync(channelName string, f func()) error {
	if _, exists := context.State[channelName]; exists {
		return errors.New(fmt.Sprintf("Unable to runAsync, channel with name '%v' already exists in State: %v", channelName, context.State))
	}

	ch := make(chan bool)
	go func() {
		f()
		close(ch)
	}()

	context.State[channelName] = ch
	return nil
}

func waitForAsync(channelName string) {
	<-context.State[channelName].(chan bool)
	context.ClearState(channelName)
}
