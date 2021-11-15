package pinger

const (
	timesToRetry = 3
	// concurrentMax        = 128
	// concurrentDefault    = 16
	//jobChanLen           = 0
	restartIntervalMin   = 10
	fileStoreNameDefault = "store.txt"
)

type Config struct {
	ExcludeIfaceNames []string `yaml:"exclude-ifaces"`
	ExcludeNetIPs     []string `yaml:"exclude-networks"`
	//	ThreadsNumber     int      `yaml:"threads"`
	RestartInterval int    `yaml:"restart-interval"`
	URI             string `yaml:"dsn"`
	StoreType       string `yaml:"store"`
	FileStoreName   string `yaml:"file-store-name"`
	LoggingLevel    string `yaml:"log-level"`
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
}
