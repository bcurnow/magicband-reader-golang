package main

import (
	"github.com/bcurnow/magicband-reader/audio"
	//"github.com/bcurnow/magicband-reader/led"
)

func main() {
	audioController := audio.Controller{}
	if err := audioController.Init(); err != nil {
		panic(err)
	}
	audioStream, err := audioController.Load("/sounds/test.wav")
	if err != nil {
		panic(err)
	}

	audioController.Play(audioStream)
}
