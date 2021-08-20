package led

import (
	"time"

	ws2811 "github.com/rpi-ws281x/rpi-ws281x-go"
)

const (
	defaultBrightness = 64
	defaultOuterRing  = 40
	defaultInnerRing  = 15
)

type Controller struct {
	// The brightness of the pixels between 0 and 255 (0 = off)
	Brightness int
	// The number of LEDs in the outer ring
	OuterRing int
	// The number of LEDs in the inner rignt
	InnerRing int
	// The type of strip, one of the WS2811StripXXX constants from ws2818
	StripType int
	// The ws2811 instance pointer
	strip *ws2811.WS2811
	// The current brightness level
	currentBrightness int
}

func (c *Controller) Init() error {
	c.handleDefaults()
	opt := ws2811.DefaultOptions
	opt.Channels[0].StripeType = c.StripType
	opt.Channels[0].Brightness = c.Brightness
	opt.Channels[0].LedCount = c.OuterRing + c.InnerRing

	dev, err := ws2811.MakeWS2811(&opt)
	if err != nil {
		return err
	}

	dev.Init()
	c.strip = dev
	if err := c.startupSequence(); err != nil {
		return err
	}
	return nil
}

func (c *Controller) handleDefaults() {
	if c.Brightness == 0 {
		c.Brightness = defaultBrightness
	}

	if c.OuterRing == 0 {
		c.OuterRing = defaultOuterRing
	}

	if c.InnerRing == 0 {
		c.InnerRing = defaultInnerRing
	}

	if c.StripType == 0 {
		c.StripType = ws2811.WS2812Strip
	}
}

/** Fini should be called to ensure that the controller properly shuts down the LED strip
 *  For example:
 *		c.Init()
 *		defer c.Fini()
 **/
func (c *Controller) Fini() {
	if c.strip != nil {
		c.strip.Fini()
	}
}

func (c *Controller) startupSequence() error {
	if err := c.Blink(WHITE, defaultBrightness, 2, 500*time.Millisecond); err != nil {
		return err
	}
	return nil
}

/**
 *  Blink implements a blink effect by on and of the lights
 */
func (c *Controller) Blink(color LedColor, brightness int, iterations int, delay time.Duration) error {
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

func (c *Controller) LightsOn(color LedColor, brightness int) error {
	c.setBrightness(brightness)
	c.fill(color)
	if err := c.strip.Render(); err != nil {
		return err
	}
	return nil
}

func (c *Controller) FadeOn(color LedColor, brightness int, delay time.Duration) error {
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

func (c *Controller) LightsOff() error {
	c.fill(0)
	if err := c.strip.Render(); err != nil {
		return err
	}
	return nil
}

func (c *Controller) FadeOff(delay time.Duration) error {
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

func (c *Controller) ColorChase(color LedColor, brightness int, delay time.Duration, reverse bool, effectLength int) error {
	c.setBrightness(brightness)
	var on int = 0
	var off int = 0
	for i := 0; i < c.OuterRing+effectLength+1; i++ {
		if i <= c.OuterRing {
			on = i
			if reverse {
				on = c.OuterRing - on
			}
			c.setLedColor(on, color)
		}

		if i >= effectLength {
			off = i - effectLength
			if reverse {
				off = c.OuterRing - off
			}
			c.setLedColor(off, 0)
		}

		if err := c.strip.Render(); err != nil {
			return err
		}
		if i < c.OuterRing+effectLength {
			time.Sleep(delay)
		}
	}
	return nil
}

func (c *Controller) Spin(color LedColor, brightness int, reverse bool, effectLength int) error {
	c.ColorChase(color, brightness, 10000*time.Microsecond, reverse, effectLength)
	c.ColorChase(color, brightness, 5000*time.Microsecond, reverse, effectLength)
	c.ColorChase(color, brightness, 2500*time.Microsecond, reverse, effectLength)
	c.ColorChase(color, brightness, 1250*time.Microsecond, reverse, effectLength)
	return nil
}

func (c *Controller) setLedColor(led int, color LedColor) {
	c.strip.Leds(0)[led] = uint32(color)
}

func (c *Controller) fill(color LedColor) {
	for i := range c.strip.Leds(0) {
		c.setLedColor(i, color)
	}
}

func (c *Controller) setBrightness(brightness int) {
	c.strip.SetBrightness(0, brightness)
	c.currentBrightness = brightness
}
