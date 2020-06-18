package httpserver

import (
	"crypto/tls"
	"errors"
	"time"

	"github.com/go-joe/joe"
	"go.uber.org/zap"
)

// An Option is used to configure the HTTP server.
type Option func(*config) error

type config struct {
	logger            *zap.Logger
	listenAddr        string
	readTimeout       time.Duration
	writeTimeout      time.Duration
	tlsConf           *tls.Config
	certFile, keyFile string
	trustedHeader     string
}

func newConf(listenAddr string, joeConf *joe.Config, opts []Option) (config, error) {
	conf := config{listenAddr: listenAddr}
	for _, opt := range opts {
		err := opt(&conf)
		if err != nil {
			return conf, err
		}
	}

	if conf.logger == nil {
		conf.logger = joeConf.Logger("http")
	}

	return conf, nil
}

// WithLogger can be used to inject a different logger for the HTTP server.
func WithLogger(logger *zap.Logger) Option {
	return func(conf *config) error {
		conf.logger = logger
		return nil
	}
}

// WithTLS enables serving HTTP requests via TLS.
func WithTLS(certFile, keyFile string) Option {
	return func(conf *config) error {
		if certFile == "" {
			return errors.New("path to certificate file cannot be empty")
		}
		if keyFile == "" {
			return errors.New("path to private key file cannot be empty")
		}

		conf.certFile = certFile
		conf.keyFile = keyFile

		return nil
	}
}

// WithTLSConfig can be used in combination with the WithTLS(…) option to
// configure the HTTPS server.
func WithTLSConfig(tlsConf *tls.Config) Option {
	return func(conf *config) error {
		conf.tlsConf = tlsConf
		return nil
	}
}

// WithTimeouts sets both the read and write timeout of the HTTP server to the
// same given value.
func WithTimeouts(d time.Duration) Option {
	return func(conf *config) error {
		conf.readTimeout = d
		conf.writeTimeout = d
		return nil
	}
}

// WithReadTimeout sets the servers maximum duration for reading the entire HTTP
// request, including the body.
func WithReadTimeout(d time.Duration) Option {
	return func(conf *config) error {
		conf.readTimeout = d
		return nil
	}
}

// WithWriteTimeout sets the servers maximum duration before timing out writes
// of the HTTP response.
func WithWriteTimeout(d time.Duration) Option {
	return func(conf *config) error {
		conf.writeTimeout = d
		return nil
	}
}

// WithTrustedHeader sets the name of the header from which client IPs will be
// populated.
func WithTrustedHeader(header string) Option {
	return func(conf *config) error {
		conf.trustedHeader = header
		return nil
	}
}
