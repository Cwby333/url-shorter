package dsn

import (
	"fmt"

	"github.com/Cwby333/url-shorter/internal/config"
)

func NewDsn(dbms string, cfgDb config.Database) string {
	var dsn string

	switch dbms {
	case "postgresPool":
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&pool_max_conns=%d&pool_min_conns=%d", cfgDb.User, cfgDb.Password, cfgDb.Host, cfgDb.Port, cfgDb.DBname, cfgDb.MaxConn, cfgDb.MinConn)
	}

	return dsn
}
