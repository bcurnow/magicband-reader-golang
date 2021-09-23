package config

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/ffyaml"
)

var (
	ApiKey               string
	ApiUrl               *url.URL
	AuthorizedSound      string
	Brightness           int
	CaCertFile           string
	ConfigFile           string
	InnerRingSize        int
	ListenAddress        string
	ListenPort           int
	OuterRingSize        int
	Permission           string
	ReadSound            string
	SoundDir             string
	UnauthorizedSound    string
	ValidateCertificates bool
	VolumeLevel          float64
)

func init() {
	fs := flag.NewFlagSet("magicband-reader", flag.ExitOnError)
	var (
		apiKey            = fs.String("api-key", "", "The API key to authenticate to rfid-security-svc")
		apiSSLVerify      = fs.String("api-ssl-verify", "ca.pem", "If 'True' or a valid file reference, performs SSL validation, if false, skips validation (this is insecure!).")
		apiUrl            = fs.String("api-url", "https://localhost:5000/api/v1.0", "The rfid-security-svc base URL.")
		authorizedSound   = fs.String("authorized-sound", "authorized.wav", "The name of the sound file played when a band is authorized (relative to sound-dir).")
		brightness        = fs.Int("brightness", 100, "The brightness level of the LEDs. Range of 0 to 255 inclusive")
		configFile        = fs.String("config-file", "/etc/magicband-reader/magicband-reader.yaml", "The YAML configuration file to load.")
		innerRingSize     = fs.Int("inner-ring-size", 20, "The number of LEDs that make up the inner ring.")
		listenAddress     = fs.String("listen-address", "localhost", "The address to listen on, since the listener has no security, it's not recommended to change this value.")
		listenPort        = fs.Int("listen-port", 8080, "The port number to listen for requests for UID (e.g. from rfid-security-svc)")
		logLevel          = fs.String("log-level", "info", "One of: debug, info, warning, error fatal.")
		logReportCaller   = fs.Bool("log-report-caller", false, "Includes the calling function, file, and line number (caller) in log lines. Only works when log-level = trace")
		outerRingSize     = fs.Int("outer-ring-size", 40, "The number of LEDs that make up the outer ring.")
		permission        = fs.String("permission", "Open Door", "The name of the permission to validate before authorizing.")
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
		panic(err)
	}

	validateCertificates, err := isBool(*apiSSLVerify)
	if err != nil {
		// This is not a boolean, must be a file
		if err := validateFileExists(*apiSSLVerify, "api-ssl-verify"); err != nil {
			panic(err)
		}
	}

	url, err := url.Parse(ensureEndsWith(*apiUrl, "/"))
	if err != nil {
		panic(err)
	}

	ApiKey = *apiKey
	ApiUrl = url
	AuthorizedSound = *authorizedSound
	Brightness = *brightness
	CaCertFile = *apiSSLVerify
	ConfigFile = *configFile
	InnerRingSize = *innerRingSize
	ListenAddress = *listenAddress
	ListenPort = *listenPort
	OuterRingSize = *outerRingSize
	Permission = *permission
	ReadSound = *readSound
	SoundDir = *soundDir
	UnauthorizedSound = *unauthorizedSound
	VolumeLevel = *volumeLevel
	ValidateCertificates = validateCertificates

	level, err := validate(*logLevel)
	if err != nil {
		panic(err)
	}

	configureLog(level, *logReportCaller)
	logConfig(*configFile, level, *logReportCaller)
}

func logConfig(configFile string, level log.Level, logReportCaller bool) {
	log.Debug("api-key: <redacted>")
	log.Debugf("api-ssl-verify: %v", CaCertFile)
	log.Debugf("api-url: %v", ApiUrl)
	log.Debugf("authorized-sound: %v", AuthorizedSound)
	log.Debugf("brightness: %v", Brightness)
	log.Debugf("config-file: %v", configFile)
	log.Debugf("inner-ring-size: %v", InnerRingSize)
	log.Debugf("listen-address: %v", ListenAddress)
	log.Debugf("listen-port: %v", ListenPort)
	log.Debugf("log-level: %v", level)
	log.Debugf("log-report-caller: %v", logReportCaller)
	log.Debugf("outer-ring-size: %v", OuterRingSize)
	log.Debugf("permission: %v", Permission)
	log.Debugf("read-sound: %v", ReadSound)
	log.Debugf("sound-dir: %v", SoundDir)
	log.Debugf("unauthorized-sound: %v", UnauthorizedSound)
	log.Debugf("volume-level: %v", VolumeLevel)
}

func validate(logLevel string) (log.Level, error) {
	if err := validateFileExists(path.Join(SoundDir, AuthorizedSound), "authorized-sound"); err != nil {
		return log.WarnLevel, err
	}

	if err := validateIntRange(Brightness, 0, 255, "brightness"); err != nil {
		return log.WarnLevel, err
	}

	if err := validateFileExists(path.Join(SoundDir, ReadSound), "read-sound"); err != nil {
		return log.WarnLevel, err
	}

	if err := validateFileExists(SoundDir, "sound-dir"); err != nil {
		return log.WarnLevel, err
	}

	if err := validateFileExists(path.Join(SoundDir, UnauthorizedSound), "unauthorized-sound"); err != nil {
		return log.WarnLevel, err
	}

	level, err := validateLogLevel(logLevel, "log-level")
	if err != nil {
		return log.WarnLevel, err
	}

	return level, nil
}

func validateIntRange(value int, low int, high int, name string) error {
	if value < low || value > high {
		return errors.New(fmt.Sprintf("Invalid value for %v: '%v'. Must be between %v and %v inclusive.", name, value, low, high))
	}
	return nil
}

func validateFileExists(file string, name string) error {
	if _, err := os.Stat(file); errors.Is(err, fs.ErrNotExist) {
		return errors.New(fmt.Sprintf("Invalid value for %v: '%v'. %v", name, file, err))
	}
	return nil
}

func validateLogLevel(level string, name string) (log.Level, error) {
	logLevel, err := log.ParseLevel(level)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("Invalid value for %v: '%v'. %v", name, level, err))
	}
	return logLevel, err
}

func configureLog(level log.Level, reportCaller bool) {
	log.SetLevel(level)
	log.SetFormatter(&log.TextFormatter{
		DisableLevelTruncation: true,
		PadLevelText:           true,
		FullTimestamp:          true,
		TimestampFormat:        "2006-01-02 15:04:05.000 -0700",
	})
	if level == log.TraceLevel {
		log.SetReportCaller(reportCaller)
	}
}

func ensureEndsWith(str string, suffix string) string {
	if strings.HasSuffix(str, suffix) {
		return str
	}
	return str + suffix
}

func isBool(str string) (bool, error) {
	result, err := strconv.ParseBool(str)
	if err != nil {
		return false, errors.New(fmt.Sprintf("String '%v' is not a boolean value", str))
	}
	return result, nil
}
