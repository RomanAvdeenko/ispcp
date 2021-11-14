package main

import (
	"flag"
	"fmt"
	"ispcp/internal/pinger"
	"log"
	"os"
	"path/filepath"

	_ "net/http/pprof"

	"github.com/facebookgo/pidfile"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/xlab/closer"
)

var (
	configFileName *string
)

// Handle termination SIGs
func cleanup() {
	pidFile := pidfile.GetPidfilePath()

	err := os.Remove(pidFile)
	if err != nil {
		fmt.Println("Can't remove PID file: ", pidFile)
	}
	pinger.Stop()
}

// Check the launch of one instance of the program
func isLaunched() (bool, error) {
	var err error
	name, err := os.Executable()
	if err != nil {
		return false, err
	}
	pidfile.SetPidfilePath(name + ".pid")
	_, err = pidfile.Read()
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			pidfile.Write()
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func init() {
	// Handle flags
	configFileName = flag.String("c", "./configs/config.yaml", "path to config file")
	flag.Parse()
	//
	closer.Bind(cleanup)

	// pprof instance
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
	launched, err := isLaunched()
	if err != nil {
		fmt.Println("Error check PID file:", err)
		os.Exit(1)
	}
	if launched {
		fmt.Println("The program is already running, or delete the PID file.")
		os.Exit(1)
	}

	config := pinger.NewConfig()

	if err := readConfig(config); err != nil {
		log.Fatalln(err)
	}

	if err := pinger.Start(config); err != nil {
		log.Fatalln(err)
	}

	closer.Hold()
}
