package core

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yezzey-gp/yproxy/config"
	"github.com/yezzey-gp/yproxy/pkg/client"
	"github.com/yezzey-gp/yproxy/pkg/clientpool"
	"github.com/yezzey-gp/yproxy/pkg/crypt"
	"github.com/yezzey-gp/yproxy/pkg/message"
	"github.com/yezzey-gp/yproxy/pkg/proc"
	"github.com/yezzey-gp/yproxy/pkg/sdnotifier"
	"github.com/yezzey-gp/yproxy/pkg/storage"
	"github.com/yezzey-gp/yproxy/pkg/ylogger"
)

type Instance struct {
	pool clientpool.Pool
}

func NewInstance() *Instance {
	return &Instance{
		pool: clientpool.NewClientPool(),
	}
}

func (i *Instance) DispatchServer(listener net.Listener, server func(net.Conn)) {
	go func() {
		defer listener.Close()
		for {
			clConn, err := listener.Accept()
			if err != nil {
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

	go func() {
		defer os.Remove(instanceCnf.SocketPath)

		defer os.Remove(instanceCnf.InterconnectSocketPath)
		defer cancelCtx()

		for {
			s := <-sigs
			ylogger.Zero.Info().Str("signal", s.String()).Msg("received signal")

			switch s {
			case syscall.SIGUSR1:
				ylogger.ReloadLogger(instanceCnf.LogPath)
			case syscall.SIGUSR2:
				return
			case syscall.SIGHUP:
				// reread config file

			case syscall.SIGINT, syscall.SIGTERM:

				// make better
				return
			default:
				return
			}
		}
	}()

	/* dispatch statistic server */
	if instanceCnf.StatPort != 0 {
		statListener, err := net.Listen("tcp", fmt.Sprintf("localhost:%v", instanceCnf.StatPort))
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

	if instanceCnf.PsqlPort != 0 {
		psqlListener, err := net.Listen("tcp", fmt.Sprintf("localhost:%v", instanceCnf.PsqlPort))
		if err != nil {
			ylogger.Zero.Error().Err(err).Msg("failed to start socket listener")
			return err
		}

		i.DispatchServer(psqlListener, PostgresIface)
	}

	/* dispatch statistic server */
	go func() {

		listener, err := net.Listen("unix", instanceCnf.InterconnectSocketPath)

		ylogger.Zero.Debug().Msg("try to start interconnect socket listener")
		if err != nil {
			ylogger.Zero.Error().Err(err).Msg("failed to start interconnect socket listener")
			return
		}
		defer listener.Close()

		for {
			clConn, err := listener.Accept()
			if err != nil {
				ylogger.Zero.Error().Err(err).Msg("failed to accept interconnection")
				continue
			}
			ylogger.Zero.Debug().Str("addr", clConn.LocalAddr().String()).Msg("accepted client interconnection")

			ycl := client.NewYClient(clConn)
			r := proc.NewProtoReader(ycl)

			mt, _, err := r.ReadPacket()

			if err != nil {
				ylogger.Zero.Error().Err(err).Msg("failed to accept interconnection")
				continue
			}

			switch mt {
			case message.MessageTypeGool:
				msg := message.ReadyForQueryMessage{}
				_, _ = ycl.GetRW().Write(msg.Encode())
			default:
				ycl.ReplyError(fmt.Errorf("wrong message type"), "")

			}

			clConn.Close()
			ylogger.Zero.Debug().Msg("interconnection closed")
		}
	}()

	listener, err := net.Listen("unix", instanceCnf.SocketPath)
	if err != nil {
		ylogger.Zero.Error().Err(err).Msg("failed to start socket listener")
		return err
	}

	s, err := storage.NewStorage(
		&instanceCnf.StorageCnf,
	)
	if err != nil {
		return err
	}
	var cr crypt.Crypter = nil
	if instanceCnf.CryptoCnf.GPGKeyPath != "" {
		cr, err = crypt.NewCrypto(&instanceCnf.CryptoCnf)
	}

	i.DispatchServer(listener, func(clConn net.Conn) {
		defer clConn.Close()
		ycl := client.NewYClient(clConn)
		i.pool.Put(ycl)
		if err := proc.ProcConn(s, cr, ycl); err != nil {
			ylogger.Zero.Debug().Uint("id", ycl.ID()).Err(err).Msg("got error serving client")
		}
		_, err := i.pool.Pop(ycl.ID())
		if err != nil {
			// ?? wtf
			ylogger.Zero.Debug().Uint("id", ycl.ID()).Err(err).Msg("got error erasing client from pool")
		}
	})

	if err != nil {
		return err
	}

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
	return nil
}
