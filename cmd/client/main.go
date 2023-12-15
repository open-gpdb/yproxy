package main

import (
	"io"
	"net"
	"os"

	"github.com/spf13/cobra"
	"github.com/yezzey-gp/yproxy/config"
	"github.com/yezzey-gp/yproxy/pkg/proc"
	"github.com/yezzey-gp/yproxy/pkg/ylogger"
)

var cfgPath string
var logLevel string
var decrypt bool
var encrypt bool

var rootCmd = &cobra.Command{
	Use:   "",
	Short: "",
}

var catCmd = &cobra.Command{
	Use:   "cat",
	Short: "cat",
	RunE: func(cmd *cobra.Command, args []string) error {

		err := config.LoadInstanceConfig(cfgPath)
		if err != nil {
			return err
		}

		instanceCnf := config.InstanceConfig()

		con, err := net.Dial("unix", instanceCnf.SocketPath)

		if err != nil {
			return err
		}

		defer con.Close()
		msg := proc.NewCatMessage(args[0], decrypt).Encode()
		_, err = con.Write(msg)
		if err != nil {
			return err
		}

		ylogger.Zero.Debug().Bytes("msg", msg).Msg("constructed message")

		_, err = io.Copy(os.Stdout, con)
		if err != nil {
			return err
		}

		return nil
	},
}

var putCmd = &cobra.Command{
	Use:   "put",
	Short: "put",
	RunE: func(cmd *cobra.Command, args []string) error {

		err := config.LoadInstanceConfig(cfgPath)
		if err != nil {
			return err
		}

		instanceCnf := config.InstanceConfig()

		con, err := net.Dial("unix", instanceCnf.SocketPath)

		if err != nil {
			return err
		}

		defer con.Close()
		msg := proc.NewPutMessage(args[0], encrypt).Encode()
		_, err = con.Write(msg)
		if err != nil {
			return err
		}

		ylogger.Zero.Debug().Bytes("msg", msg).Msg("constructed message")

		const SZ = 65536
		chunk := make([]byte, SZ)
		for {
			n, err := os.Stdin.Read(chunk)
			if n > 0 {
				msg := proc.NewCopyDataMessage()
				msg.Sz = uint64(n)
				msg.Data = make([]byte, msg.Sz)
				copy(msg.Data, chunk[:n])

				nwr, err := con.Write(msg.Encode())
				if err != nil {
					return err
				}

				ylogger.Zero.Debug().Int("len", nwr).Msg("written copy data msg")
			}

			if err == nil {
				continue
			}
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		ylogger.Zero.Debug().Msg("send command complete msg")

		msg = proc.NewCommandCompleteMessage().Encode()
		_, err = con.Write(msg)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgPath, "config", "c", "/etc/yproxy/yproxy.yaml", "path to yproxy config file")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "", "log level")

	catCmd.PersistentFlags().BoolVarP(&decrypt, "decrypt", "d", false, "decrypt external object or not")
	rootCmd.AddCommand(catCmd)

	putCmd.PersistentFlags().BoolVarP(&encrypt, "encrypt", "e", false, "encrypt external object before put")
	rootCmd.AddCommand(putCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		ylogger.Zero.Fatal().Err(err).Msg("")
	}
}
