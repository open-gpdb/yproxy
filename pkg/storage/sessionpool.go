package storage

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/yezzey-gp/aws-sdk-go/aws"
	"github.com/yezzey-gp/aws-sdk-go/aws/client"
	"github.com/yezzey-gp/aws-sdk-go/aws/credentials"
	"github.com/yezzey-gp/aws-sdk-go/aws/defaults"
	"github.com/yezzey-gp/aws-sdk-go/aws/request"
	"github.com/yezzey-gp/aws-sdk-go/aws/session"
	"github.com/yezzey-gp/aws-sdk-go/service/s3"
	"github.com/yezzey-gp/yproxy/config"
	"github.com/yezzey-gp/yproxy/pkg/ylogger"

	"golang.org/x/sync/semaphore"
)

type SessionPool interface {
	GetSession(ctx context.Context, cr *config.StorageCreds) (*s3.S3, error)
	StorageUsedConcurrency() int
}

type S3SessionPool struct {
	cnf *config.Storage

	usedConnections atomic.Int32
	sem             *semaphore.Weighted
}

// StorageUsedConcurrency implements SessionPool.
func (sp *S3SessionPool) StorageUsedConcurrency() int {
	return int(sp.usedConnections.Load())
}

func NewSessionPool(cnf *config.Storage) SessionPool {
	return &S3SessionPool{
		cnf: cnf,
		sem: semaphore.NewWeighted(cnf.StorageConcurrency),
	}
}

// TODO : unit tests
func (sp *S3SessionPool) createSession(cr *config.StorageCreds) (*session.Session, error) {
	s, err := session.NewSession(&aws.Config{
		Retryer: client.DefaultRetryer{
			NumMaxRetries: 20,
			MinRetryDelay: time.Second,
			MaxRetryDelay: time.Second * 20,
		},
	})
	if err != nil {
		return nil, err
	}

	provider := &credentials.StaticProvider{Value: credentials.Value{
		AccessKeyID:     cr.AccessKeyId,
		SecretAccessKey: cr.SecretAccessKey,
	}}

	ylogger.Zero.Debug().Str("endpoint", sp.cnf.StorageEndpoint).Msg("acquire external storage session")

	providers := make([]credentials.Provider, 0)
	providers = append(providers, provider)
	providers = append(providers, defaults.CredProviders(s.Config, defaults.Handlers())...)
	newCredentials := credentials.NewCredentials(&credentials.ChainProvider{
		VerboseErrors: aws.BoolValue(s.Config.CredentialsChainVerboseErrors),
		Providers:     providers,
	})

	s.Config.WithRegion(sp.cnf.StorageRegion)
	s.Config.S3ForcePathStyle = aws.Bool(true)

	s.Config.WithEndpoint(sp.cnf.StorageEndpoint)

	if sp.cnf.EndpointSourceHost != "" {
		s.Handlers.Validate.PushBack(func(request *request.Request) {
			endpoint, err := requestEndpoint(sp.cnf.EndpointSourceHost, sp.cnf.EndpointSourcePort)
			if err == nil {
				ylogger.Zero.Debug().Str("endpoint", endpoint).Msg("using requested endpoint")
				host := strings.TrimPrefix(*s.Config.Endpoint, "https://")
				request.HTTPRequest.Host = host
				request.HTTPRequest.Header.Add("Host", host)
				request.HTTPRequest.URL.Host = endpoint
				request.HTTPRequest.URL.Scheme = sp.cnf.EndpointSourceScheme
			} else {
				ylogger.Zero.Debug().Str("endpoint", *s.Config.Endpoint).Msg("using default endpoint")
			}
		})
	}

	s.Config.WithCredentials(newCredentials)
	return s, err
}

func requestEndpoint(endpointSource, port string) (string, error) {
	ylogger.Zero.Debug().Str("source host", endpointSource).Msg("requesting storage endpoint")
	t := http.DefaultTransport
	c := http.DefaultClient
	if tr, ok := t.(*http.Transport); ok {
		tr.DisableKeepAlives = true
		c = &http.Client{Transport: tr}
	}
	resp, err := c.Get(endpointSource)
	if err != nil {
		ylogger.Zero.Error().Err(err).Msg("failed to get S3 endpoint")
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		ylogger.Zero.Error().Int("status code", resp.StatusCode).Msg("endpoint source bad status code")
		return "", fmt.Errorf("endpoint source bad status code %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ylogger.Zero.Error().Err(err).Msg("error reading endpoint source reply")
		return "", err
	}
	if port != "" {
		return net.JoinHostPort(string(body), port), err
	}
	return string(body), err
}

func (s *S3SessionPool) GetSession(ctx context.Context, cr *config.StorageCreds) (*s3.S3, error) {
	_ = s.sem.Acquire(ctx, 1)
	s.usedConnections.Add(1)
	defer func() {
		s.sem.Release(1)
		s.usedConnections.Add(-1)
	}()

	sess, err := s.createSession(cr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new session")
	}
	return s3.New(sess), nil
}
