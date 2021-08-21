package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/audio"
	"github.com/bcurnow/magicband-reader/led"
)

func main() {
	config := Config{}
	if err := config.Init(); err != nil {
		panic(err)
	}

	log.Debug("Initializing LED")
	ledController := led.Controller{OuterRing: 144, InnerRing: 0}
	if err := ledController.Init(); err != nil {
		panic(err)
	}

	log.Debug("Initializing Audio")
	audioController := audio.Controller{}
	if err := audioController.Init(); err != nil {
		panic(err)
	}
	audioStream, err := audioController.Load("/sounds/test.wav")
	if err != nil {
		panic(err)
	}

	log.Debug("Playing audio")
	audioController.Play(audioStream)
	log.Debug("Exiting")
}
