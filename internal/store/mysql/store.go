package mysql

import (
	"database/sql"
	"ispcp/internal/model"
)

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) store(pongs *model.Pongs) error {
	//const INSERT_QUERY = "REPLACE INTO alive(alive_ip, alive_is, alive_ts, alive_mac) VALUES(?, ?, ?, ?)"
	//
	//_, err := s.db.Exec(INSERT_QUERY, v.IP, v.Alive, v.Time, v.MAC)
	return nil
}
