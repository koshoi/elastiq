package config

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type Authorization struct {
	User     string
	Password string
}

type Env struct {
	Endpoints     []string       `toml:"endpoints"`
	Authorization *Authorization `toml:"authorization"`
	Index         string         `toml:"index"`
	IsDefault     bool           `toml:"default"`
	TZ            string         `toml:"timezone"`
	Timezone      *time.Location `toml:"-"`
	TimeFormat    string         `toml:"time_format"`
	Limit         int            `toml:"limit"`
	Output        string         `toml:"output"`
}

type Output struct {
	Format    string   `toml:"format"`
	Exclude   []string `toml:"exclude"`
	Only      []string `toml:"only"`
	IsDefault bool     `toml:"default"`
}

type Config struct {
	Envs    map[string]Env    `toml:"env"`
	Outputs map[string]Output `toml:"output"`
}

func (e *Env) GetEndpoint() string {
	rand.Seed(time.Now().UTC().UnixNano())
	return e.Endpoints[rand.Intn(len(e.Endpoints))]
}

func (e *Env) GetTimezone(tz string) (*time.Location, error) {
	if tz != "" {
		return e.Timezone, nil
	}

	timezone, err := time.LoadLocation(tz)
	if err != nil {
		return nil, fmt.Errorf("failed to get timezone from string='%s': %w", tz, err)
	}

	return timezone, nil
}

func (e *Env) GetTimeFormat(tf string) string {
	if tf != "" {
		return tf
	}

	if e.TimeFormat != "" {
		return e.TimeFormat
	}

	return time.RFC3339
}

func (e *Env) GetLimit(limit int) int {
	if limit > 0 {
		return limit
	}

	if e.Limit > 0 {
		return e.Limit
	}

	return 10
}

func (c *Config) GetEnv(env string) (*Env, error) {
	if env != "" {
		e, ok := c.Envs[env]
		if !ok {
			return nil, fmt.Errorf("env='%s' not found", env)
		}

		return &e, nil
	}

	for _, v := range c.Envs {
		if v.IsDefault {
			return &v, nil
		}
	}

	return nil, fmt.Errorf("env was not specified and no default env was found in config")
}

func (c *Config) GetOutput(env *Env, output string) (*Output, error) {
	if output == "" {
		output = env.Output
	}

	if output != "" {
		o, ok := c.Outputs[output]
		if !ok {
			return nil, fmt.Errorf("output='%s' not found", output)
		}

		return &o, nil
	}

	for _, v := range c.Outputs {
		if v.IsDefault {
			return &v, nil
		}
	}

	return &Output{Format: "json"}, nil
}

func (c *Config) Validate() error {
	defaults := []string{}
	for k, v := range c.Envs {
		if v.IsDefault {
			defaults = append(defaults, k)
		}
	}

	if len(defaults) > 1 {
		return fmt.Errorf("only one default env is allowed, found multiple: %s", strings.Join(defaults, ", "))
	}

	defaults = []string{}
	for k, v := range c.Outputs {
		if v.IsDefault {
			defaults = append(defaults, k)
		}
	}

	if len(defaults) > 1 {
		return fmt.Errorf("only one default output is allowed, found multiple: %s", strings.Join(defaults, ", "))
	}

	for k, v := range c.Envs {
		if len(v.Endpoints) == 0 {
			return fmt.Errorf("env='%s' has zero endpoints", k)
		}
	}

	return nil
}

func ReadConfig(configPath string) (*Config, error) {
	cfg := Config{}
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return nil, fmt.Errorf("failed to read config from file='%s': %w", configPath, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}
