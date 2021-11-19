package mysql

import (
	"database/sql"
	"fmt"
	"ispcp/internal/model"
	"os"
	"time"

	mynet "github.com/RomanAvdeenko/utils/net"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Store(pongs *model.Pongs) error {
	const INSERT_QUERY = "REPLACE INTO alive(alive_ip, alive_is, alive_ts, alive_mac) VALUES(?, ?, ?, ?)"

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.Stamp})
	ps := pongs.LoadAll()
	for _, v := range *ps {
		_, err := s.db.Exec(INSERT_QUERY, mynet.Ip2int(v.IpAddr), v.Alive, v.Time, v.MACAddr.String())
		if err != nil {
			return err
		}
	}
	log.Info().Msg(fmt.Sprintf("Written to store %v records", len(*ps)))
	return nil
}
