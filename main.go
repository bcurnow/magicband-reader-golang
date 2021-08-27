package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/context"
	"github.com/bcurnow/magicband-reader/event"
	"github.com/bcurnow/magicband-reader/led"
)

const (
	blinkBrightness = 64
	blinkIterations = 2
	blinkDelay      = 500 * time.Millisecond
)

func main() {
	router, err := NewRouter()
	if err != nil {
		panic(err)
	}

	shutdown := make(chan bool)

	reader, err := NewDefaultReader()
	if err != nil {
		panic(err)
	}

	// Make sure to clean up everything when we exit
	go func() {
		osSignals := make(chan os.Signal, 1)
		signal.Notify(osSignals, os.Interrupt, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT)
		sig := <-osSignals
		log.Tracef("Received %v", sig)
		close(osSignals)
		router.Close()
		reader.Close()
		context.Close()
		shutdown <- true
	}()

	//Blink the LED strip to indicate that the software is started and we're reading
	//the UID
	context.LEDController.Blink(led.WHITE, blinkBrightness, blinkIterations, blinkDelay)
	log.Info("Waiting for MagicBand...")
	for !router.Closed() {
		// Set a really long wait time
		uid, err := reader.UID(24 * time.Hour)
		if err != nil {
			if _, ok := err.(TimeoutError); ok {
				continue
			}
			// Don't know what happened
			panic(err)
		}

		// uid will be "" if the reader was shutdown due to an OS signal
		if uid != "" {
			router.Route(event.NewEvent(uid, event.UNKNOWN))
		}
	}

	log.Debug("Waiting for shutdown")
	<-shutdown
	log.Info("Shutdown complete")
}
