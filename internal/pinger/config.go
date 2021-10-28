package pinger

const concurrentMax = 8

type Config struct {
	ExcludeIfaceNames []string `yaml:"exclude-ifaces"`
	TestNetwork       string   `yaml:"test-network"`
	ThreadsNumber     int      `yaml:"threads"`
}

func NewConfig() *Config {
	return &Config{}
}

func (cfg *Config) Correct() {
	if cfg.ThreadsNumber > concurrentMax {
		cfg.ThreadsNumber = concurrentMax
	}
}
