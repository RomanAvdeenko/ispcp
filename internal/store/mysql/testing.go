package mysql

import (
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestDB(t *testing.T, databaseURL string) (*sql.DB, func(...string)) {
	t.Helper()

	db, err := sql.Open("mysql", databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		t.Fatal(err)
	}
	//
	foo := func(tables ...string) {
		for _, table := range tables {
			if len(table) > 0 {
				if _, err := db.Exec("TRUNCATE TABLE " + table); err != nil {
					t.Fatal(err)
				}
			}
		}

	}
	return db, foo
}
