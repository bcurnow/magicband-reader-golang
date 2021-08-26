package context

import (
	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/audio"
	"github.com/bcurnow/magicband-reader/config"
	"github.com/bcurnow/magicband-reader/led"
)

type ctx struct {
	AudioController audio.Controller
	LEDController   led.Controller
	State           map[string]interface{}
}

var Current ctx

func init() {
	log.Debug("Initializing Context")

	ledController, err := led.NewController(config.Values.Brightness, config.Values.OuterRingSize, config.Values.InnerRingSize, 0)
	if err != nil {
		panic(err)
	}

	audioController, err := audio.NewController(config.Values.VolumeLevel, 0)
	if err != nil {
		panic(err)
	}

	Current = ctx{
		LEDController:   ledController,
		AudioController: audioController,
		State:           make(map[string]interface{}),
	}
}

func Close() error {
	log.Debug("Closing context")
	Current.LEDController.Close()
	return nil
}
