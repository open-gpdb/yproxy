package proc

import (
	"fmt"
	"io"
)

type contentReaderFactory func() (io.Reader, io.Closer, error)

func copyWithFullRestart(dst io.Writer, open contentReaderFactory, skip int64,
	attempts int) (int64, error) {
	var delivered int64
	for attempt := 0; attempt < attempts; attempt++ {
		reader, closer, err := open()
		if err != nil {
			return delivered, err
		}

		_, err = io.CopyN(io.Discard, reader, skip+delivered)
		if err == nil {
			var copied int64
			copied, err = io.Copy(dst, reader)
			delivered += copied
		}
		closeErr := closer.Close()
		if err == nil {
			return delivered, closeErr
		}
	}
	return delivered, fmt.Errorf("failed to read after %d full restart attempts", attempts)
}
