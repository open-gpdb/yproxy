package core

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sys/unix"

	"github.com/yezzey-gp/yproxy/config"
	"github.com/yezzey-gp/yproxy/pkg/client"
	"github.com/yezzey-gp/yproxy/pkg/clientpool"
	"github.com/yezzey-gp/yproxy/pkg/core/pg"
	"github.com/yezzey-gp/yproxy/pkg/crypt"
	"github.com/yezzey-gp/yproxy/pkg/message"
	"github.com/yezzey-gp/yproxy/pkg/proc"
	"github.com/yezzey-gp/yproxy/pkg/sdnotifier"
	"github.com/yezzey-gp/yproxy/pkg/storage"
	"github.com/yezzey-gp/yproxy/pkg/ylogger"
)

type Instance struct {
	pool clientpool.Pool

	startTs time.Time
}

func NewInstance() *Instance {
	return &Instance{
		pool:    clientpool.NewClientPool(),
		startTs: time.Now(),
	}
}

func (i *Instance) DispatchServer(listener net.Listener, server func(net.Conn)) {
	go func() {
		defer listener.Close()
		for {
			clConn, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					break
				}
				ylogger.Zero.Error().Err(err).Msg("failed to accept connection")
				continue
			}
			ylogger.Zero.Debug().Str("addr", clConn.LocalAddr().String()).Msg("accepted client connection")

			go server(clConn)
		}
	}()
}

func (i *Instance) Run(instanceCnf *config.Instance) error {

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)

	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	var listener net.Listener
	var iclistener net.Listener
	var dws *DebugWebServer

	go func() {
		defer cancelCtx()

		for {
			s := <-sigs
			ylogger.Zero.Info().Str("signal", s.String()).Msg("received signal")

			switch s {
			case syscall.SIGUSR1:
				ylogger.ReloadLogger(instanceCnf.LogPath)
			case syscall.SIGHUP:
				if dws != nil {
					err := dws.ServeFor(time.Duration(instanceCnf.DebugMinutes) * time.Minute)
					if err != nil {
						ylogger.Zero.Error().Err(err).Msg("Error in debug server")
					}
				}
			case syscall.SIGUSR2:
				fallthrough
			case syscall.SIGINT, syscall.SIGTERM:
				// make better
				fallthrough
			default:
				if listener != nil {
					if err := listener.Close(); err != nil {
						ylogger.Zero.Error().Err(err).Msg("failed to close socket")
					}
				}
				if iclistener != nil {
					if err := iclistener.Close(); err != nil {
						ylogger.Zero.Error().Err(err).Msg("failed to close ic socket")
					}
				}
				return
			}
		}
	}()

	/* dispatch statistic server */
	if instanceCnf.StatPort != 0 {
		config := &net.ListenConfig{Control: reusePort}
		statListener, err := config.Listen(context.Background(), "tcp", fmt.Sprintf("localhost:%v", instanceCnf.StatPort))
		if err != nil {
			ylogger.Zero.Error().Err(err).Msg("failed to start socket listener")
			return err
		}

		i.DispatchServer(statListener, func(clConn net.Conn) {
			defer clConn.Close()

			clConn.Write([]byte("Hello from stats server!!\n"))
			clConn.Write([]byte("Client id | Optype | External Path \n"))

			i.pool.ClientPoolForeach(func(cl client.YproxyClient) error {
				_, err := clConn.Write([]byte(fmt.Sprintf("%v | %v | %v\n", cl.ID(), cl.OPType(), cl.ExternalFilePath())))
				return err
			})
		})
	}

	/* dispatch /debug/pprof server */
	if instanceCnf.DebugPort != 0 {
		debugAddr := fmt.Sprintf("[::1]:%d", instanceCnf.DebugPort)
		dws = NewDebugWebServer(debugAddr)
	}

	s, err := storage.NewStorage(
		&instanceCnf.StorageCnf,
	)
	if err != nil {
		return err
	}

	bs, err := storage.NewStorage(&instanceCnf.BackupStorageCnf)
	if err != nil {
		return err
	}

	if instanceCnf.PsqlPort != 0 {
		config := &net.ListenConfig{Control: reusePort}
		psqlListener, err := config.Listen(context.Background(), "tcp", fmt.Sprintf("localhost:%v", instanceCnf.PsqlPort))
		if err != nil {
			ylogger.Zero.Error().Err(err).Msg("failed to start socket listener")
			return err
		}

		i.DispatchServer(psqlListener, func(c net.Conn) {
			pg.PostgresIface(c, i.pool, i.startTs, s)
		})
	}

	activeConnections := sync.WaitGroup{}
	retryTicker := time.NewTicker(time.Second)

	for listener == nil {
		select {
		case <-retryTicker.C:
			listener, err = net.Listen("unix", instanceCnf.SocketPath)
			if err != nil {
				if errors.Is(err, syscall.EADDRINUSE) {
					continue
				}
				ylogger.Zero.Error().Err(err).Msg("failed to start socket listener")
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	ylogger.Zero.Info().Str("socket", instanceCnf.SocketPath).Msg("yproxy is listening unix socket")

	var cr crypt.Crypter = nil
	if instanceCnf.CryptoCnf.GPGKeyPath != "" {
		cr, err = crypt.NewCrypto(&instanceCnf.CryptoCnf)
	}

	i.DispatchServer(listener, func(clConn net.Conn) {
		activeConnections.Add(1)
		defer activeConnections.Done()
		defer clConn.Close()
		ycl := client.NewYClient(clConn)
		i.pool.Put(ycl)
		if err := proc.ProcConn(s, bs, cr, ycl, &instanceCnf.VacuumCnf); err != nil {
			ylogger.Zero.Debug().Uint("id", ycl.ID()).Err(err).Msg("error serving client")
		}
		_, err := i.pool.Pop(ycl.ID())
		if err != nil {
			// ?? wtf
			ylogger.Zero.Debug().Uint("id", ycl.ID()).Err(err).Msg("error erasing client from pool")
		}
	})

	if err != nil {
		return err
	}

	for iclistener == nil {
		select {
		case <-retryTicker.C:
			iclistener, err = net.Listen("unix", instanceCnf.InterconnectSocketPath)
			if err != nil {
				if errors.Is(err, syscall.EADDRINUSE) {
					continue
				}
				ylogger.Zero.Error().Err(err).Msg("failed to start socket listener")
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	ylogger.Zero.Debug().Msg("try to start interconnect socket listener")
	if err != nil {
		ylogger.Zero.Error().Err(err).Msg("failed to start interconnect socket listener")
		return err
	}

	i.DispatchServer(iclistener, func(clConn net.Conn) {
		activeConnections.Add(1)
		defer activeConnections.Done()
		defer clConn.Close()
		ycl := client.NewYClient(clConn)
		r := proc.NewProtoReader(ycl)

		mt, _, err := r.ReadPacket()

		if err != nil {
			ylogger.Zero.Error().Err(err).Msg("failed to accept interconnection")
		}

		switch mt {
		case message.MessageTypeGool:
			msg := message.ReadyForQueryMessage{}
			_, _ = ycl.GetRW().Write(msg.Encode())
		default:
			ycl.ReplyError(fmt.Errorf("wrong message type"), "")

		}
		ylogger.Zero.Debug().Msg("interconnection closed")
	})

	notifier, err := sdnotifier.NewNotifier(instanceCnf.GetSystemdSocketPath(), instanceCnf.SystemdNotificationsDebug)
	if err != nil {
		ylogger.Zero.Error().Err(err).Msg("failed to initialize systemd notifier")
		if instanceCnf.SystemdNotificationsDebug {
			return err
		}
	}
	notifier.Ready()

	go func() {
		for {
			notifier.Notify()
			time.Sleep(sdnotifier.Timeout)
		}
	}()

	<-ctx.Done()
	activeConnections.Wait()
	return nil
}

func reusePort(network, address string, conn syscall.RawConn) error {
	return conn.Control(func(descriptor uintptr) {
		syscall.SetsockoptInt(int(descriptor), unix.SOL_SOCKET, unix.SO_REUSEADDR|unix.SO_REUSEPORT, 1)
	})
}
