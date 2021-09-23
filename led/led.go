package led

import (
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
	Blink(color LedColor, brightness int, iterations int, delay time.Duration) error
	LightsOn(color LedColor, brightness int) error
	FadeOn(color LedColor, brightness int, delay time.Duration) error
	LightsOff() error
	FadeOff(delay time.Duration) error
	ColorChase(color LedColor, brightness int, delay time.Duration, reverse bool, effectLength int) error
	Spin(color LedColor, brightness int, reverse bool, effectLength int, stop <-chan bool) error
	Close()
}

type controller struct {
	// The brightness of the pixels between 0 and 255 (0 = off)
	Brightness int
	// The number of LEDs in the outer ring
	OuterRingSize int
	// The number of LEDs in the inner rignt
	InnerRingSize int
	// The type of strip, one of the WS2811StripXXX constants from ws2818
	StripType int
	// The ws2811 instance pointer
	strip *ws2811.WS2811
	// The current brightness level
	currentBrightness int
}

func NewController(brightness int, outerRingSize int, innerRingSize int, stripType int) (Controller, error) {
	log.Trace("Creating new led.Controller")
	c := controller{
		Brightness:    brightness,
		OuterRingSize: outerRingSize,
		InnerRingSize: innerRingSize,
		StripType:     stripType,
	}
	c.handleDefaults()

	opt := ws2811.DefaultOptions
	opt.Channels[0].StripeType = c.StripType
	opt.Channels[0].Brightness = c.Brightness
	opt.Channels[0].LedCount = c.OuterRingSize + c.InnerRingSize

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
func (c *controller) Blink(color LedColor, brightness int, iterations int, delay time.Duration) error {
	for i := 0; i < iterations; i++ {
		if err := c.LightsOn(color, brightness); err != nil {
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

func (c *controller) LightsOn(color LedColor, brightness int) error {
	c.setBrightness(brightness)
	c.fill(color)
	if err := c.strip.Render(); err != nil {
		return err
	}
	return nil
}

func (c *controller) FadeOn(color LedColor, brightness int, delay time.Duration) error {
	c.fill(color)

	for currentBrightness := 1; currentBrightness <= brightness; currentBrightness++ {
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

func (c *controller) ColorChase(color LedColor, brightness int, delay time.Duration, reverse bool, effectLength int) error {
	c.setBrightness(brightness)
	var on int = 0
	var off int = 0
	for i := 0; i < c.OuterRingSize+effectLength+1; i++ {
		if i <= c.OuterRingSize {
			on = i
			if reverse {
				on = c.OuterRingSize - on
			}
			c.setLedColor(on, color)
		}

		if i >= effectLength {
			off = i - effectLength
			if reverse {
				off = c.OuterRingSize - off
			}
			c.setLedColor(off, 0)
		}

		if err := c.strip.Render(); err != nil {
			return err
		}
		if i < c.OuterRingSize+effectLength {
			time.Sleep(delay)
		}
	}
	return nil
}

/*
 * Spin will spin (ColorChase) 3 times at increasingly faster intervals and then continue to spin at the fastest
 * interval until it receives on the stop channel.
 */
func (c *controller) Spin(color LedColor, brightness int, reverse bool, effectLength int, stop <-chan bool) error {
	c.ColorChase(color, brightness, 10*time.Millisecond, reverse, effectLength)
	c.ColorChase(color, brightness, 5*time.Millisecond, reverse, effectLength)
	c.ColorChase(color, brightness, 2500*time.Microsecond, reverse, effectLength)
	for {
		select {
		case <-stop:
			return nil
		default:
			c.ColorChase(color, brightness, 1250*time.Microsecond, reverse, effectLength)
		}
	}
	return nil
}

func (c *controller) handleDefaults() {
	if c.Brightness == 0 {
		c.Brightness = defaultBrightness
	}

	if c.OuterRingSize == 0 {
		c.OuterRingSize = defaultOuterRingSize
	}

	if c.InnerRingSize == 0 {
		c.InnerRingSize = defaultInnerRingSize
	}

	if c.StripType == 0 {
		c.StripType = ws2811.WS2812Strip
	}
}

func (c *controller) setLedColor(led int, color LedColor) {
	c.strip.Leds(0)[led] = uint32(color)
}

func (c *controller) fill(color LedColor) {
	for i := range c.strip.Leds(0) {
		c.setLedColor(i, color)
	}
}

func (c *controller) setBrightness(brightness int) {
	c.strip.SetBrightness(0, brightness)
	c.currentBrightness = brightness
}
