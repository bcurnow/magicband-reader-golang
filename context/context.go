package context

import (
	"errors"
	"fmt"
	"sort"

	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/audio"
	"github.com/bcurnow/magicband-reader/auth"
	"github.com/bcurnow/magicband-reader/config"
	"github.com/bcurnow/magicband-reader/event"
	"github.com/bcurnow/magicband-reader/led"
)

type Handler interface {
	Handle(event event.Event) error
}

var (
	AudioController audio.Controller
	AuthController  auth.Controller
	Handlers        map[int]Handler
	LEDController   led.Controller
	State           map[string]interface{}
)

func init() {
	log.Debug("Initializing Context")

	audioController, err := audio.NewController(config.VolumeLevel, 0)
	if err != nil {
		panic(err)
	}

	authController, err := auth.NewController(config.ApiKey, config.CaCertFile, config.ApiUrl, config.Permission, config.ValidateCertificates)
	if err != nil {
		panic(err)
	}

	ledController, err := led.NewController(config.Brightness, config.OuterRingSize, config.InnerRingSize, 0)
	if err != nil {
		panic(err)
	}

	AudioController = audioController
	AuthController = authController
	Handlers = make(map[int]Handler)
	LEDController = ledController
	State = make(map[string]interface{})
}

func RegisterHandler(priority int, handler Handler) error {
	existingHandler, exists := Handlers[priority]

	if exists {
		return errors.New(fmt.Sprintf("Handler '%T' already registered with priority %v", existingHandler, priority))
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

func Close() error {
	log.Debug("Closing context")
	LEDController.Close()
	log.Trace("context closed")
	return nil
}

func ClearState(key string) {
	delete(State, key)
}
