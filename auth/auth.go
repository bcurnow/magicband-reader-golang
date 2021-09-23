package auth

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/event"
)

const (
	permissionUrlFormat = "authorized/media/%v/perm/%v"
)

type Controller interface {
	Authorized(event event.Event) bool
}

type controller struct {
	apiUrl     *url.URL
	client     *http.Client
	permission string
}

type authorizingTransport struct {
	transport http.RoundTripper
	apiKey    string
}

func (t *authorizingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("X-RFIDSECURITYSVC-API-KEY", url.QueryEscape(t.apiKey))
	return t.transport.RoundTrip(req)
}

func NewController(apiKey string, caCertFile string, apiUrl *url.URL, permission string, validateCertificates bool) (Controller, error) {
	log.Trace("Creating new auth.Controller")
	transport, err := createTransport(caCertFile, validateCertificates, apiKey)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout:   60 * time.Second,
		Transport: transport,
	}

	c := &controller{
		apiUrl:     apiUrl,
		client:     client,
		permission: permission,
	}
	return c, nil
}

func (c *controller) Authorized(event event.Event) bool {
	url := fmt.Sprintf(permissionUrlFormat, url.PathEscape(event.UID()), url.PathEscape(c.permission))
	permissionUrl, err := c.apiUrl.Parse(url)
	if err != nil {
		// This is a very strange error because we checked dthe apiUrl during config and the
		// url is based on a format string which is properly escaped, let's just print a warning
		log.Warnf("Error during parse of '%v', returning false from Authorized: %v", url, err)
		return false
	}

	response, err := c.client.Get(permissionUrl.String())
	if err != nil {
		log.Warnf("Error calling '%v', returning false from Authorized: %v", permissionUrl, err)
		return false
	}
	// We don't care about the body, just the status code
	defer response.Body.Close()

	// The service returns 200 (OK) when authorized and 403 (Forbidden) when not authorized
	if response.StatusCode != 200 && response.StatusCode != 403 {
		log.Errorf("Unexpected status code %v from '%v'", response.StatusCode, permissionUrl)
	}
	return response.StatusCode == 200
}

func createTransport(caCertFile string, validateCertificates bool, apiKey string) (*authorizingTransport, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	tlsConfig := &tls.Config{}

	if validateCertificates {
		caCertPool := x509.NewCertPool()
		caCerts, err := os.ReadFile(caCertFile)
		if err != nil {
			return nil, err
		}
		caCertPool.AppendCertsFromPEM(caCerts)
		tlsConfig.RootCAs = caCertPool
	} else {
		tlsConfig.InsecureSkipVerify = true
	}

	transport.TLSClientConfig = tlsConfig

	authorizingTransport := &authorizingTransport{
		transport: transport,
		apiKey:    apiKey,
	}

	return authorizingTransport, nil
}
