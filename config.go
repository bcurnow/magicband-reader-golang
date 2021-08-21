package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path"

	log "github.com/sirupsen/logrus"

	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/ffyaml"
)

type Config struct {
	ApiKey            string
	ApiSSLVerify      string
	ApiURL            string
	AuthorizedSound   string
	Brightness        int
	ConfigFile        string
	InnerRingSize     int
	LogLevel          log.Level
	OuterRingSize     int
	Permission        string
	PortNumber        int
	ReadSound         string
	SoundDir          string
	UnauthorizedSound string
	VolumeLevel       float64
}

func (c *Config) Init() error {
	fs := flag.NewFlagSet("magicband-reader", flag.ExitOnError)
	var (
		apiKey            = fs.String("api-key", "", "The API key to authenticate to rfid-security-svc")
		apiSSLVerify      = fs.String("api-ssl-verify", "ca.pem", "If 'True' or a valid file reference, performs SSL validation, if false, skips validation (this is insecure!).")
		apiURL            = fs.String("api-url", "https://localhost:5000/api/v1.0", "The rfid-security-svc base URL.")
		authorizedSound   = fs.String("authorized-sound", "authorized.wav", "The name of the sound file played when a band is authorized (relative to sound-dir).")
		brightness        = fs.Int("brightness", 100, "The brightness level of the LEDs. Range of 0 to 255 inclusive")
		configFile        = fs.String("config-file", "/etc/magicband-reader/magicband-reader.yaml", "The YAML configuration file to load.")
		innerRingSize     = fs.Int("inner-ring-size", 10, "The number of LEDs that make up the inner ring.")
		logLevel          = fs.String("log-level", "warning", "One of: debug, info, warning, error fatal.")
		outerRingSize     = fs.Int("outer-ring-size", 10, "The number of LEDs that make up the outer ring.")
		permission        = fs.String("permission", "Open Door", "The name of the permission to validate before authorizing.")
		portNumber        = fs.Int("port-number", 8080, "The port number to listen for requests for UID (e.g. from rfid-security-svc)")
		readSound         = fs.String("read-sound", "read.wav", "The name of the sound file played when a band is read (relative to sound-dir).")
		soundDir          = fs.String("sound-dir", "/sounds", "The directory containing the sound files.")
		unauthorizedSound = fs.String("unauthorized-sound", "unauthorized.wav", "The name of the sound file played when a band is unauthorized (relative to sound-dir).")
		volumeLevel       = fs.Float64("volume-level", 0, "Positive or negative value which is applied to the volume base to adjust the sound.")
	)

	if err := ff.Parse(fs, os.Args[1:],
		ff.WithEnvVarPrefix("MR"),
		ff.WithConfigFileFlag("config-file"),
		ff.WithConfigFileParser(ffyaml.Parser),
		ff.WithAllowMissingConfigFile(true),
	); err != nil {
		return err
	}

	c.ApiKey = *apiKey
	c.ApiSSLVerify = *apiSSLVerify
	c.ApiURL = *apiURL
	c.AuthorizedSound = *authorizedSound
	c.Brightness = *brightness
	c.ConfigFile = *configFile
	c.InnerRingSize = *innerRingSize
	c.OuterRingSize = *outerRingSize
	c.Permission = *permission
	c.PortNumber = *portNumber
	c.ReadSound = *readSound
	c.SoundDir = *soundDir
	c.UnauthorizedSound = *unauthorizedSound
	c.VolumeLevel = *volumeLevel

	if err := c.validate(*logLevel); err != nil {
		return err
	}

	c.configureLog()

	log.Debug(fmt.Sprintf("api-key: %v", c.ApiKey))
	log.Debug(fmt.Sprintf("api-ssl-verify: %v", c.ApiSSLVerify))
	log.Debug(fmt.Sprintf("api-url: %v", c.ApiURL))
	log.Debug(fmt.Sprintf("authorized-sound: %v", c.AuthorizedSound))
	log.Debug(fmt.Sprintf("brightness: %v", c.Brightness))
	log.Debug(fmt.Sprintf("config-file: %v", c.ConfigFile))
	log.Debug(fmt.Sprintf("inner-ring-size: %v", c.InnerRingSize))
	log.Debug(fmt.Sprintf("log-level: %v", c.LogLevel))
	log.Debug(fmt.Sprintf("outer-ring-size: %v", c.OuterRingSize))
	log.Debug(fmt.Sprintf("permission: %v", c.Permission))
	log.Debug(fmt.Sprintf("port-number: %v", c.PortNumber))
	log.Debug(fmt.Sprintf("read-sound: %v", c.ReadSound))
	log.Debug(fmt.Sprintf("sound-dir: %v", c.SoundDir))
	log.Debug(fmt.Sprintf("unauthorized-sound: %v", c.UnauthorizedSound))
	log.Debug(fmt.Sprintf("volume-level: %v", c.VolumeLevel))

	return nil
}

func (c *Config) validate(logLevel string) error {
	if err := c.validateFileExists(path.Join(c.SoundDir, c.AuthorizedSound), "authorized-sound"); err != nil {
		return err
	}

	if err := c.validateIntRange(c.Brightness, 0, 255, "brightness"); err != nil {
		return err
	}

	level, err := c.validateLogLevel(logLevel, "log-level")
	if err != nil {
		return err
	}
	c.LogLevel = level

	if err := c.validateFileExists(path.Join(c.SoundDir, c.ReadSound), "read-sound"); err != nil {
		return err
	}

	if err := c.validateFileExists(c.SoundDir, "sound-dir"); err != nil {
		return err
	}

	if err := c.validateFileExists(path.Join(c.SoundDir, c.UnauthorizedSound), "unauthorized-sound"); err != nil {
		return err
	}
	return nil
}

func (c *Config) validateIntRange(value int, low int, high int, name string) error {
	if value < low || value > high {
		return errors.New(fmt.Sprintf("Invalid value for %v: '%v'. Must be between %v and %v inclusive.", name, value, low, high))
	}
	return nil
}

func (c *Config) validateFileExists(file string, name string) error {
	if _, err := os.Stat(file); err != nil {
		return errors.New(fmt.Sprintf("Invalid value for %v: '%v'. %v", name, file, err))
	}
	return nil
}

func (c *Config) validateLogLevel(level string, name string) (log.Level, error) {
	logLevel, err := log.ParseLevel(level)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("Invalid value for %v: '%v'. %v", name, level, err))
	}
	return logLevel, err
}

func (c *Config) configureLog() {
	log.SetLevel(c.LogLevel)
	log.SetFormatter(&log.TextFormatter{
		DisableLevelTruncation: true,
		PadLevelText:           true,
		FullTimestamp:          true,
	})
}
