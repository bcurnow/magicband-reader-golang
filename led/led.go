package led

import (
	"fmt"
	"time"

	ws2811 "github.com/rpi-ws281x/rpi-ws281x-go"
	log "github.com/sirupsen/logrus"
)

const (
	defaultBrightness    = 64
	defaultOuterRingSize = 40
	defaultInnerRingSize = 15
)

type Controller interface {
	Blink(color Color, iterations int, delay time.Duration) error
	LightsOn(color Color) error
	FadeOn(color Color, delay time.Duration) error
	LightsOff() error
	FadeOff(delay time.Duration) error
	ColorChase(color Color, delay time.Duration, reverse bool, effectLength int) error
	Spin(color Color, reverse bool, effectLength int, stop <-chan bool) error
	Close()
}

type controller struct {
	// The brightness of the pixels between 0 and 255 (0 = off)
	brightness int
	// The number of LEDs in the outer ring
	outerRingSize int
	// The number of LEDs in the inner rignt
	innerRingSize int
	// The type of strip, one of the WS2811StripXXX constants from ws2818
	stripType int
	// The ws2811 instance pointer
	strip *ws2811.WS2811
	// The current brightness level
	currentBrightness int
}

func NewController(brightness int, outerRingSize int, innerRingSize int, stripType int) (Controller, error) {
	log.Trace("Creating new led.Controller")

	if brightness < 0 || brightness > 255 {
		return nil, fmt.Errorf("Invalid value for brightness: '%v'. Must be between 0 and 255 inclusive.", brightness)
	}

	c := controller{
		brightness:    brightness,
		outerRingSize: outerRingSize,
		innerRingSize: innerRingSize,
		stripType:     stripType,
	}
	c.handleDefaults()

	opt := ws2811.DefaultOptions
	opt.Channels[0].StripeType = c.stripType
	opt.Channels[0].Brightness = c.brightness
	opt.Channels[0].LedCount = c.outerRingSize + c.innerRingSize

	dev, err := ws2811.MakeWS2811(&opt)
	if err != nil {
		return nil, err
	}

	c.strip = dev
	c.strip.Init()
	return &c, nil
}

/** Close should be called to ensure that the controller properly shuts down the LED strip
 *  For example:
 *		c.Init()
 *		defer c.Close()
 **/
func (c *controller) Close() {
	log.Trace("Closing led.Controller")
	if c.strip != nil {
		c.strip.Fini()
	}
	log.Trace("led.Controller closed")
}

/**
 *  Blink implements a blink effect by on and of the lights
 */
func (c *controller) Blink(color Color, iterations int, delay time.Duration) error {
	for i := 0; i < iterations; i++ {
		if err := c.LightsOn(color); err != nil {
			return err
		}
		time.Sleep(delay)
		if err := c.LightsOff(); err != nil {
			return err
		}
		if i < iterations-1 {
			time.Sleep(delay)
		}
	}
	return nil
}

func (c *controller) LightsOn(color Color) error {
	c.setBrightness(c.brightness)
	c.fill(color)
	if err := c.strip.Render(); err != nil {
		return err
	}
	return nil
}

func (c *controller) FadeOn(color Color, delay time.Duration) error {
	c.fill(color)

	for currentBrightness := 1; currentBrightness <= c.brightness; currentBrightness++ {
		log.Debug(currentBrightness)
		c.setBrightness(currentBrightness)
		if err := c.strip.Render(); err != nil {
			return err
		}
		time.Sleep(delay)
	}
	return nil
}

func (c *controller) LightsOff() error {
	c.fill(0)
	if err := c.strip.Render(); err != nil {
		return err
	}
	return nil
}

func (c *controller) FadeOff(delay time.Duration) error {
	for currentBrightness := c.currentBrightness; currentBrightness >= 0; currentBrightness-- {
		c.setBrightness(currentBrightness)
		if err := c.strip.Render(); err != nil {
			return err
		}
		time.Sleep(delay)
	}
	// Make sure to turn all the LEDs back to 0 otherwise they'll remember the color
	// they were and come back on with the next call to Render()
	c.fill(0)
	return nil
}

func (c *controller) ColorChase(color Color, delay time.Duration, reverse bool, effectLength int) error {
	c.setBrightness(c.brightness)
	var on int = 0
	var off int = 0
	for i := 0; i < c.outerRingSize+effectLength+1; i++ {
		if i <= c.outerRingSize {
			on = i
			if reverse {
				on = c.outerRingSize - on
			}
			c.setLedColor(on, color)
		}

		if i >= effectLength {
			off = i - effectLength
			if reverse {
				off = c.outerRingSize - off
			}
			c.setLedColor(off, 0)
		}

		if err := c.strip.Render(); err != nil {
			return err
		}
		if i < c.outerRingSize+effectLength {
			time.Sleep(delay)
		}
	}
	return nil
}

/*
 * Spin will spin (ColorChase) 3 times at increasingly faster intervals and then continue to spin at the fastest
 * interval until it receives on the stop channel.
 */
func (c *controller) Spin(color Color, reverse bool, effectLength int, stop <-chan bool) error {
	c.ColorChase(color, 10*time.Millisecond, reverse, effectLength)
	c.ColorChase(color, 5*time.Millisecond, reverse, effectLength)
	c.ColorChase(color, 2500*time.Microsecond, reverse, effectLength)
	for {
		select {
		case <-stop:
			return nil
		default:
			c.ColorChase(color, 1250*time.Microsecond, reverse, effectLength)
		}
	}
	return nil
}

func (c *controller) handleDefaults() {
	if c.brightness == 0 {
		c.brightness = defaultBrightness
	}

	if c.outerRingSize == 0 {
		c.outerRingSize = defaultOuterRingSize
	}

	if c.innerRingSize == 0 {
		c.innerRingSize = defaultInnerRingSize
	}

	if c.stripType == 0 {
		c.stripType = ws2811.WS2812Strip
	}
}

func (c *controller) setLedColor(led int, color Color) {
	c.strip.Leds(0)[led] = uint32(color)
}

func (c *controller) fill(color Color) {
	for i := range c.strip.Leds(0) {
		c.setLedColor(i, color)
	}
}

func (c *controller) setBrightness(brightness int) {
	c.strip.SetBrightness(0, brightness)
	c.currentBrightness = brightness
}
