package main

import (
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/devices/v3/mfrc522"
	"periph.io/x/host/v3"
	"periph.io/x/host/v3/rpi"
)

type TimeoutError struct{}

func (e TimeoutError) Error() string {
	return "Timeout waiting for device data"
}

type Reader interface {
	UID(timeout time.Duration) (string, error)
	Close()
}

type reader struct {
	portCloser spi.PortCloser
	mfrc522    *mfrc522.Dev
	closed     bool
	sync.Mutex
}

func NewDefaultReader() (*reader, error) {
	log.Debug("Creating new default reader")
	reader, err := NewReader("", rpi.P1_22, rpi.P1_18)
	if err != nil {
		return nil, err
	}
	return reader, err
}

func NewReader(spiPort string, resetPin gpio.PinOut, irqPin gpio.PinIn) (*reader, error) {
	log.Debug(fmt.Sprintf("Creating new reader with spiPort='%v', resetPin='%v', irqPin='%v'", spiPort, resetPin, irqPin))
	var reader = reader{}

	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		return nil, err
	}

	pc, err := spireg.Open(spiPort)
	if err != nil {
		return nil, err
	}
	reader.portCloser = pc

	mfrc522, err := mfrc522.NewSPI(pc, resetPin, irqPin)
	if err != nil {
		return nil, err
	}
	// Setting the antenna signal strength.
	mfrc522.SetAntennaGain(5)

	reader.mfrc522 = mfrc522

	return &reader, nil
}

func (r *reader) Close() {
	log.Debug("Closing Reader")
	r.closed = true
	// Halt the device before we lock, this will stop any in-progress reads
	r.mfrc522.Halt()

	//Make sure to lock before we close the underlying port or we'll cause a SIGSEGV
	r.Lock()
	defer r.Unlock()
	//Close the underlying SPI port
	r.portCloser.Close()
}

func (r *reader) UID(timeout time.Duration) (string, error) {
	timedOut := false
	uidChannel := make(chan []byte)
	haltChannel := make(chan error)
	timer := time.NewTimer(timeout)

	// Stopping timer, flagging reader as timed out
	defer func() {
		timer.Stop()
		timedOut = true
		close(uidChannel)
	}()

	go func() {
		for !r.closed {
			// Trying to read card UID.
			r.Lock()
			if r.closed {
				return
			}
			uid, err := r.mfrc522.ReadUID(timeout)
			r.Unlock()

			// If main thread timed out just exit.
			if timedOut {
				return
			}

			if err != nil {
				// The underlying card was halted, probably by the os signal handler
				// We're done
				if err.Error() == "mfrc522 lowlevel: halt" {
					haltChannel <- err
					return
				}
				// Not sure what the error is, we should just try to read again
				log.Trace(fmt.Sprintf("reader.UID: Unexpected error: %v", err))
				continue
			}

			// We got a UID, we're done
			uidChannel <- uid
			return
		}
	}()

	for {
		select {
		case <-timer.C:
			return "", TimeoutError{}
		case uid := <-uidChannel:
			// TODO Remove this, this is a crazy work around
			time.Sleep(1 * time.Second)
			return strings.ToUpper(hex.EncodeToString(uid)), nil
		case <-haltChannel:
			return "", nil
		}
	}
}
