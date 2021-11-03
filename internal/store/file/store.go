package file

import (
	"fmt"
	"ispcp/internal/model"
	"os"
	"time"
)

type Store struct {
	file *os.File
}

func New(file *os.File) *Store {
	return &Store{file: file}
}

func (s *Store) Store(pongs *model.Pongs) error {
	ps := pongs.LoadAll()
	_, err := s.file.WriteString("->: " + time.Now().Format(time.Stamp) + "\n")
	if err != nil {
		return err
	}
	for _, v := range *ps {
		_, err := s.file.WriteString(v.Human() + "\n")
		if err != nil {
			return err
		}
	}
	_, err = s.file.WriteString(fmt.Sprintf("->Total:\t%v\n", len(*ps)))
	if err != nil {
		return err
	}

	return nil
}
