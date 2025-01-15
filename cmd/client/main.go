package main

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/spf13/cobra"
	"github.com/yezzey-gp/yproxy/config"
	"github.com/yezzey-gp/yproxy/pkg/client"
	"github.com/yezzey-gp/yproxy/pkg/message"
	"github.com/yezzey-gp/yproxy/pkg/object"
	"github.com/yezzey-gp/yproxy/pkg/proc"
	"github.com/yezzey-gp/yproxy/pkg/settings"
	"github.com/yezzey-gp/yproxy/pkg/tablespace"
	"github.com/yezzey-gp/yproxy/pkg/ylogger"
)

var (
	cfgPath    string
	oldCfgPath string
	logLevel   string

	decrypt bool
	/* Put command flags */
	encrypt            bool
	storageClass       string
	tableSpace         string
	multipartChunksize int64
	multipartUpload    bool

	offset uint64

	segmentPort uint64
	segmentNum  uint64
	confirm     bool
	garbage     bool
)

// TODOV
func Runner(f func(net.Conn, *config.Instance, []string) error) func(*cobra.Command, []string) error {

	return func(cmd *cobra.Command, args []string) error {

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
		return f(con, instanceCnf, args)
	}
}

func catFunc(con net.Conn, instanceCnf *config.Instance, args []string) error {
	msg := message.NewCatMessage(args[0], decrypt, offset).Encode()
	_, err := con.Write(msg)
	if err != nil {
		return err
	}

	ylogger.Zero.Debug().Bytes("msg", msg).Msg("constructed cat message")

	_, err = io.Copy(os.Stdout, con)
	if err != nil {
		return err
	}

	return nil
}

func copyFunc(con net.Conn, instanceCnf *config.Instance, args []string) error {
	ylogger.Zero.Info().Msg("Execute copy command")
	ylogger.Zero.Info().Str("name", args[0]).Msg("copy")
	msg := message.NewCopyMessage(args[0], oldCfgPath, encrypt, decrypt, confirm, segmentPort).Encode()
	_, err := con.Write(msg)
	if err != nil {
		return err
	}

	ylogger.Zero.Debug().Bytes("msg", msg).Msg("constructed copy msg")

	client := client.NewYClient(con)
	protoReader := proc.NewProtoReader(client)

	ansType, body, err := protoReader.ReadPacket()
	if err != nil {
		ylogger.Zero.Debug().Err(err).Msg("error while answer")
		return err
	}

	if ansType != message.MessageTypeReadyForQuery {
		return fmt.Errorf("failed to copy, msg: %v", body)
	}
	return nil
}

func putFunc(con net.Conn, instanceCnf *config.Instance, args []string) error {
	ycl := client.NewYClient(con)
	r := proc.NewProtoReader(ycl)

	msg := message.NewPutMessageV2(args[0], encrypt, []settings.StorageSettings{
		{
			Name:  message.StorageClassSetting,
			Value: storageClass,
		},
		{
			Name:  message.TableSpaceSetting,
			Value: tableSpace,
		},
		{
			Name:  message.MultipartChunksize,
			Value: fmt.Sprintf("%d", multipartChunksize),
		},
		{
			Name:  message.MultipartUpload,
			Value: fmt.Sprintf("%t", multipartUpload),
		},
	}).Encode()
	_, err := con.Write(msg)
	if err != nil {
		return err
	}

	ylogger.Zero.Debug().Bytes("msg", msg).Msg("constructed put message")

	const SZ = 65536
	chunk := make([]byte, SZ)
	for {
		n, err := os.Stdin.Read(chunk)
		if n > 0 {
			msg := message.NewCopyDataMessage()
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

	msg = message.NewCommandCompleteMessage().Encode()
	_, err = con.Write(msg)
	if err != nil {
		return err
	}

	tp, _, err := r.ReadPacket()
	if err != nil {
		return err
	}

	if tp == message.MessageTypeReadyForQuery {
		ylogger.Zero.Debug().Msg("got rfq")
		return nil
	} else {
		return fmt.Errorf("failed to get rfq")
	}
}

func listFunc(con net.Conn, instanceCnf *config.Instance, args []string) error {
	msg := message.NewListMessage(args[0]).Encode()
	_, err := con.Write(msg)
	if err != nil {
		return err
	}

	ylogger.Zero.Debug().Bytes("msg", msg).Msg("constructed list message")

	ycl := client.NewYClient(con)
	r := proc.NewProtoReader(ycl)

	done := false
	res := make([]*object.ObjectInfo, 0)
	for {
		if done {
			break
		}
		tp, body, err := r.ReadPacket()
		if err != nil {
			return err
		}

		switch tp {
		case message.MessageTypeObjectMeta:
			meta := message.ObjectInfoMessage{}
			meta.Decode(body)

			res = append(res, meta.Content...)
		case message.MessageTypeReadyForQuery:
			done = true
		default:
			return fmt.Errorf("incorrect message type: %s", tp.String())
		}
	}

	for _, meta := range res {
		fmt.Printf("Object: {Name: \"%s\", size: %d}\n", meta.Path, meta.Size)
	}
	return nil
}

func deleteFunc(con net.Conn, instanceCnf *config.Instance, args []string) error {
	ylogger.Zero.Info().Msg("Execute delete command")

	ylogger.Zero.Info().Str("name", args[0]).Msg("delete")
	msg := message.NewDeleteMessage(args[0], segmentPort, segmentNum, confirm, garbage).Encode()
	_, err := con.Write(msg)
	if err != nil {
		return err
	}

	ylogger.Zero.Debug().Bytes("msg", msg).Msg("constructed delete msg")

	client := client.NewYClient(con)
	protoReader := proc.NewProtoReader(client)

	ansType, body, err := protoReader.ReadPacket()
	if err != nil {
		ylogger.Zero.Debug().Err(err).Msg("error while recieving answer")
		return err
	}

	if ansType != message.MessageTypeReadyForQuery {
		return fmt.Errorf("failed to delete, msg: %v", body)
	}

	return nil
}

func goolFunc(con net.Conn, instanceCnf *config.Instance, args []string) error {
	msg := message.NewGoolMessage(args[0]).Encode()
	_, err := con.Write(msg)
	if err != nil {
		ylogger.Zero.Debug().Err(err)
		return err
	}

	ylogger.Zero.Debug().Bytes("msg", msg).Msg("constructed gool message")

	ycl := client.NewYClient(con)
	r := proc.NewProtoReader(ycl)

	done := false
	for {
		if done {
			break
		}
		tp, _, err := r.ReadPacket()
		if err != nil {
			return err
		}

		switch tp {
		case message.MessageTypeReadyForQuery:
			ylogger.Zero.Debug().Bytes("msg", msg).Msg("got RFQ message")
			done = true
		default:
			return fmt.Errorf("incorrect gool")
		}
	}

	return nil
}

var rootCmd = &cobra.Command{
	Use:   "",
	Short: "",
}

var catCmd = &cobra.Command{
	Use:   "cat",
	Short: "cat",
	Args:  cobra.ExactArgs(1),
	RunE:  Runner(catFunc),
}

var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "copy",
	Args:  cobra.ExactArgs(1),
	RunE:  Runner(copyFunc),
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete",
	RunE:  Runner(deleteFunc),
}

var putCmd = &cobra.Command{
	Use:   "put",
	Short: "put",
	Args:  cobra.ExactArgs(1),
	RunE:  Runner(putFunc),
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list",
	Args:  cobra.ExactArgs(1),
	RunE:  Runner(listFunc),
}

var goolCmd = &cobra.Command{
	Use:   "gool",
	Short: "gool",
	RunE:  Runner(goolFunc),
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgPath, "config", "c", "/etc/yproxy/yproxy.yaml", "path to yproxy config file")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "", "log level")

	catCmd.PersistentFlags().BoolVarP(&decrypt, "decrypt", "d", false, "decrypt external object or not")
	catCmd.PersistentFlags().Uint64VarP(&offset, "offset", "o", 0, "start offset for read")
	rootCmd.AddCommand(catCmd)

	copyCmd.PersistentFlags().BoolVarP(&decrypt, "decrypt", "d", false, "decrypt external object or not")
	copyCmd.PersistentFlags().BoolVarP(&encrypt, "encrypt", "e", false, "encrypt external object before put")
	copyCmd.PersistentFlags().StringVarP(&oldCfgPath, "old-config", "", "/etc/yproxy/yproxy.yaml", "path to old yproxy config file")
	copyCmd.PersistentFlags().Uint64VarP(&segmentPort, "port", "p", 6000, "port that segment is listening on")
	copyCmd.PersistentFlags().BoolVarP(&confirm, "confirm", "", false, "confirm copy")
	rootCmd.AddCommand(copyCmd)

	putCmd.PersistentFlags().BoolVarP(&encrypt, "encrypt", "e", false, "encrypt external object before put")
	putCmd.PersistentFlags().StringVarP(&storageClass, "storage-class", "s", "STANDARD", "storage class for message upload")
	putCmd.PersistentFlags().StringVarP(&tableSpace, "tablespace", "t", tablespace.DefaultTableSpace, "storage class for message upload")

	putCmd.PersistentFlags().Int64VarP(&multipartChunksize, "multipart-chunk-size", "", int64(64*1024*1024), "S3 chunk size for multipart upload")
	putCmd.PersistentFlags().BoolVarP(&multipartUpload, "multipart-upload", "", true, "S3 multipart or single part upload")
	rootCmd.AddCommand(putCmd)

	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(goolCmd)

	deleteCmd.PersistentFlags().Uint64VarP(&segmentPort, "port", "p", 6000, "port that segment is listening on")
	deleteCmd.PersistentFlags().Uint64VarP(&segmentNum, "segnum", "s", 0, "logical number of a segment")
	deleteCmd.PersistentFlags().BoolVarP(&confirm, "confirm", "", false, "confirm deletion")
	deleteCmd.PersistentFlags().BoolVarP(&garbage, "garbage", "g", false, "delete garbage")
	rootCmd.AddCommand(deleteCmd)

}

func main() {
	if err := rootCmd.Execute(); err != nil {
		ylogger.Zero.Fatal().Err(err).Msg("")
	}
}
