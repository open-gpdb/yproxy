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
	"github.com/yezzey-gp/yproxy/pkg/metrics"
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
		defer func() { _ = listener.Close() }()
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
	var mws *metrics.MetricsWebServer

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
			defer func() { _ = clConn.Close() }()

			_, _ = clConn.Write([]byte("Hello from stats server!!\n"))
			_, _ = clConn.Write([]byte("Client id | Optype | External Path \n"))

			if err := i.pool.ClientPoolForeach(func(cl client.YproxyClient) error {
				_, err := clConn.Write(fmt.Appendf(nil, "%v | %v | %v\n", cl.ID(), cl.OPType(), cl.ExternalFilePath()))
				return err
			}); err != nil {
				ylogger.Zero.Error().Err(err).Msg("failed to write client stats data")
			}
		})
	}

	/* dispatch /debug/pprof server */
	if instanceCnf.DebugPort != 0 {
		debugAddr := fmt.Sprintf("[::1]:%d", instanceCnf.DebugPort)
		dws = NewDebugWebServer(debugAddr)
	}

	if instanceCnf.MetricsPort != 0 {
		metricsAddr := fmt.Sprintf("[::1]:%d", instanceCnf.MetricsPort)
		mws = metrics.NewMetricsWebServer(metricsAddr)
		if err := mws.Serve(); err != nil {
			ylogger.Zero.Error().Err(err).Msg("failed to start metrics server")
		}
	}

	s, err := storage.NewStorage(&instanceCnf.StorageCnf, "yezzey")
	if err != nil {
		return err
	}

	bs, err := storage.NewStorage(&instanceCnf.BackupStorageCnf, "backup")
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
		defer func() { _ = clConn.Close() }()
		ycl := client.NewYClient(clConn)
		if err := i.pool.Put(ycl); err != nil {
			// ?? wtf
			ylogger.Zero.Debug().Uint("id", ycl.ID()).Err(err).Msg("error puting client to pool")
		}
		if err := proc.ProcConn(s, bs, cr, ycl, &instanceCnf.VacuumCnf); err != nil {
			ylogger.Zero.Debug().Uint("id", ycl.ID()).Err(err).Msg("error serving client")
		}
		if _, err := i.pool.Pop(ycl.ID()); err != nil {
			// ?? wtf
			ylogger.Zero.Debug().Uint("id", ycl.ID()).Err(err).Msg("error erasing client from pool")
		}
	})

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
		defer func() { _ = clConn.Close() }()
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
			_ = ycl.ReplyError(fmt.Errorf("wrong message type"), "")

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
	_ = notifier.Ready()

	go func() {
		for {
			_ = notifier.Notify()
			time.Sleep(sdnotifier.Timeout)
		}
	}()

	<-ctx.Done()
	activeConnections.Wait()
	return nil
}

func reusePort(network, address string, conn syscall.RawConn) error {
	return conn.Control(func(descriptor uintptr) {
		_ = syscall.SetsockoptInt(int(descriptor), unix.SOL_SOCKET, unix.SO_REUSEADDR|unix.SO_REUSEPORT, 1)
	})
}
