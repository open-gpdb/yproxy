package database

import (
	"fmt"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/pkg/errors"
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

func (database *DatabaseHandler) populateIndex() {

}
func (database *DatabaseHandler) GetVirtualExpireIndex(port uint64, db DB, virtualIndex *map[string]bool, expireIndex *map[string]uint64) error {
	ylogger.Zero.Debug().Str("database name", db.name).Msg("received database")
	conn, err := connectToDatabase(port, db.name)
	if err != nil {
		return err
	}
	defer conn.Close() //error
	ylogger.Zero.Debug().Msg("connected to database")

	/* Todo: check that yezzey version >= 1.8.1 */
	if true { //yezzey version >=1.8.4 or didnt works
		rows, err := conn.Query(`SELECT x_path, expire_lsn FROM yezzey.yezzey_expire_hint;`)
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

	rows2, err := conn.Query(`SELECT x_path FROM yezzey.yezzey_virtual_index;`)
	if err != nil {
		return fmt.Errorf("unable to get ao/aocs tables %v", err) //fix
	}
	defer rows2.Close()

	for rows2.Next() {
		xpath := ""
		if err := rows2.Scan(&xpath); err != nil {
			return fmt.Errorf("unable to parse query output %v", err)
		}
		(*virtualIndex)[xpath] = true
		ylogger.Zero.Debug().Str("x_path", xpath).Msg("added")
	}
	ylogger.Zero.Debug().Msg("fetched virtual index info")

	return err
}
func (database *DatabaseHandler) GetNextLSN(port uint64, dbname string) (uint64, error) {
	databases, err := getDatabase(port)
	if err != nil || databases == nil {
		return 0, fmt.Errorf("unable to get ao/aocs tables %v", err) // fix
	}
	for _, db := range databases {
		if dbname != db.name {
			continue
		}
		ylogger.Zero.Debug().Str("database name", db.name).Msg("recieved database")
		conn, err := connectToDatabase(port, db.name)
		if err != nil {
			return 0, err
		}
		defer conn.Close() //error
		ylogger.Zero.Debug().Msg("connected to database")

		rows, err := conn.Query(`select pg_current_xlog_location();`)
		if err != nil {
			return 0, fmt.Errorf("unable to next lsn %v", err) //fix
		}
		defer rows.Close()
		ylogger.Zero.Debug().Msg("executed select")

		for rows.Next() {
			row := LSN{}
			if err := rows.Scan(&row.lsn); err != nil {
				return 0, fmt.Errorf("unable to parse query output %v", err)
			}
			lsn, err := pgx.ParseLSN(row.lsn)
			if err != nil {
				return 0, fmt.Errorf("unable to parse query output %v", err)
			}

			ylogger.Zero.Debug().Uint64("lsn", lsn).Msg("getted lsn")
			return lsn, nil
		}
		// unexcepted
		return 0, fmt.Errorf("unexpected error while getting next lsn")
	}

	return 0, fmt.Errorf("didnt find db with name %s", dbname)
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

func (database *DatabaseHandler) AddToExpireIndex(port uint64, dbname string, filename string, lsn uint64) error {

	databases, err := getDatabase(port)
	if err != nil || databases == nil {
		return fmt.Errorf("unable to get ao/aocs tables %v", err) // fix
	}
	for _, db := range databases {
		if dbname != db.name {
			continue
		}
		ylogger.Zero.Debug().Str("database name", db.name).Msg("recieved database")
		conn, err := connectToDatabase(port, db.name)
		if err != nil {
			return err
		}
		defer conn.Close() //error
		ylogger.Zero.Debug().Msg("connected to database")

		rows, err := conn.Query(`INSERT INTO yezzey.yezzey_expire_hint (x_path, expire_lsn) VALUES (%s , %d);`, filename, lsn)
		if err != nil {
			return fmt.Errorf("unable to update yezzey_expire_hint %v", err) //fix
		}
		defer rows.Close()
		ylogger.Zero.Debug().Msg("executed insert")

		return nil
	}

	return fmt.Errorf("didnt find db with name %s", dbname)
}

func (database *DatabaseHandler) DeleteFromExpireIndex(port uint64, dbname string, filename string) error {

	databases, err := getDatabase(port)
	if err != nil || databases == nil {
		return fmt.Errorf("unable to get ao/aocs tables %v", err) // fix
	}
	for _, db := range databases {
		if dbname != db.name {
			continue
		}
		ylogger.Zero.Debug().Str("database name", db.name).Msg("recieved database")
		conn, err := connectToDatabase(port, db.name)
		if err != nil {
			return err
		}
		defer conn.Close() //error
		ylogger.Zero.Debug().Msg("connected to database")

		rows, err := conn.Query(`DELETE FROM yezzey.yezzey_expire_hint WHERE x_path == "%s";`, filename)
		if err != nil {
			return fmt.Errorf("unable to delete from yezzey_expire_hint %v", err) //fix
		}
		defer rows.Close()
		ylogger.Zero.Debug().Msg("executed delete")

		return nil
	}

	return fmt.Errorf("didnt find db with name %s", dbname)
}
func getDatabase(port uint64) ([]DB, error) {
	var databases = []DB{}
	conn, err := connectToDatabase(port, "postgres")
	if err != nil {
		return nil, err
	}
	defer conn.Close() //error
	ylogger.Zero.Debug().Msg("connected to db")
	rows, err := conn.Query(`SELECT dattablespace, oid, datname FROM pg_database WHERE datallowconn;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ylogger.Zero.Debug().Msg("received db list")

	for rows.Next() {
		row := DB{}
		ylogger.Zero.Debug().Msg("cycle 1")
		if err := rows.Scan(&row.tablespace, &row.oid, &row.name); err != nil {
			return nil, err
		}
		ylogger.Zero.Debug().Msg("cycle 2")
		ylogger.Zero.Debug().Str("db", row.name).Int("db", int(row.oid)).Int("db", int(row.tablespace)).Msg("database")
		if row.name == "postgres" {
			continue
		}

		ylogger.Zero.Debug().Str("db", row.name).Msg("check database")
		connDb, err := connectToDatabase(port, row.name)
		if err != nil {
			return nil, err
		}
		defer connDb.Close() //error
		ylogger.Zero.Debug().Msg("cycle 3")

		rowsdb, err := connDb.Query(`SELECT exists(SELECT * FROM information_schema.schemata WHERE schema_name='yezzey');`)
		if err != nil {
			return nil, err
		}
		defer rowsdb.Close()
		ylogger.Zero.Debug().Msg("cycle 4")
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
	if len(databases) == 0 {
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
