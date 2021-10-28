package main

import (
	"flag"
	"ispcp/internal/pinger"
	"log"
	"os"
	"path/filepath"

	"github.com/ilyakaznacheev/cleanenv"
)

var (
	configFileName *string
	config         *pinger.Config
)

func init() {
	configFileName = flag.String("c", "./configs/config.yaml", "path to config file")
	flag.Parse()
}

//Read and parse config file
func readConfig(config *pinger.Config) error {
	configFileName, err := filepath.Abs(*configFileName)
	if err != nil {
		return err
	}
	log.Printf("Loading config: %v", configFileName)

	configFile, err := os.Open(configFileName)
	if err != nil {
		return err
	}

	defer configFile.Close()

	err = cleanenv.ReadConfig(configFileName, config)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	config = pinger.NewConfig()

	if err := readConfig(config); err != nil {
		log.Fatalln(err)
	}

	log.Printf("%+v\n", *config)
}
