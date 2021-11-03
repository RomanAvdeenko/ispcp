package pinger

const (
	concurrentMax      = 256
	concurrentDefault  = 16
	jobChanLen         = 4096
	restartIntervalMin = 5
)

type Config struct {
	ExcludeIfaceNames []string `yaml:"exclude-ifaces"`
	ExcludeNetIPs     []string `yaml:"exclude-networks"`
	ThreadsNumber     int      `yaml:"threads"`
	RestartInterval   int      `yaml:"restart-interval"`
}

func NewConfig() *Config {
	return &Config{}
}

func (cfg *Config) Correct() {
	switch {
	case cfg.ThreadsNumber > concurrentMax:
		cfg.ThreadsNumber = concurrentMax
	case cfg.ThreadsNumber == 0:
		cfg.ThreadsNumber = concurrentDefault
	}

	if cfg.RestartInterval < restartIntervalMin {
		cfg.RestartInterval = restartIntervalMin
	}
}
