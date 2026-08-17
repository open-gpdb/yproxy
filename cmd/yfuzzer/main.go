package main

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/yezzey-gp/yproxy/config"
	"github.com/yezzey-gp/yproxy/pkg"
	"github.com/yezzey-gp/yproxy/pkg/client"
	"github.com/yezzey-gp/yproxy/pkg/core"
	"github.com/yezzey-gp/yproxy/pkg/crypt"
	pio "github.com/yezzey-gp/yproxy/pkg/io"
	"github.com/yezzey-gp/yproxy/pkg/message"
	"github.com/yezzey-gp/yproxy/pkg/object"
	"github.com/yezzey-gp/yproxy/pkg/proc"
	"github.com/yezzey-gp/yproxy/pkg/proc/yio"
	"github.com/yezzey-gp/yproxy/pkg/settings"
	"github.com/yezzey-gp/yproxy/pkg/storage"
	"github.com/yezzey-gp/yproxy/pkg/ylogger"
)

type FuzzerProtoMgr struct {
	proc.ProtoMgrImpl
}

func (f *FuzzerProtoMgr) ProcessListExtended(
	prefix string,
	_ []settings.StorageSettings,
	_ storage.StorageInteractor,
	_ crypt.Crypter,
	ycl client.YproxyClient,
	_ *config.Vacuum,
) error {
	ycl.SetExternalFilePath(prefix)

	ylogger.Zero.Info().Str("prefix", prefix).Msg("yfuzzer: processing list request")

	/* XXX: move to config or something */
	const maxIterations = 100
	const maxBatchSize = 500
	const maxPathLen = 256

	iterations := rand.Intn(maxIterations) + 1
	ylogger.Zero.Info().Int("iterations", iterations).Msg("yfuzzer: sending random listing messages")

	for i := 0; i < iterations; i++ {
		batchSize := rand.Intn(maxBatchSize) + 1
		metas := make([]*object.ObjectInfo, batchSize)
		for j := range metas {
			pathLen := rand.Intn(maxPathLen) + 1
			pathBytes := make([]byte, pathLen)
			for k := range pathBytes {
				pathBytes[k] = byte(rand.Intn(95) + 32)
			}
			metas[j] = &object.ObjectInfo{
				Path: string(pathBytes),
				Size: rand.Int63(),
			}
		}

		msgBytes := message.NewObjectMetaMessage(metas).Encode()

		for len(msgBytes) > 0 {
			var writeChunk []byte

			/* Dummy for short tails */
			if len(msgBytes) < 2 {
				writeChunk = msgBytes
				msgBytes = nil
			} else {
				currLen := rand.Intn(len(msgBytes)) + 1
				writeChunk = msgBytes[:currLen]
				msgBytes = msgBytes[currLen:]
			}

			if _, err := ycl.GetRW().Write(writeChunk); err != nil {
				_ = ycl.ReplyError(err, "yfuzzer: failed to write listing batch")
				return err
			}
		}
	}

	/* XXX: we can fuzz this too */
	if _, err := ycl.GetRW().Write(message.NewReadyForQueryMessage().Encode()); err != nil {
		_ = ycl.ReplyError(err, "yfuzzer: failed to write ready for query")
		return err
	}

	return nil
}

// ProcessCatExtended proxies the real object read from storage, but writes
// the contents back to the client in randomly-sized chunks, mimicking the
// fragmented writes used for listing batches.
func (f *FuzzerProtoMgr) ProcessCatExtended(
	s storage.StorageInteractor,
	_ *pio.ProtoReader,
	name string,
	decrypt bool,
	kek bool,
	startOffset uint64,
	settings []settings.StorageSettings,
	cr crypt.Crypter,
	ycl client.YproxyClient,
) error {
	ycl.SetExternalFilePath(name)

	yr := yio.NewYRetryReader(yio.NewRestartReader(s, name, settings), ycl)

	var contentReader io.Reader = yr
	defer func() { _ = yr.Close() }()

	if decrypt {
		if cr == nil {
			err := fmt.Errorf("failed to decrypt object, decrypter not configured")
			ylogger.Zero.Error().Err(err).Msg("yfuzzer: cat failed")
			return err
		}
		ylogger.Zero.Debug().Str("object-path", name).Msg("yfuzzer: decrypt object")
		var err error
		contentReader, err = cr.Decrypt(yr)
		if err != nil {
			ylogger.Zero.Error().Err(err).Msg("yfuzzer: failed to decrypt object")
			return err
		}
	}

	if kek {
		err := fmt.Errorf("KEK is currently unsupported")
		ylogger.Zero.Error().Err(err).Msg("yfuzzer: cat failed")
	}

	if startOffset != 0 {
		if _, err := io.CopyN(io.Discard, contentReader, int64(startOffset)); err != nil {
			return err
		}
	}

	rw := ycl.GetRW()
	chunk := make([]byte, maxCatChunk)
	var total int64
	for {
		n, rerr := contentReader.Read(chunk)
		if n > 0 {
			/* Write the freshly read slice in small random-sized pieces,
			 * exercising partial-write handling on the client side. */
			pending := chunk[:n]
			for len(pending) > 0 {
				var writeChunk []byte
				if len(pending) < 2 {
					writeChunk = pending
					pending = nil
				} else {
					currLen := rand.Intn(len(pending)) + 1
					writeChunk = pending[:currLen]
					pending = pending[currLen:]
				}

				if _, werr := rw.Write(writeChunk); werr != nil {
					if errors.Is(werr, syscall.EPIPE) || errors.Is(werr, io.ErrClosedPipe) {
						ylogger.Zero.Warn().Err(werr).Uint("client id", ycl.ID()).Int64("copied bytes", total).Msg("yfuzzer: client disconnected during cat")
					} else {
						ylogger.Zero.Error().Err(werr).Uint("client id", ycl.ID()).Int64("copied bytes", total).Msg("yfuzzer: failed to cat object")
					}
					_ = ycl.ReplyError(werr, "yfuzzer: failed to write cat chunk")
					return werr
				}
			}
			total += int64(n)
		}
		if rerr != nil {
			if !errors.Is(rerr, io.EOF) {
				if errors.Is(rerr, syscall.EPIPE) || errors.Is(rerr, io.ErrClosedPipe) {
					ylogger.Zero.Warn().Err(rerr).Uint("client id", ycl.ID()).Int64("copied bytes", total).Msg("yfuzzer: client disconnected during cat")
				} else {
					ylogger.Zero.Error().Err(rerr).Uint("client id", ycl.ID()).Int64("copied bytes", total).Msg("yfuzzer: failed to read object")
				}
				return rerr
			}
			break
		}
	}

	ylogger.Zero.Debug().Int64("copied bytes", total).Msg("yfuzzer: cat object")

	if _, err := rw.Write(message.NewReadyForQueryMessage().Encode()); err != nil {
		_ = ycl.ReplyError(err, "yfuzzer: failed to write ready for query")
		return err
	}

	return nil
}

var cfgPath string

var logLevel string

<<<<<<< HEAD
=======
var (
	maxIterations = 100
	maxBatchSize  = 500
	maxPathLen    = 256
	maxCatChunk   = 8192
)

>>>>>>> f60a47a (fuzz reads)
var rootCmd = &cobra.Command{
	Use: "yfuzzer",
	RunE: func(cmd *cobra.Command, args []string) error {
		ylogger.Zero.Debug().Str("config-path", cfgPath).Msg("using config path")
		err := config.LoadInstanceConfig(cfgPath)
		if err != nil {
			return err
		}

		instanceCnf := config.InstanceConfig()

		instance := core.NewInstance()
		/* Proxies every call to standard implementation,
		* expect for fuzzed (and stub) List output */
		instance.ProtoMgr = &FuzzerProtoMgr{}

		if logLevel == "" {
			logLevel = instanceCnf.LogLevel
		}

		if instanceCnf.LogPath != "" {
			ylogger.ReloadLogger(instanceCnf.LogPath, logLevel)
		}

		return instance.Run(instanceCnf)
	},
	Version: pkg.YproxyVersionRevision,
}

func init() {
	/* For regular storage config, use this until we (if ever) proxy r-w calls.  */
	rootCmd.PersistentFlags().StringVarP(&cfgPath, "config", "c", "/etc/yproxy/yproxy.yaml", "path to yproxy config file")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "", "log level")
	rootCmd.PersistentFlags().IntVar(&maxCatChunk, "max-cat-chunk", maxCatChunk, "upper bound (exclusive) for cat read buffer size in bytes")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		ylogger.Zero.Fatal().Err(err).Msg("failed to execute root command")
	}
}
