package store

import "ispcp/internal/model"

type Store interface {
	Store(pongs *model.Pongs) error
}
