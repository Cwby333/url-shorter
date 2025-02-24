package dsn

import (
	"fmt"

	"github.com/Cwby333/url-shorter/internal/config"
)

func NewDsn(dbms string, cfgDB config.Database) string {
	var dsn string

	switch dbms {
	case "postgresPool":
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&pool_max_conns=%d&pool_min_conns=%d", cfgDB.User, cfgDB.Password, cfgDB.Host, cfgDB.Port, cfgDB.DBname, cfgDB.MaxConn, cfgDB.MinConn)
	}

	return dsn
}
