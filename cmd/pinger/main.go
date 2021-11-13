package main

import (
	"flag"
	"fmt"
	"ispcp/internal/pinger"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/ilyakaznacheev/cleanenv"
)

var (
	configFileName *string
)

func init() {
	// Handle flags
	configFileName = flag.String("c", "./configs/config.yaml", "path to config file")
	flag.Parse()
}

//Read and parse config file
func readConfig(cfg *pinger.Config) error {
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

	err = cleanenv.ReadConfig(configFileName, cfg)
	if err != nil {
		return err
	}
	cfg.Correct()

	return nil
}

func main() {
	fmt.Println("NG:", runtime.NumGoroutine())
	config := pinger.NewConfig()

	if err := readConfig(config); err != nil {
		log.Fatalln(err)
	}

	if err := pinger.Start(config); err != nil {
		log.Fatalln(err)
	}

	//log.Printf("config: %+v\n", config)
}
