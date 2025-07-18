package database

import (
	"fmt"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/pkg/errors"
	"github.com/yezzey-gp/yproxy/config"
	"github.com/yezzey-gp/yproxy/pkg/testutils"
	"github.com/yezzey-gp/yproxy/pkg/ylogger"
)

//go:generate mockgen -destination=../mock/mock_database_interractor.go -package mock
type DatabaseInterractor interface {
	GetVirtualExpireIndexes(port uint64) (map[string]bool, map[string]uint64, error)
}

type DatabaseHandler struct {
}

type DB struct {
	name       string
	tablespace pgtype.OID
	oid        pgtype.OID
}

type ExpireHint struct {
	expireLsn string
	x_path    string
}
type LSN struct {
	lsn string
}

func checkVersion(c *pgx.Conn, exp string) (bool, error) {
	rows, err := c.Query(`SELECT extversion FROM pg_extension WHERE extname = 'yezzey';`)
	if err != nil {
		return false, fmt.Errorf("unable to get ao/aocs tables %v", err) //fix
	}

	defer rows.Close()
	ylogger.Zero.Debug().Msg("executed select")

	if rows.Next() {
		var ver string
		if err := rows.Scan(&ver); err != nil {
			return false, fmt.Errorf("unable to parse query output %v", err)
		}
		if rows.Next() {
			return false, fmt.Errorf("unable to get yezzey extension version: duplicate output")
		}

		/* we compare versions lexicographically */
		return ver >= exp, nil
	}

	return false, fmt.Errorf("unable to get yezzey extension version")
}

func (database *DatabaseHandler) GetVirtualExpireIndex(port uint64, db DB, virtualIndex *map[string]bool, expireIndex *map[string]uint64) error {
	ylogger.Zero.Debug().Str("database name", db.name).Msg("received database")
	conn, err := connectToDatabase(port, db.name)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }() //error
	ylogger.Zero.Debug().Str("database name", db.name).Msg("GetVirtualExpireIndex: connected to database")

	/* Todo: check that yezzey version >= 1.8.4 */
	if ch, err := checkVersion(conn, "1.8.4"); err != nil {
		ylogger.Zero.Debug().Err(err).Msg("GetVirtualExpireIndex: failed")
		return err
	} else if ch {
		rows, err := conn.Query(`SELECT x_path, lsn FROM yezzey.yezzey_expire_hint;`)
		if err != nil {
			return fmt.Errorf("unable to get ao/aocs tables %v", err) //fix
		}
		defer rows.Close()
		ylogger.Zero.Debug().Msg("executed select")

		for rows.Next() {
			row := ExpireHint{}
			if err := rows.Scan(&row.x_path, &row.expireLsn); err != nil {
				return fmt.Errorf("unable to parse query output %v", err)
			}

			lsn, err := pgx.ParseLSN(row.expireLsn)
			if err != nil {
				return fmt.Errorf("unable to parse query output %v", err)
			}

			ylogger.Zero.Debug().Str("x_path", row.x_path).Str("lsn", row.expireLsn).Msg("added file to expire hint")
			(*expireIndex)[row.x_path] = lsn
		}
		ylogger.Zero.Debug().Msg("fetched expire hint info")
	}

	viRows, err := conn.Query(`SELECT x_path FROM yezzey.yezzey_virtual_index;`)
	if err != nil {
		return fmt.Errorf("unable to get ao/aocs tables %v", err) //fix
	}
	defer viRows.Close()

	for viRows.Next() {
		xpath := ""
		if err := viRows.Scan(&xpath); err != nil {
			return fmt.Errorf("unable to parse query output %v", err)
		}
		(*virtualIndex)[xpath] = true
		ylogger.Zero.Debug().Str("x_path", xpath).Msg("added")
	}
	ylogger.Zero.Debug().Msg("fetched virtual index info")

	return err
}

func (database *DatabaseHandler) GetNextLSN(port uint64, dbname string) (uint64, error) {
	ylogger.Zero.Debug().Str("database name", dbname).Msg("received database")
	conn, err := connectToDatabase(port, dbname)
	if err != nil {
		return 0, err
	}
	defer func() { _ = conn.Close() }() //error
	ylogger.Zero.Debug().Str("database name", dbname).Msg("GetNextLSN: connected to database")

	dbRow := conn.QueryRow(`select pg_current_xlog_location();`)
	ylogger.Zero.Debug().Msg("executed select")

	row := LSN{}
	if err := dbRow.Scan(&row.lsn); err != nil {
		return 0, fmt.Errorf("unable to parse query output %v", err)
	}
	lsn, err := pgx.ParseLSN(row.lsn)
	if err != nil {
		return 0, fmt.Errorf("unable to parse query output %v", err)
	}

	ylogger.Zero.Debug().Uint64("lsn", lsn).Msg("received lsn")
	return lsn, nil
}

func (database *DatabaseHandler) GetVirtualExpireIndexes(port uint64) (map[string]bool, map[string]uint64, error) { //TODO несколько баз
	databases, err := getDatabase(port)
	if err != nil || databases == nil {
		return nil, nil, fmt.Errorf("unable to get ao/aocs tables %v", err) //fix
	}

	expireIndex := make(map[string]uint64, 0)
	virtualIndex := make(map[string]bool, 0)
	for _, db := range databases {
		err = database.GetVirtualExpireIndex(port, db, &virtualIndex, &expireIndex)
		if err != nil {
			return nil, nil, err
		}

	}
	return virtualIndex, expireIndex, nil
}

func (database *DatabaseHandler) GetConnectToDatabase(port uint64, dbname string) (*pgx.Conn, error) {
	ylogger.Zero.Debug().Str("database name", dbname).Msg("received database")
	conn, err := connectToDatabase(port, dbname)
	if err != nil {
		return nil, err
	}
	ylogger.Zero.Debug().Str("database name", dbname).Msg("connected to database")
	return conn, nil

}
func (database *DatabaseHandler) AddToExpireIndex(conn *pgx.Conn, port uint64, dbname string, filename string, lsn uint64) error {
	rows, err := conn.Query(`INSERT INTO yezzey.yezzey_expire_hint (lsn,x_path) VALUES ($1 , $2);`, pgx.FormatLSN(lsn), filename)
	if err != nil {
		return fmt.Errorf("unable to update yezzey_expire_hint %v", err) //fix
	}
	defer rows.Close()

	return nil
}

func (database *DatabaseHandler) DeleteFromExpireIndex(conn *pgx.Conn, port uint64, dbname string, filename string) error {
	rows, err := conn.Query(`DELETE FROM yezzey.yezzey_expire_hint WHERE x_path = $1;`, filename)
	if err != nil {
		return fmt.Errorf("unable to delete from yezzey_expire_hint %v", err) //fix
	}
	defer rows.Close()

	return nil
}
func getDatabase(port uint64) ([]DB, error) {
	var databases = []DB{}
	conn, err := connectToDatabase(port, "postgres")
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }() //error
	ylogger.Zero.Debug().Msg("connected to db")
	rows, err := conn.Query(`SELECT dattablespace, oid, datname FROM pg_database WHERE datallowconn;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ylogger.Zero.Debug().Msg("received db list")

	for rows.Next() {
		row := DB{}
		if err := rows.Scan(&row.tablespace, &row.oid, &row.name); err != nil {
			return nil, err
		}
		ylogger.Zero.Debug().Str("db", row.name).Int("db", int(row.oid)).Int("db", int(row.tablespace)).Msg("database")
		if row.name == "postgres" {
			continue
		}

		ylogger.Zero.Debug().Str("db", row.name).Msg("check database")
		connDb, err := connectToDatabase(port, row.name)
		if err != nil {
			return nil, err
		}
		defer func() { _ = connDb.Close() }() //error

		rowsdb, err := connDb.Query(`SELECT exists(SELECT * FROM information_schema.schemata WHERE schema_name='yezzey');`)
		if err != nil {
			return nil, err
		}
		defer rowsdb.Close()
		var ans bool
		rowsdb.Next()
		err = rowsdb.Scan(&ans)
		if err != nil {
			ylogger.Zero.Error().AnErr("error", err).Msg("error during yezzey check")
			return nil, err
		}
		ylogger.Zero.Debug().Bool("result", ans).Msg("find yezzey schema")
		if ans {
			ylogger.Zero.Debug().Str("db", row.name).Msg("found yezzey schema in database")
			ylogger.Zero.Debug().Int("db", int(row.oid)).Int("db", int(row.tablespace)).Msg("found yezzey schema in database")
			databases = append(databases, row)
		}

		ylogger.Zero.Debug().Str("db", row.name).Msg("no yezzey schema in database")
	}
	if len(databases) == 0 && config.InstanceConfig().YezzeyRestoreParanoid {
		return nil, fmt.Errorf("no yezzey schema across databases")

	} else {
		return databases, nil
	}
}

func connectToDatabase(port uint64, database string) (*pgx.Conn, error) {
	config, err := pgx.ParseEnvLibpq()
	if err != nil {
		return nil, errors.Wrap(err, "Connect: unable to read environment variables")
	}

	config.Port = uint16(port)
	config.Database = database

	if testutils.TestMode {
		// Do not set GP-specific params
		return pgx.Connect(config)
	}
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
