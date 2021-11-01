package pinger

const concurrentMax = 8

type Config struct {
	ExcludeIfaceNames []string `yaml:"exclude-ifaces"`
	ThreadsNumber     int      `yaml:"threads"`
	ExcludeNetIPs     []string `yaml:"exclude-networks"`
}

func NewConfig() *Config {
	return &Config{}
}

func (cfg *Config) Correct() {
	if cfg.ThreadsNumber > concurrentMax {
		cfg.ThreadsNumber = concurrentMax
	}
}
