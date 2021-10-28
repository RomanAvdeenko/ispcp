package alive

type Config struct {
	ExcludeIfaceNames []string `yaml:"exclude-ifaces"`
}

func NewConfig() *Config {
	return &Config{}
}
