// SPDX-License-Identifier: BSD-2-Clause
package config

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Server    Server              `toml:"server"`
	Limits    Limits              `toml:"limits"`
	Logging   Logging             `toml:"logging"`
	Operators map[string]Operator `toml:"operators"`
	Channels  Channels            `toml:"channels"`
}

type Server struct {
	Name      string   `toml:"name"`
	Network   string   `toml:"network"`
	Hostname  string   `toml:"hostname"`
	MOTDFile  string   `toml:"motd_file"`
	PassOn    bool     `toml:"server_pass_on"`
	Password  string   `toml:"server_pass"`
	ListenTCP string   `toml:"listen_tcp"`
	ListenWS  string   `toml:"listen_ws"`
	WSPath    string   `toml:"ws_path"`
	TLS       TLSBlock `toml:"tls"`
}

type TLSBlock struct {
	Enable   bool   `toml:"enable"`
	CertFile string `toml:"cert_file"`
	KeyFile  string `toml:"key_file"`
}

type Limits struct {
	MaxLineBytes        int           `toml:"max_line_bytes"`
	MaxConnections      int           `toml:"max_connections"`
	PingInterval        time.Duration `toml:"ping_interval"`
	PingTimeout         time.Duration `toml:"ping_timeout"`
	WriteQueue          int           `toml:"write_queue"`
	RegistrationTimeout time.Duration `toml:"registration_timeout"`
}

type Logging struct {
	Level string `toml:"level"`
	JSON  bool   `toml:"json"`
}

type Operator struct {
	PasswordBcrypt string `toml:"password_bcrypt"`
}

type Channels struct {
	Default  []string `toml:"default"`
	Prefixes string   `toml:"prefixes"`
}

func defaultConfig() Config {
	return Config{
		Server: Server{
			Name:      "aircd",
			Network:   "AirNet",
			Hostname:  "localhost",
			MOTDFile:  "motd.txt",
			PassOn:    false,
			Password:  "@$$W0RD",
			ListenTCP: ":6667",
			ListenWS:  "",
			WSPath:    "/ws",
		},
		Limits: Limits{
			MaxLineBytes:        512,
			MaxConnections:      5000,
			PingInterval:        2 * time.Minute,
			PingTimeout:         30 * time.Second,
			WriteQueue:          64,
			RegistrationTimeout: 30 * time.Second,
		},
		Logging: Logging{
			Level: "info",
			JSON:  false,
		},
		Operators: map[string]Operator{},
		Channels: Channels{
			Default:  []string{"#tomopona"},
			Prefixes: "#",
		},
	}
}

func save(path string, cfg *Config) error {
	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func Load(path string) (*Config, bool, error) {
	cfg := defaultConfig()

	if path == "" {
		// no file: just return defaults
		if err := cfg.Validate(); err != nil {
			return nil, false, err
		}
		return &cfg, false, nil
	}

	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// create new file
			if err := save(path, &cfg); err != nil {
				return nil, false, err
			}
			if err := cfg.Validate(); err != nil {
				return nil, false, err
			}
			return &cfg, true, nil
		}
		return nil, false, err
	}

	if err := toml.Unmarshal(b, &cfg); err != nil {
		return nil, false, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, false, err
	}
	return &cfg, false, nil
}

func (c *Config) Validate() error {
	if c.Server.ListenTCP == "" && c.Server.ListenWS == "" {
		return errors.New("at least one listener (tcp or ws) must be set")
	}
	if c.Limits.MaxLineBytes <= 0 || c.Limits.MaxLineBytes > 2048 {
		return errors.New("limits.max_line_bytes must be in (0, 2048]")
	}
	return nil
}
