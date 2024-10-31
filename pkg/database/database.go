package database

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/pkg/errors"
	"github.com/yezzey-gp/yproxy/pkg/ylogger"
)

//go:generate mockgen -destination=../mock/mock_database_interractor.go -package mock
type DatabaseInterractor interface {
	GetVirtualExpireIndexes(int) (map[string]bool, map[string]uint64, error)
}

type DatabaseHandler struct {
}

type DB struct {
	name       string
	tablespace pgtype.OID
	oid        pgtype.OID
}

type Ei struct {
	reloid     pgtype.OID
	relfileoid pgtype.OID
	expireLsn  string
	fqnmd5     string
}

/*
* Retrieve virtual index info and expire hint index info.
 */
func (database *DatabaseHandler) GetVirtualExpireIndexes(port int) (map[string]bool, map[string]uint64, error) { //TODO несколько баз
	db, err := getDatabase(port)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get ao/aocs tables %v", err) //fix
	}
	ylogger.Zero.Debug().Str("database name", db.name).Msg("recieved database")
	conn, err := connectToDatabase(port, db.name)
	if err != nil {
		return nil, nil, err
	}
	defer conn.Close() //error
	ylogger.Zero.Debug().Msg("connected to database")

	/*
	* if yezzey version >= 1.8.1, then
	* SELECT x_path, expire_lsn FROM yezzey.yezzey_expire_hint;
	 */

	rows, err := conn.Query(`SELECT x_path FROM yezzey.yezzey_virtual_index;`)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get ao/aocs tables %v", err) //fix
	}
	defer rows.Close()
	ylogger.Zero.Debug().Msg("yezzey virtual index fetched")

	c2 := make(map[string]bool, 0)
	for rows.Next() {
		xpath := ""
		if err := rows.Scan(&xpath); err != nil {
			return nil, nil, fmt.Errorf("unable to parse query output %v", err)
		}
		p1 := strings.Split(xpath, "/")
		p2 := p1[len(p1)-1]
		p3 := strings.Split(p2, "_")
		if len(p3) >= 4 {
			p2 = fmt.Sprintf("%s_%s_%s_%s_", p3[0], p3[1], p3[2], p3[3])
		}
		c2[p2] = true
		ylogger.Zero.Debug().Str("file", p2).Msg("added")
	}
	ylogger.Zero.Debug().Msg("read 3")

	return c2, map[string]uint64{}, err
}

func getDatabase(port int) (DB, error) {
	conn, err := connectToDatabase(port, "postgres")
	if err != nil {
		return DB{}, err
	}
	defer conn.Close() //error
	ylogger.Zero.Debug().Msg("connected to db")
	rows, err := conn.Query(`SELECT dattablespace, oid, datname FROM pg_database WHERE datallowconn;`)
	if err != nil {
		return DB{}, err
	}
	defer rows.Close()
	ylogger.Zero.Debug().Msg("recieved db list")

	for rows.Next() {
		row := DB{}
		ylogger.Zero.Debug().Msg("cycle 1")
		if err := rows.Scan(&row.tablespace, &row.oid, &row.name); err != nil {
			return DB{}, err
		}
		ylogger.Zero.Debug().Msg("cycle 2")
		ylogger.Zero.Debug().Str("db", row.name).Int("db", int(row.oid)).Int("db", int(row.tablespace)).Msg("database")
		if row.name == "postgres" {
			continue
		}

		ylogger.Zero.Debug().Str("db", row.name).Msg("check database")
		connDb, err := connectToDatabase(port, row.name)
		if err != nil {
			return DB{}, err
		}
		defer connDb.Close() //error
		ylogger.Zero.Debug().Msg("cycle 3")

		rowsdb, err := connDb.Query(`SELECT exists(SELECT * FROM information_schema.schemata WHERE schema_name='yezzey');`)
		if err != nil {
			return DB{}, err
		}
		defer rowsdb.Close()
		ylogger.Zero.Debug().Msg("cycle 4")
		var ans bool
		rowsdb.Next()
		err = rowsdb.Scan(&ans)
		if err != nil {
			ylogger.Zero.Error().AnErr("error", err).Msg("error during yezzey check")
			return DB{}, err
		}
		ylogger.Zero.Debug().Bool("result", ans).Msg("find yezzey schema")
		if ans {
			ylogger.Zero.Debug().Str("db", row.name).Msg("found yezzey schema in database")
			ylogger.Zero.Debug().Int("db", int(row.oid)).Int("db", int(row.tablespace)).Msg("found yezzey schema in database")
			return row, nil
		}

		ylogger.Zero.Debug().Str("db", row.name).Msg("no yezzey schema in database")
	}
	return DB{}, fmt.Errorf("no yezzey schema across databases")
}

func connectToDatabase(port int, database string) (*pgx.Conn, error) {
	config, err := pgx.ParseEnvLibpq()
	if err != nil {
		return nil, errors.Wrap(err, "Connect: unable to read environment variables")
	}

	config.Port = uint16(port)
	config.Database = database

	config.RuntimeParams["gp_role"] = "utility"
	conn, err := pgx.Connect(config)
	if err != nil {
		config.RuntimeParams["gp_session_role"] = "utility"
		conn, err = pgx.Connect(config)
		if err != nil {
			fmt.Printf("error in connection %v", err) // delete this
			return nil, err
		}
	}
	return conn, nil
}
