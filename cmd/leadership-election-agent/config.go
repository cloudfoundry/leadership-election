package main

import envstruct "code.cloudfoundry.org/go-envstruct"

type Config struct {
	Port       uint16 `env:"PORT"`
	HealthPort uint16 `env:"HEALTH_PORT"`
}

func loadConfig() (Config, error) {
	cfg := Config{
		Port:       8080,
		HealthPort: 6060,
	}

	if err := envstruct.Load(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
