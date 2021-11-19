package file

import (
	"fmt"
	"ispcp/internal/model"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Store struct {
	file *os.File
}

func New(file *os.File) *Store {
	return &Store{file: file}
}

func (s *Store) Store(pongs *model.Pongs) error {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.Stamp})
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
	log.Info().Msg(fmt.Sprintf("Written to store %v records", len(*ps)))
	return nil
}
