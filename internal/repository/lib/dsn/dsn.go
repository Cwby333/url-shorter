package dsn

import (
	"errors"
	"fmt"

	"github.com/Cwby333/url-shorter/internal/config"
)

func NewDSN(dbms string, cfgDB config.Database) (string, error) {
	const op = "internal/repository/lib/dsn/dsn.go/NewDsn"

	var DSN string

	switch dbms {
	case "postgresPool":
		DSN = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&pool_max_conns=%d&pool_min_conns=%d", cfgDB.User, cfgDB.Password, cfgDB.Host, cfgDB.Port, cfgDB.DBname, cfgDB.MaxConn, cfgDB.MinConn)
	default:
		return "", fmt.Errorf("%s: %w", op, errors.New("not allows dbms in input"))
	}

	return DSN, nil
}
