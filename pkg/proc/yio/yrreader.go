package yio

import (
	"fmt"
	"io"
	"time"

	"github.com/yezzey-gp/yproxy/pkg/client"
	"github.com/yezzey-gp/yproxy/pkg/settings"
	"github.com/yezzey-gp/yproxy/pkg/storage"
	"github.com/yezzey-gp/yproxy/pkg/ylogger"
)

type RestartReader interface {
	io.ReadCloser
	Restart(offsetStart int64) error
}

type YRestartReader struct {
	underlying io.ReadCloser
	s          storage.StorageInteractor
	name       string
	settings   []settings.StorageSettings
}

// Close implements RestartReader.
func (y *YRestartReader) Close() error {
	if y.underlying != nil {
		return y.underlying.Close()
	}
	return nil
}

// Read implements RestartReader.
func (y *YRestartReader) Read(p []byte) (n int, err error) {
	return y.underlying.Read(p)
}

func NewRestartReader(s storage.StorageInteractor,
	name string, setts []settings.StorageSettings) RestartReader {

	return &YRestartReader{
		s:        s,
		name:     name,
		settings: setts,
	}
}

func (y *YRestartReader) Restart(offsetStart int64) error {
	if y.underlying != nil {
		_ = y.underlying.Close()
	}
	if offsetStart == 0 {
		ylogger.Zero.Debug().Str("object-path", y.name).Msg("cat object with offset")
	} else {
		ylogger.Zero.Error().Str("object-path", y.name).Int64("offset", offsetStart).Msg("cat object with offset after possible error")
	}
	r, err := y.s.CatFileFromStorage(y.name, offsetStart, y.settings)
	if err != nil {
		return err
	}

	y.underlying = r

	return nil
}

type YproxyRetryReader struct {
	io.ReadCloser
	underlying RestartReader

	selfCl client.YproxyClient

	offsetReached int64
	retryLimit    int
	needReacquire bool
}

// Close implements io.ReadCloser.
func (y *YproxyRetryReader) Close() error {
	err := y.underlying.Close()
	if err != nil {
		ylogger.Zero.Error().Err(err).Msg("encounter close error")
	}
	return err
}

// Read implements io.ReadCloser.
func (y *YproxyRetryReader) Read(p []byte) (int, error) {

	for retry := range y.retryLimit {

		if y.needReacquire {

			err := y.underlying.Restart(y.offsetReached)

			if err != nil {
				// log error and continue.
				// Try to mitigate overload problems with random sleep
				ylogger.Zero.Error().Err(err).Int("offset reached", int(y.offsetReached)).Int("retry count", int(retry)).Msg("failed to reacquire external storage connection, wait and retry")

				time.Sleep(time.Second)
				continue
			}

			y.needReacquire = false
		}

		n, err := y.underlying.Read(p)
		if err == io.EOF {
			return n, err
		}
		if err != nil || n < 0 {
			ylogger.Zero.Error().Err(err).Int("offset reached", int(y.offsetReached)).Int("retry count", int(retry)).Msg("encounter read error")

			// what if close failed?
			_ = y.underlying.Close()

			// try to reacquire connection to external storage and continue read
			// from previously reached point

			y.needReacquire = true
			continue
		} else {
			y.offsetReached += int64(n)

			y.selfCl.SetByteOffset(y.offsetReached)

			return n, err
		}
	}
	return -1, fmt.Errorf("failed to unpload within retries")
}

const (
	defaultRetryLimit = 100
)

func NewYRetryReader(r RestartReader, selfCl client.YproxyClient) io.ReadCloser {
	return &YproxyRetryReader{
		underlying:    r,
		retryLimit:    defaultRetryLimit,
		selfCl:        selfCl,
		offsetReached: 0,
		needReacquire: true, /* do initial storage request */
	}
}

var _ io.ReadCloser = &YproxyRetryReader{}
