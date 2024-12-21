package core

import (
	"net"

	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/yezzey-gp/yproxy/pkg/core/parser"
	"github.com/yezzey-gp/yproxy/pkg/ylogger"
)

func PostgresIface(cl net.Conn) {
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
	conn.Send(&pgproto3.AuthenticationOk{})
	conn.Flush()
	conn.Send(&pgproto3.ReadyForQuery{})
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

			switch node.(type) {
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
				conn.Send(&pgproto3.RowDescription{
					Fields: []pgproto3.FieldDescription{
						{
							Name:        []byte("row"),
							DataTypeOID: 25, /* textoid*/
						},
					},
				})

				conn.Send(&pgproto3.DataRow{
					Values: [][]byte{[]byte("here will be stats")},
				})
				conn.Send(&pgproto3.CommandComplete{CommandTag: []byte("STATS")})

				conn.Send(&pgproto3.ReadyForQuery{
					TxStatus: 'I',
				})
				conn.Flush()
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
