package rfidsecuritysvc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type authorizingTransport struct {
	transport http.RoundTripper
	apiKey    string
}

func (t *authorizingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("X-RFIDSECURITYSVC-API-KEY", url.QueryEscape(t.apiKey))
	return t.transport.RoundTrip(req)
}

func createTransport(apiSSLVerify string, apiKey string) (*authorizingTransport, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	tlsConfig := &tls.Config{}

	// apiSSLVerify can either be a boolean false or a file name
	validateCertificates, err := isBool(apiSSLVerify)
	if err != nil {
		// this is a file name
		caCertPool := x509.NewCertPool()
		caCerts, err := os.ReadFile(apiSSLVerify)
		if err != nil {
			return nil, err
		}
		caCertPool.AppendCertsFromPEM(caCerts)
		tlsConfig.RootCAs = caCertPool
	} else {
		if validateCertificates {
			return nil, fmt.Errorf("apiSSLVerify can not be set to true")
		} else {
			tlsConfig.InsecureSkipVerify = true
		}
	}

	transport.TLSClientConfig = tlsConfig

	authorizingTransport := &authorizingTransport{
		transport: transport,
		apiKey:    apiKey,
	}

	return authorizingTransport, nil
}

func isBool(str string) (bool, error) {
	result, err := strconv.ParseBool(str)
	if err != nil {
		return false, fmt.Errorf("String '%v' is not a boolean value", str)
	}
	return result, nil
}
