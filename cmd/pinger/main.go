package main

import (
	"flag"
	"ispcp/internal/pinger"
	"os/signal"
	"syscall"

	"os"
	"path/filepath"
	"time"

	_ "net/http/pprof"

	"github.com/facebookgo/pidfile"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/jinzhu/copier"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	configFileName *string
	config         = pinger.NewConfig()
)

// Close the program carefully
func clean() {
	pidFile := pidfile.GetPidfilePath()

	err := os.Remove(pidFile)
	if err != nil {
		log.Error().Msg("Can't remove PID file: " + pidFile)
	}
	pinger.Clean()
}

// Reload config
func reloadConfig() {
	tmp := pinger.NewConfig()
	if err := loadConfig(tmp); err != nil {
		log.Error().Msg("can't load config: " + err.Error())
		return
	}
	config.Lock()
	defer config.Unlock()

	copier.Copy(config, tmp)
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
	// Setup logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.Stamp})
	// Check if a copy of the program is running
	launched, err := isLaunched()
	if err != nil {
		log.Error().Msg("Error check PID file: " + err.Error())
		os.Exit(1)
	}
	if launched {
		log.Error().Msg("The program is already running, or delete the PID file.")
		os.Exit(1)
	}
	// Handle flags
	configFileName = flag.String("c", "./configs/config.yaml", "path to config file")
	flag.Parse()
	// Handle syscalls
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range sigCh {
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				log.Info().Msg("Terminate the program.")
				clean()
				os.Exit(0)
			case syscall.SIGHUP:
				log.Info().Msg("Reloading configuration.")
				reloadConfig()
			default:
				log.Warn().Msg("Received an unprocessed signal ")
			}
		}
	}()
	//
	err = loadConfig(config)
	if err != nil {
		log.Error().Msg("can't load config: " + err.Error())
		clean()
		os.Exit(1)
	}

	// pprof instance
	// go func() {
	// 	log.Println(http.ListenAndServe(":6060", nil))
	// }()
}

//Read and parse config file
func loadConfig(cfg *pinger.Config) error {
	defer cfg.Unlock()

	configFileName, err := filepath.Abs(*configFileName)
	if err != nil {
		return err
	}
	log.Info().Msg("Loading config: " + configFileName)

	configFile, err := os.Open(configFileName)
	if err != nil {
		return err
	}
	defer configFile.Close()

	cfg.Lock()
	err = cleanenv.ReadConfig(configFileName, cfg)
	if err != nil {
		return err
	}
	cfg.Correct()
	return nil
}

func main() {
	defer clean()

	if err := pinger.Start(config); err != nil {
		clean()
		log.Error().Msg("Can't start main : " + err.Error())
		os.Exit(1)
	}
}
