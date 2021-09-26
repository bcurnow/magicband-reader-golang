package context

import (
	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/audio"
	"github.com/bcurnow/magicband-reader/config"
	"github.com/bcurnow/magicband-reader/led"
	"github.com/bcurnow/magicband-reader/rfidsecuritysvc"
)

var (
	AudioController audio.Controller
	RFIDSecuritySvc rfidsecuritysvc.Service
	LEDController   led.Controller
	State           map[string]interface{}
)

func init() {
	log.Debug("Initializing Context")

	audioController, err := audio.NewController(config.VolumeLevel, 0)
	if err != nil {
		panic(err)
	}
	AudioController = audioController

	service, err := rfidsecuritysvc.New(config.ApiKey, config.CaCertFile, config.ApiUrl, config.ValidateCertificates)
	if err != nil {
		panic(err)
	}
	RFIDSecuritySvc = service

	ledController, err := led.NewController(config.Brightness, config.OuterRingSize, config.InnerRingSize, 0)
	if err != nil {
		panic(err)
	}
	LEDController = ledController

	State = make(map[string]interface{})
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
