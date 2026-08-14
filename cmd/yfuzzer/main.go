package main

import (
	"math/rand"

	"github.com/spf13/cobra"

	"github.com/yezzey-gp/yproxy/config"
	"github.com/yezzey-gp/yproxy/pkg"
	"github.com/yezzey-gp/yproxy/pkg/client"
	"github.com/yezzey-gp/yproxy/pkg/core"
	"github.com/yezzey-gp/yproxy/pkg/crypt"
	"github.com/yezzey-gp/yproxy/pkg/message"
	"github.com/yezzey-gp/yproxy/pkg/object"
	"github.com/yezzey-gp/yproxy/pkg/proc"
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

var cfgPath string

var logLevel string

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
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		ylogger.Zero.Fatal().Err(err).Msg("failed to execute root command")
	}
}
