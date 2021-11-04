package mysql

import (
	"database/sql"
	mynet "github.com/RomanAvdeenko/utils/net"
	"ispcp/internal/model"
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
		_, err := s.db.Exec(INSERT_QUERY, mynet.Ip2int(v.IpAddr), 1, v.Time, v.MACAddr.String())
		if err != nil {
			return err
		}
	}
	return nil
}

//`alive` (
//`alive_ip` int(10) unsigned NOT NULL,
//`alive_is` tinyint(1) NOT NULL,
//`alive_ts` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
//`alive_mac` char(20) NOT NULL DEFAULT '0'
