package store

import "alive/internal/model"

type Store interface {
	Add(model.CheckHostRecord) error
}
