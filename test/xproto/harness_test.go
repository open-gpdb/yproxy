package xproto

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yezzey-gp/yproxy/config"
	"github.com/yezzey-gp/yproxy/pkg/client"
	pio "github.com/yezzey-gp/yproxy/pkg/io"
	"github.com/yezzey-gp/yproxy/pkg/message"
	"github.com/yezzey-gp/yproxy/pkg/proc"
	"github.com/yezzey-gp/yproxy/pkg/storage"
	"github.com/yezzey-gp/yproxy/pkg/ylogger"
)

func TestMain(m *testing.M) {
	ylogger.ReloadLogger("", "disabled")
	os.Exit(m.Run())
}

type wireMessage interface {
	Encode() []byte
}

// Request are frontend messages
// Response are backend messages
type MessageGroup struct {
	Name     string
	Request  []wireMessage
	Response []wireMessage
}

type server struct {
	sockPath string
	st       storage.StorageInteractor
	ln       net.Listener
}

func newServer(t *testing.T) *server {
	t.Helper()

	root := filepath.Join(t.TempDir(), "storage") + string(os.PathSeparator)
	require.NoError(t, os.MkdirAll(root, 0700))

	f, err := os.CreateTemp("", "yproxy-xproto-*.sock")
	require.NoError(t, err)
	sockPath := f.Name()
	_ = f.Close()
	_ = os.Remove(sockPath)

	st, err := storage.NewStorage(&config.Storage{
		StorageType:   "fs",
		StoragePrefix: root,
		StorageBucket: "test-bucket",
	}, "yezzey")
	require.NoError(t, err)

	ln, err := net.Listen("unix", sockPath)
	require.NoError(t, err)

	s := &server{sockPath: sockPath, st: st, ln: ln}
	go s.serve()

	t.Cleanup(func() {
		_ = ln.Close()
		_ = os.Remove(sockPath)
	})
	return s
}

func (s *server) serve() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return
		}
		go func(c net.Conn) {
			_ = proc.ProcConn(&proc.ProtoMgrImpl{}, s.st, s.st, nil, client.NewYClient(c), &config.Vacuum{})
		}(conn)
	}
}

func (s *server) dial(t *testing.T) net.Conn {
	t.Helper()
	conn, err := net.DialTimeout("unix", s.sockPath, 5*time.Second)
	require.NoError(t, err)
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	return conn
}

// TODO: maybe reuse kernel logic here.
func readMessage(t *testing.T, conn net.Conn) wireMessage {
	t.Helper()

	tp, body, err := pio.NewProtoReader(client.NewYClient(conn)).ReadPacket()
	require.NoError(t, err)

	switch tp {
	case message.MessageTypeReadyForQuery:
		return message.NewReadyForQueryMessage()
	case message.MessageTypeObjectMeta:
		m := &message.ObjectInfoMessage{}
		m.Decode(body)
		return m
	case message.MessageTypeError:
		m := &message.ErrorMessage{}
		m.Decode(body)
		return m
	default:
		t.Fatalf("unexpected response message type %s", tp)
		return nil
	}
}

func protoTestRunner(t *testing.T, s *server, groups []MessageGroup) {
	t.Helper()

	for i, group := range groups {
		name := group.Name
		if name == "" {
			name = fmt.Sprintf("group_%d", i)
		}

		t.Run(name, func(t *testing.T) {
			conn := s.dial(t)
			defer func() { _ = conn.Close() }()

			for _, req := range group.Request {
				_, err := conn.Write(req.Encode())
				require.NoError(t, err)
			}
			for _, want := range group.Response {
				assert.Equal(t, want, readMessage(t, conn))
			}
		})
	}
}

func readRaw(t *testing.T, conn net.Conn, n int) []byte {
	t.Helper()
	buf := make([]byte, n)
	_, err := io.ReadFull(conn, buf)
	require.NoError(t, err)
	return buf
}
