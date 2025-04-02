package main

import (
	"github.com/spf13/cobra"

	"github.com/yezzey-gp/yproxy/config"
	"github.com/yezzey-gp/yproxy/pkg"
	"github.com/yezzey-gp/yproxy/pkg/core"
	"github.com/yezzey-gp/yproxy/pkg/testutils"
	"github.com/yezzey-gp/yproxy/pkg/ylogger"
)

var cfgPath string

var logLevel string

var rootCmd = &cobra.Command{
	Use:   "",
	Short: "",
	RunE: func(cmd *cobra.Command, args []string) error {
		ylogger.Zero.Debug().Str("config-path", cfgPath).Msg("using config path")
		err := config.LoadInstanceConfig(cfgPath)
		if err != nil {
			return err
		}

		instanceCnf := config.InstanceConfig()

		instance := core.NewInstance()

		if instanceCnf.LogPath != "" {
			ylogger.ReloadLogger(instanceCnf.LogPath)
		}
		if logLevel == "" {
			logLevel = instanceCnf.LogLevel
		}
		ylogger.UpdateZeroLogLevel(logLevel)

		return instance.Run(instanceCnf)
	},
	Version: pkg.YproxyVersionRevision,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgPath, "config", "c", "/etc/yproxy/yproxy.yaml", "path to yproxy config file")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "", "log level")
	rootCmd.PersistentFlags().BoolVarP(&testutils.TestMode, "test-mode", "", false, "enable test mode. Use only in regression tests")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		ylogger.Zero.Fatal().Err(err).Msg("")
	}
}
