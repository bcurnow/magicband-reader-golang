package rfidsecuritysvc

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/url"
	"os"
)

type authorizingTransport struct {
	transport http.RoundTripper
	apiKey    string
}

func (t *authorizingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("X-RFIDSECURITYSVC-API-KEY", url.QueryEscape(t.apiKey))
	return t.transport.RoundTrip(req)
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
