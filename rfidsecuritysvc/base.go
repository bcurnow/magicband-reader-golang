package rfidsecuritysvc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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

func New(apiKey string, caCertFile string, apiUrl *url.URL, validateCertificates bool) (Service, error) {
	log.Trace("Creating new rfidsecuritysvc")
	transport, err := createTransport(caCertFile, validateCertificates, apiKey)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout:   60 * time.Second,
		Transport: transport,
	}

	s := &service{
		apiUrl: apiUrl,
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
		return errors.New(fmt.Sprintf("Bad response from '%v', expected %v but received %v", url, requiredStatusCode, response.StatusCode))
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
