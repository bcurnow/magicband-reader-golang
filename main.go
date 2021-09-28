package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/config"
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
	router, err := NewRouter(config.ListenAddress, config.ListenPort)
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
		close(shutdown)
	}()

	if err := context.AudioCache.Sync(); err != nil {
		panic(err)
	}

	//Blink the LED strip to indicate that the software is started and we're reading
	//the UID
	context.LEDController.Blink(led.WHITE, blinkBrightness, blinkIterations, blinkDelay)
	log.Info("Waiting for MagicBand...")

	// There's an issue with the library we're using that I didn't see in the Python version
	// As long as you hold the band over the reader, it just keeps reporting the UID over and over
	// To fix this, we're going to throw away any uid reads which are the same as the last UID and
	// happened within 10 seconds
	var lastUid string
	stopUidCleanup := make(chan bool)
	uidCleanupInit := 24 * time.Hour
	uidCleanupPeriod := 10 * time.Second
	uidCleanup := time.NewTicker(uidCleanupInit)
	go func() {
		for {
			select {
			case <-stopUidCleanup:
				return
			case <-uidCleanup.C:
				log.Tracef("Clearing '%v' from lastUid", lastUid)
				// Time to clean out the UID
				lastUid = ""
				uidCleanup.Reset(uidCleanupInit)
			}
		}
	}()

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
		// we only want to route the uid once so only route if it's a different uid
		if uid != "" && uid != lastUid {
			router.Route(event.NewEvent(uid, event.UNKNOWN))
			lastUid = uid
			uidCleanup.Reset(uidCleanupPeriod)
		}
	}

	// Stop the UID cleaner
	log.Trace("Stopping the UID cleaner")
	uidCleanup.Stop()
	close(stopUidCleanup)

	log.Debug("Waiting for shutdown")
	<-shutdown
	log.Info("Shutdown complete")
}
