package file

import (
	"fmt"
	"ispcp/internal/model"
	"os"
)

type Store struct {
	file *os.File
}

func New(file *os.File) *Store {
	return &Store{file: file}
}

func (s *Store) Store(pongs *model.Pongs) error {
	ps := pongs.LoadAll()
	for _, v := range *ps {
		_, err := fmt.Fprintf(s.file, "%v\n", v)
		if err != nil {
			return err
		}
	}
	return nil
}
