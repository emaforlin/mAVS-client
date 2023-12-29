package cli

import (
	"errors"
	"flag"
)

type Config struct {
	Difficulty uint
	ListenPort uint
	TargetAddr string
}

func (c *Config) Validate() error {
	if c.Difficulty < 1 {
		return errors.New("low mining difficulty")
	}
	if c.ListenPort > 65535 {
		return errors.New("invalid port")
	}
	return nil
}

func ParseFlags() (Config, error) {
	config := Config{}
	flag.UintVar(&config.Difficulty, "m", 5, "Sets the mining difficulty (default: 5)")
	flag.UintVar(&config.ListenPort, "l", 0, "Wait for connections on the desired port (0 for a random port)")
	flag.StringVar(&config.TargetAddr, "d", "", "Target peer")

	flag.Parse()
	err := config.Validate()
	if err != nil {
		return Config{}, err
	}
	return config, nil
}
