package yio_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	mockcl "github.com/yezzey-gp/yproxy/pkg/mock/client"
	mockio "github.com/yezzey-gp/yproxy/pkg/mock/proc/yio"
	"github.com/yezzey-gp/yproxy/pkg/proc/yio"
	"go.uber.org/mock/gomock"
)

func TestYproxyRetryReaderEmpty(t *testing.T) {

	ctrl := gomock.NewController(t)

	rr := mockio.NewMockRestartReader(ctrl)

	ycl := mockcl.NewMockYproxyClient(ctrl)

	yr := yio.NewYRetryReader(rr, ycl)

	buf := []byte{1, 233, 45}

	rr.EXPECT().Restart(int64(0)).Return(nil)
	rr.EXPECT().Read(gomock.Any()).Return(0, io.EOF)
	rr.EXPECT().Close().Times(1)

	_, err := yr.Read(buf)

	assert.Equal(t, io.EOF, err)

	assert.Nil(t, yr.Close())
}

func TestYproxyRetryReaderSimpleRead(t *testing.T) {

	ctrl := gomock.NewController(t)

	rr := mockio.NewMockRestartReader(ctrl)

	ycl := mockcl.NewMockYproxyClient(ctrl)
	yr := yio.NewYRetryReader(rr, ycl)

	ycl.EXPECT().SetByteOffset(int64(3)).Times(1)

	buf := []byte{0, 0, 0}

	rr.EXPECT().Restart(int64(0)).Return(nil)
	rr.EXPECT().Read(gomock.Any()).Do(
		func(rbuf []byte) {
			rbuf[0] = 1
			rbuf[1] = 27
			rbuf[2] = 33
		},
	).Return(3, nil)
	rr.EXPECT().Read(gomock.Any()).Return(0, io.EOF)
	rr.EXPECT().Close().Times(1)

	n, err := yr.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, 3, n)

	assert.Equal(t, buf[0], uint8(1))
	assert.Equal(t, buf[1], uint8(27))
	assert.Equal(t, buf[2], uint8(33))

	_, err = yr.Read(buf)

	assert.Equal(t, io.EOF, err)

	assert.Nil(t, yr.Close())
}

func TestYproxyRetryReaderSimpleReadRetry(t *testing.T) {

	ctrl := gomock.NewController(t)

	rr := mockio.NewMockRestartReader(ctrl)

	ycl := mockcl.NewMockYproxyClient(ctrl)
	yr := yio.NewYRetryReader(rr, ycl)

	ycl.EXPECT().SetByteOffset(int64(3)).Times(1)

	ycl.EXPECT().SetByteOffset(int64(6)).Times(1)

	buf := []byte{0, 0, 0}

	rr.EXPECT().Restart(int64(0)).Return(nil).Times(1)
	rr.EXPECT().Restart(int64(3)).Return(nil).Times(1)
	rr.EXPECT().Read(buf).Do(
		func(rbuf []byte) {
			rbuf[0] = 1
			rbuf[1] = 27
			rbuf[2] = 33
		},
	).Return(3, nil)
	rr.EXPECT().Read(gomock.Any()).Return(0, fmt.Errorf("no"))
	rr.EXPECT().Read(gomock.Any()).Do(
		func(rbuf []byte) {
			rbuf[0] = 1
			rbuf[1] = 27
			rbuf[2] = 33
		},
	).Return(3, nil)
	rr.EXPECT().Read(gomock.Any()).Return(0, io.EOF)

	rr.EXPECT().Close().Times(2)

	n, err := yr.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, 3, n)

	assert.Equal(t, buf[0], uint8(1))
	assert.Equal(t, buf[1], uint8(27))
	assert.Equal(t, buf[2], uint8(33))

	// yr got error, but retries
	buf2 := []byte{0, 0, 0}

	n, err = yr.Read(buf2)
	assert.NoError(t, err)
	assert.Equal(t, 3, n)

	assert.Equal(t, buf2[0], uint8(1))
	assert.Equal(t, buf2[1], uint8(27))
	assert.Equal(t, buf2[2], uint8(33))

	_, err = yr.Read([]byte{0, 0, 0})

	assert.Equal(t, io.EOF, err)

	assert.Nil(t, yr.Close())
}
