package config

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type Source string

const (
	SourceDataDog       Source = "datadog"
	SourceElasticSearch Source = "elasticsearch"
)

type AuthHeaderSpecification struct {
	Value   *string  `toml:"value"`
	Command []string `toml:"command"`
}

func (ahs *AuthHeaderSpecification) GetValue() string {
	if ahs.Value != nil {
		return *ahs.Value
	}

	if len(ahs.Command) > 0 {
		// FIXME implement it
		panic("not implemented yet")
	}

	return ""
}

type Authorization struct {
	Header map[string]AuthHeaderSpecification `toml:"header"`

	Basic *struct {
		User     string `toml:"user"`
		Password string `toml:"password"`
	} `toml:"basic"`

	Cloud *struct {
		CloudID string `toml:"cloud_id"`
		APIKey  string `toml:"api_key"`
	} `toml:"cloud"`
}

type DatadogEnv struct {
	DDAPIKey string `toml:"dd_api_key"`
	DDAppKey string `toml:"dd_personal_key"`
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
	Order         string         `toml:"order"`
	Source        Source         `toml:"source"`

	DatadogEnv
}

type Output struct {
	Format    string          `toml:"format"`
	Exclude   []string        `toml:"exclude"`
	Only      []string        `toml:"only"`
	D         interface{}     `toml:"decode_recursively"`
	Decode    map[string]bool `toml:"-"`
	IsDefault bool            `toml:"default"`
}

type Config struct {
	Envs    map[string]*Env    `toml:"env"`
	Outputs map[string]*Output `toml:"output"`
	Aliases map[string]string  `toml:"aliases"`
}

func FromStringList(l []string) map[string]bool {
	vv := map[string]bool{}
	for _, v := range l {
		vv[v] = true
	}

	return vv
}

// copy-pasted from https://github.com/elastic/go-elasticsearch/blob/7.17/elasticsearch.go#L447
// addrFromCloudID extracts the Elasticsearch URL from CloudID.
// See: https://www.elastic.co/guide/en/cloud/current/ec-cloud-id.html
//
func addrFromCloudID(input string) (string, error) {
	var scheme = "https://"

	values := strings.Split(input, ":")
	if len(values) != 2 {
		return "", fmt.Errorf("unexpected format: %q", input)
	}
	data, err := base64.StdEncoding.DecodeString(values[1])
	if err != nil {
		return "", err
	}
	parts := strings.Split(string(data), "$")

	if len(parts) < 2 {
		return "", fmt.Errorf("invalid encoded value: %s", parts)
	}

	return fmt.Sprintf("%s%s.%s", scheme, parts[1], parts[0]), nil
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
	if e.Source == SourceDataDog {
		return "timestamp"
	}

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

		return e, nil
	}

	for _, v := range c.Envs {
		if v.IsDefault {
			return v, nil
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

		return o, nil
	}

	for _, v := range c.Outputs {
		if v.IsDefault {
			return v, nil
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
		if v.Source == "" {
			v.Source = SourceElasticSearch
		}

		if auth := v.Authorization; auth != nil {
			authHeader := ""
			if auth.Basic != nil {
				authHeader = fmt.Sprintf(
					"Basic %s",
					base64.StdEncoding.EncodeToString([]byte(auth.Basic.User+":"+auth.Basic.Password)),
				)
			}

			if auth.Cloud != nil {
				ep, err := addrFromCloudID(auth.Cloud.CloudID)
				if err != nil {
					return fmt.Errorf("env='%s' has invalid cloud_id: %w", k, err)
				}

				v.Endpoints = []string{ep}

				authHeader = "APIKey " + auth.Cloud.APIKey
			}

			if authHeader != "" {
				if auth.Header == nil {
					auth.Header = map[string]AuthHeaderSpecification{}
				}

				auth.Header["Authorization"] = AuthHeaderSpecification{Value: &authHeader}
			}
		}

		if len(v.Endpoints) == 0 {
			return fmt.Errorf("env='%s' has zero endpoints", k)
		}
	}

	for k, v := range c.Outputs {
		switch vv := v.D.(type) {
		case nil:
			v.Decode = map[string]bool{}

		case bool:
			if vv {
				v.Decode = map[string]bool{
					"http": true,
					"json": true,
				}
			}

		case []string:
			v.Decode = map[string]bool{}
			for _, vvv := range vv {
				v.Decode[vvv] = true
			}

		case []interface{}:
			v.Decode = map[string]bool{}
			for _, vvv := range vv {
				v.Decode[fmt.Sprint(vvv)] = true
			}

		default:
			return fmt.Errorf("can't handle %v (type=%T) as decode_recursively in output='%s', allowed values are bool or []string", vv, vv, k)
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
