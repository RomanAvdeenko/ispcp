package mysql

import (
	"database/sql"
	"ispcp/internal/model"
	"log"

	mynet "github.com/RomanAvdeenko/utils/net"
)

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Store(pongs *model.Pongs) error {
	const INSERT_QUERY = "REPLACE INTO alive(alive_ip, alive_is, alive_ts, alive_mac) VALUES(?, ?, ?, ?)"

	ps := pongs.LoadAll()
	for _, v := range *ps {
		log.Default().Println("Write to db: " + v.Human())
		_, err := s.db.Exec(INSERT_QUERY, mynet.Ip2int(v.IpAddr), v.Alive, v.Time, v.MACAddr.String())
		if err != nil {
			return err
		}
	}
	return nil
}
