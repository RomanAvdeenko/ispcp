package mysql

import (
	"alive/internal/model"
	"database/sql"
)

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Add(v *model.CheckHostRecord) error {
	const INSERT_QUERY = "REPLACE INTO alive(alive_ip, alive_is, alive_ts, alive_mac) VALUES(?, ?, ?, ?)"

	_, err := s.db.Exec(INSERT_QUERY, v.IP, v.Alive, v.Time, v.MAC)
	return err
}
