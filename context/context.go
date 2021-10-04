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
	AudioCache      audio.Cache
	RFIDSecuritySvc rfidsecuritysvc.Service
	LEDController   led.Controller
	Permission      string
	State           map[string]interface{}
)

func init() {
	log.Debug("Initializing Context")

	service, err := rfidsecuritysvc.New(config.ApiKey, config.ApiSSLVerify, config.ApiUrl)
	if err != nil {
		panic(err)
	}
	RFIDSecuritySvc = service

	audioCache, err := audio.NewCache(RFIDSecuritySvc, config.SoundDir)
	if err != nil {
		panic(err)
	}
	AudioCache = audioCache

	audioController, err := audio.NewController(config.VolumeLevel, 0, audioCache, config.AuthorizedSound, config.ReadSound, config.UnauthorizedSound)
	if err != nil {
		panic(err)
	}
	AudioController = audioController

	ledController, err := led.NewController(config.Brightness, config.OuterRingSize, config.InnerRingSize, 0)
	if err != nil {
		panic(err)
	}
	LEDController = ledController

	State = make(map[string]interface{})
	// The permission is really part of the context for the application, this also
	// reduces the direct dependencies on config
	Permission = config.Permission
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
