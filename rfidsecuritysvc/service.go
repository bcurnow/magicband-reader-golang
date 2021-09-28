package rfidsecuritysvc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/event"
)

type Service interface {
	Authorized(event event.Event, permission string) bool
	Sounds() SoundService
}

type service struct {
	apiUrl *url.URL
	client *http.Client
}

func New(apiKey string, apiSSLVerify string, apiUrl string) (Service, error) {
	log.Trace("Creating new rfidsecuritysvc")

	url, err := url.Parse(ensureEndsWith(apiUrl, "/"))
	if err != nil {
		return nil, err
	}

	transport, err := createTransport(apiSSLVerify, apiKey)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout:   60 * time.Second,
		Transport: transport,
	}

	s := &service{
		apiUrl: url,
		client: client,
	}
	return s, nil
}

func (s *service) Get(urlString string, requiredStatusCode int, jsonStruct interface{}) error {
	url, err := s.apiUrl.Parse(urlString)
	if err != nil {
		return err
	}

	response, err := s.client.Get(url.String())
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != requiredStatusCode {
		return fmt.Errorf("Bad response from '%v', expected %v but received %v", url, requiredStatusCode, response.StatusCode)
	}

	if jsonStruct != nil {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(body, jsonStruct); err != nil {
			return err
		}
	}
	return nil
}

func ensureEndsWith(str string, suffix string) string {
	if strings.HasSuffix(str, suffix) {
		return str
	}
	return str + suffix
}
