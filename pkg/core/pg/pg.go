package pg

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/yezzey-gp/yproxy/config"
	"github.com/yezzey-gp/yproxy/pkg/client"
	"github.com/yezzey-gp/yproxy/pkg/clientpool"
	"github.com/yezzey-gp/yproxy/pkg/core/parser"
	"github.com/yezzey-gp/yproxy/pkg/ylogger"
)

func PostgresIface(cl net.Conn, p clientpool.Pool, instanceStart time.Time) {
	defer cl.Close()

	conn := pgproto3.NewBackend(cl, cl)

init:
	for {
		msg, err := conn.ReceiveStartupMessage()
		if err != nil {
			ylogger.Zero.Error().Err(err)
			return
		}

		switch q := msg.(type) {
		case *pgproto3.SSLRequest:
			/* negotiate */
			ylogger.Zero.Info().Msg("negotiate ssl proto version")
			if _, err := cl.Write([]byte{'N'}); err != nil {
				ylogger.Zero.Error().Err(err).Msg("proto mess up")
			}
		case *pgproto3.StartupMessage:
			ylogger.Zero.Info().Uint32("proto", q.ProtocolVersion).Msg("accept psql proto version")
			break init
		default:
			ylogger.Zero.Error().Msg("proto mess up")
			return
		}
	}

	/* send aut ok */
	conn.Send(&pgproto3.AuthenticationCleartextPassword{})
	conn.Flush()
	_, err := conn.Receive()
	if err != nil {
		ylogger.Zero.Error().Err(err).Msg("failed to complete AUTH")
	}
	conn.Send(&pgproto3.AuthenticationOk{})
	conn.Flush()
	conn.Send(&pgproto3.ReadyForQuery{
		TxStatus: 'I',
	})
	conn.Flush()

	/* main cycle */

	for {
		msg, err := conn.Receive()

		if err != nil {
			ylogger.Zero.Error().Err(err).Msg("failed to recieve message")
			return
		}

		switch q := msg.(type) {
		case *pgproto3.Query:
			ylogger.Zero.Info().Str("query", q.String).Msg("serving request")

			node, err := parser.Parse(q.String)

			if err != nil {
				conn.Send(&pgproto3.ErrorResponse{
					Message: "failed to parse query",
				})
				conn.Send(&pgproto3.ReadyForQuery{})
				conn.Flush()
				continue
			}

			ylogger.Zero.Info().Interface("node", node).Msg("parsed nodetree")

			switch q := node.(type) {
			case *parser.SayHelloCommand:
				conn.Send(&pgproto3.RowDescription{
					Fields: []pgproto3.FieldDescription{
						{
							Name:        []byte("row"),
							DataTypeOID: 25, /* textoid*/
						},
					},
				})

				conn.Send(&pgproto3.DataRow{
					Values: [][]byte{[]byte("hi")},
				})
				conn.Send(&pgproto3.CommandComplete{CommandTag: []byte("YPROXYHELLO")})

				conn.Send(&pgproto3.ReadyForQuery{
					TxStatus: 'I',
				})
				conn.Flush()
			case *parser.ShowCommand:
				_ = ProcessShow(conn, q.Type, p, instanceStart)
			case *parser.KKBCommand:
				ylogger.Zero.Error().Msg("recieved die command, exiting")

				conn.Send(&pgproto3.RowDescription{
					Fields: []pgproto3.FieldDescription{
						{
							Name:        []byte("row"),
							DataTypeOID: 25, /* textoid*/
						},
					},
				})

				conn.Send(&pgproto3.DataRow{
					Values: [][]byte{[]byte("exit")},
				})
				conn.Send(&pgproto3.CommandComplete{CommandTag: []byte("EXIT")})

				conn.Send(&pgproto3.ReadyForQuery{
					TxStatus: 'I',
				})
				conn.Flush()
				/* TDB: remove this kostyl */
				time.Sleep(time.Second * 5)
				os.Exit(2)
			default:
				conn.Send(&pgproto3.ErrorResponse{
					Message: "unknown command",
				})

				conn.Send(&pgproto3.ReadyForQuery{
					TxStatus: 'I',
				})
				conn.Flush()
			}

		default:
			ylogger.Zero.Error().Interface("msg", q).Msg("unssuported message type")
		}
	}
}

func quantToString(ct int) string {
	switch ct {
	case 0:
		return "< 1 MB"
	case 1:
		return "< 16 MB"
	case 2:
		return ">= 16 MB"
	default:
		return ""
	}
}

func reportQuants(conn *pgproto3.Backend, qs []clientpool.QuantInfo, ct int) {
	for _, info := range qs {
		values := [][]byte{
			[]byte(info.Op),
			[]byte(quantToString(ct)),
		}

		for _, qs := range info.Q {
			values = append(values, []byte(fmt.Sprintf("%v", qs)))
		}

		conn.Send(&pgproto3.DataRow{
			Values: values})
	}
}

func ProcessShow(conn *pgproto3.Backend, s string, p clientpool.Pool, instanceStart time.Time) error {
	switch s {
	case "clients":
		/*
		* OPType, client_id, byte offset, opstart, xpath.
		 */

		var infos []client.YproxyClient
		if err := p.ClientPoolForeach(func(c client.YproxyClient) error {
			infos = append(infos, c)
			return nil
		}); err != nil {
			return err
		}

		conn.Send(&pgproto3.RowDescription{
			Fields: []pgproto3.FieldDescription{
				{
					Name:        []byte("opType"),
					DataTypeOID: 25, /* textoid */
				},
				{
					Name:        []byte("client_id"),
					DataTypeOID: 25, /* textoid */
				},
				{
					Name:        []byte("byte offset"),
					DataTypeOID: 25, /* textoid */
				},
				{
					Name:        []byte("opstart"),
					DataTypeOID: 25, /* textoid */
				},
				{
					Name:        []byte("xpath"),
					DataTypeOID: 25, /* textoid */
				},
			},
		})

		for _, info := range infos {

			conn.Send(&pgproto3.DataRow{
				Values: [][]byte{
					[]byte(info.OPType().String()),
					[]byte(fmt.Sprintf("%d", info.ID())),
					[]byte(fmt.Sprintf("%d", info.ByteOffset())),
					[]byte(fmt.Sprintf("%v", info.OPStart())),
					[]byte(info.ExternalFilePath()),
				},
			})
		}

		conn.Send(&pgproto3.CommandComplete{CommandTag: []byte("CLIENTS")})

		conn.Send(&pgproto3.ReadyForQuery{
			TxStatus: 'I',
		})

		return conn.Flush()

	case "stats":
		/*
		* OPType, client_id, xpath, quantilies.
		 */

		quants := []float64{
			.001,
			.01,
			.1,
			.25,
			.5,
			.75,
			.9,
			.99,
			.999,
		}

		fields := []pgproto3.FieldDescription{
			{
				Name:        []byte("opType"),
				DataTypeOID: 25, /* textoid */
			},
			{
				Name:        []byte("size category"),
				DataTypeOID: 25, /* textoid */
			},
		}

		for _, q := range quants {
			fields = append(fields, pgproto3.FieldDescription{
				Name:        []byte(fmt.Sprintf("quantiles_%v", q)),
				DataTypeOID: 25, /* textoid */
			})
		}

		conn.Send(&pgproto3.RowDescription{
			Fields: fields,
		})

		for ct := 0; ct < 3; ct++ {
			reportQuants(conn, p.Quantile(ct, quants), ct)
		}

		conn.Send(&pgproto3.CommandComplete{CommandTag: []byte("STATS")})

		conn.Send(&pgproto3.ReadyForQuery{
			TxStatus: 'I',
		})

		return conn.Flush()
	case "stat_system":
		conn.Send(&pgproto3.RowDescription{
			Fields: []pgproto3.FieldDescription{
				{
					Name:        []byte("start time"),
					DataTypeOID: 25, /* textoid */
				},
				{
					Name:        []byte("storage concurrency"),
					DataTypeOID: 25, /* textoid */
				},
			},
		})

		conn.Send(&pgproto3.DataRow{
			Values: [][]byte{
				[]byte(fmt.Sprintf("%v", instanceStart)),
				[]byte(fmt.Sprintf("%v", config.InstanceConfig().StorageCnf.StorageConcurrency)),
			},
		})

		conn.Send(&pgproto3.CommandComplete{CommandTag: []byte("STATS")})

		conn.Send(&pgproto3.ReadyForQuery{
			TxStatus: 'I',
		})

		return conn.Flush()
	default:

		conn.Send(&pgproto3.ErrorResponse{
			Message: "unrecognized SHOW type",
		})
		conn.Send(&pgproto3.ReadyForQuery{
			TxStatus: 'I',
		})

		return conn.Flush()
	}
}
