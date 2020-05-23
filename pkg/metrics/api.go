package metrics

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-rootcerts"
)

type Config struct {
	Address    string
	TLSConfig  *TLSConfig
	httpClient *http.Client
}

type TLSConfig struct {
	CACert        string
	ClientCert    string
	ClientCertKey string
	Insecure      bool
}

type Client struct {
	config Config
}

func DefaultConfig() *Config {
	config := Config{
		Address:    "http://127.0.0.1:4646",
		TLSConfig:  &TLSConfig{},
		httpClient: cleanhttp.DefaultClient(),
	}

	if addr := os.Getenv("NOMAD_ADDR"); addr != "" {
		config.Address = addr
	}

	// Read TLS specific env vars
	if v := os.Getenv("NOMAD_CACERT"); v != "" {
		config.TLSConfig.CACert = v
	}
	if v := os.Getenv("NOMAD_CLIENT_CERT"); v != "" {
		config.TLSConfig.ClientCert = v
	}
	if v := os.Getenv("NOMAD_CLIENT_KEY"); v != "" {
		config.TLSConfig.ClientCertKey = v
	}
	if v := os.Getenv("NOMAD_SKIP_VERIFY"); v != "" {
		if insecure, err := strconv.ParseBool(v); err == nil {
			config.TLSConfig.Insecure = insecure
		}
	}

	transport := config.httpClient.Transport.(*http.Transport)
	transport.TLSHandshakeTimeout = 10 * time.Second
	transport.TLSClientConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	return &config
}

func NewClient(config *Config) (*Client, error) {
	if _, err := url.Parse(config.Address); err != nil {
		return nil, fmt.Errorf("invalid address '%s': %v", config.Address, err)
	}

	// Configure the TLS configurations
	if err := config.ConfigureTLS(); err != nil {
		return nil, err
	}

	return &Client{
		config: *config,
	}, nil
}

func (c *Config) ConfigureTLS() error {
	if c.TLSConfig == nil {
		return nil
	}

	var clientCert tls.Certificate

	foundClientCert := false
	if c.TLSConfig.ClientCert != "" || c.TLSConfig.ClientCertKey != "" {
		if c.TLSConfig.ClientCert != "" && c.TLSConfig.ClientCertKey != "" {
			var err error
			clientCert, err = tls.LoadX509KeyPair(c.TLSConfig.ClientCert, c.TLSConfig.ClientCertKey)
			if err != nil {
				return err
			}
			foundClientCert = true
		} else {
			return fmt.Errorf("client cert and client key must be provided")
		}
	}

	clientTLSConfig := c.httpClient.Transport.(*http.Transport).TLSClientConfig
	rootConfig := &rootcerts.Config{CAPath: c.TLSConfig.CACert}

	if err := rootcerts.ConfigureTLS(clientTLSConfig, rootConfig); err != nil {
		return err
	}

	clientTLSConfig.InsecureSkipVerify = c.TLSConfig.Insecure

	if foundClientCert {
		clientTLSConfig.Certificates = []tls.Certificate{clientCert}
	}
	return nil
}

func (c *Client) get(endpoint string, out interface{}) error {
	r, err := c.newRequest(http.MethodGet, endpoint)
	if err != nil {
		return err
	}

	resp, err := c.doRequest(r)
	resp, err = requireOK(resp, err, http.StatusOK)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := decodeBody(&resp.Body, out); err != nil {
		return err
	}
	return nil
}

func (c *Client) newRequest(method, path string) (*http.Request, error) {
	base, _ := url.Parse(c.config.Address)

	// Create the HTTP request
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		return nil, err
	}

	req.URL.Host = base.Host
	req.URL.Scheme = base.Scheme
	req.Host = base.Host

	return req, nil
}

func (c *Client) doRequest(r *http.Request) (*http.Response, error) {
	return c.config.httpClient.Do(r)
}

func decodeBody(body *io.ReadCloser, out interface{}) error {
	return json.NewDecoder(*body).Decode(out)
}

func requireOK(resp *http.Response, e error, expected int) (*http.Response, error) {
	if e != nil {
		if resp != nil {
			_ = resp.Body.Close()
		}
		return nil, e
	}
	if resp.StatusCode != expected {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, resp.Body)
		_ = resp.Body.Close()
		return nil, fmt.Errorf(strings.TrimSpace(fmt.Sprintf("unexpected response code %d: %s", resp.StatusCode, buf.Bytes())))
	}
	return resp, nil
}
