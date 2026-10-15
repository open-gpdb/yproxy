package proc

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

type failingReader struct {
	reader io.Reader
	err    error
}

type testCloser struct{}

func (testCloser) Close() error { return nil }

func (r *failingReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if err == io.EOF {
		return n, r.err
	}
	return n, err
}

func TestCopyWithFullRestartReplaysDeliveredBytes(t *testing.T) {
	var opens int
	open := func() (io.Reader, io.Closer, error) {
		opens++
		if opens == 1 {
			return &failingReader{reader: strings.NewReader("abcdef"), err: errors.New("unexpected EOF")}, testCloser{}, nil
		}
		return strings.NewReader("abcdef"), testCloser{}, nil
	}

	var dst bytes.Buffer
	n, err := copyWithFullRestart(&dst, open, 0, 2)
	if err != nil {
		t.Fatalf("copyWithFullRestart failed: %v", err)
	}
	if n != 6 || dst.String() != "abcdef" || opens != 2 {
		t.Fatalf("got n=%d data=%q opens=%d", n, dst.String(), opens)
	}
}
