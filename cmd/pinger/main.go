package main

import (
	"flag"
	"ispcp/internal/pinger"
	"log"
	"os"
	"path/filepath"

	_ "net/http/pprof"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/xlab/closer"
)

var (
	configFileName *string
)

func cleanup() {
	pinger.Stop()
}

func init() {
	closer.Bind(cleanup)

	// Handle flags
	configFileName = flag.String("c", "./configs/config.yaml", "path to config file")
	flag.Parse()

	//
	// go func() {
	// 	log.Println(http.ListenAndServe(":6060", nil))
	// }()
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
	config := pinger.NewConfig()

	if err := readConfig(config); err != nil {
		log.Fatalln(err)
	}

	if err := pinger.Start(config); err != nil {
		log.Fatalln(err)
	}

	closer.Hold()
}
