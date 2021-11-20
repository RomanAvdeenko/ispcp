package pinger

import "sync"

const (
	timesToMainLoopRetry = 3
	timesToArpIPRetry    = 3
	// concurrentMax        = 128
	// concurrentDefault    = 16
	//jobChanLen           = 0
	restartIntervalMin   = 10
	responseMaitTimeMin  = 10
	fileStoreNameDefault = "store.txt"
	locationDefault      = "Europe/Kiev"
)

var configLock = new(sync.RWMutex)

type Config struct {
	ExcludeIfaceNames []string `yaml:"exclude-ifaces"`
	ExcludeNetIPs     []string `yaml:"exclude-networks"`
	//	ThreadsNumber     int      `yaml:"threads"`
	RestartInterval  int    `yaml:"restart-interval"`
	URI              string `yaml:"dsn"`
	StoreType        string `yaml:"store"`
	FileStoreName    string `yaml:"file-store-name"`
	LoggingLevel     string `yaml:"log-level"`
	ResponseWaitTime int    `yaml:"response-wait"`
	Location         string `yaml:"location"`
}

func NewConfig() *Config {
	return &Config{}
}

func (cfg *Config) Correct() {
	// switch {
	// case cfg.ThreadsNumber > concurrentMax:
	// 	cfg.ThreadsNumber = concurrentMax
	// case cfg.ThreadsNumber == 0:
	// 	cfg.ThreadsNumber = concurrentDefault
	// }
	if cfg.ResponseWaitTime < responseMaitTimeMin {
		cfg.ResponseWaitTime = responseMaitTimeMin
	}
	if cfg.RestartInterval < restartIntervalMin {
		cfg.RestartInterval = restartIntervalMin
	}
	if cfg.FileStoreName == "" {
		cfg.FileStoreName = fileStoreNameDefault
	}
	if cfg.StoreType == "" {
		cfg.StoreType = "file"
	}
	if cfg.LoggingLevel == "" {
		cfg.LoggingLevel = "INFO"
	}
	if cfg.Location == "" {
		cfg.Location = locationDefault
	}
}

func (cfg *Config) Lock()   { configLock.Lock() }
func (cfg *Config) Unlock() { configLock.Unlock() }
