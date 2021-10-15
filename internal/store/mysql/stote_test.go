package mysql_test

import (
	"alive/internal/model"
	"alive/internal/store/mysql"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var databaseURL string

func TestMain(m *testing.M) {
	if databaseURL = os.Getenv("DATABASE_URL"); databaseURL == "" {
		databaseURL = "ispcp:911ispcp911@tcp(localhost)/billing_test"
	}
	os.Exit(m.Run())
}

func TestAdd(t *testing.T) {
	db, teardown := mysql.TestDB(t, databaseURL)
	defer teardown("")

	store := mysql.New(db)

	r := model.TestCheckHostRecord(t)
	assert.NoError(t, store.Add(r))

}
