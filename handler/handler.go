/*
 * The handler package contains all the various steps in the MagicBand reader flow, the general approach is to break the flow into sections
 * using a two digit number (e.g. priority 10 = section one, step zero, priority 11 = section one, step 1, priority 20 = section 2, step 0).
 * The current flow is broken down as follows:
 * Section One
 *   10 - readSound
 *   11 - spin
 *   12 - authorize
 *   13 - stopSpin
 * Section Two
 *   20 - showStatus
 *   21 - authSound
 *   22 - stopStatus
 *
 * There is also a handler at priority math.MaxInt which is designed to be the last step and will log the results.
 */
package handler

import (
	"errors"
	"fmt"
	"time"

	"github.com/bcurnow/magicband-reader/context"
)

const (
	colorChaseWidth = 8
	reverseSpin     = true
	ledBrightness   = 150
	fadeEffectDelay = 5 * time.Millisecond
)

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
