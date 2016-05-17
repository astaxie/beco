package main

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

func parseConifg(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var config Config
	err = toml.Unmarshal(buf, &config)
	return &config, err
}

// Config holds the beco's config
type Config struct {
	Port   int     `toml:"port"`
	Addr   string  `toml:"addr"`
	Proxys []Proxy `toml:"proxys"`
	SSL    SSL     `toml:"ssl"`
	Auth   Auth    `toml:"auth"`
}

// Proxy holds the config for backend host
type Proxy struct {
	Prefix     string    `toml:"prefix"`
	Auth       Auth      `toml:"auth"`
	Backends   []Backend `toml:"backends"`
	SetHeaders []Header  `toml:"setheaders"`
}

// SSL holds the SSL configuration
type SSL struct {
	Addr string `toml:"addr"`
	Port int    `toml:"port"`
	Cert string `toml:"cert"`
	Key  string `toml:"key"`
}

// Header is the http Header. Key:Value
type Header struct {
	Key   string `toml:"key"`
	Value string `toml:"value"`
}

// Backend is the proxy host
type Backend struct {
	//support local static files with the prefix static:
	Host        string        `toml:"host"`
	Weight      int           `toml:"weight"`
	MaxFails    int           `toml:"max_fails"`
	FailTimeout time.Duration `toml:"fail_timeout"`
	Backup      bool          `toml:"backup"`
	HealthCheck HealthCheck   `toml:"health_check"`
}

// HealthCheck to Match the response
type HealthCheck struct {
	Status  int      `toml:"status"`
	Body    string   `toml:"body"`
	Herders []Header `toml:"headers"`
}

// Auth provider
type Auth struct {
	Provider string `toml:"provider"`
	Config   string `toml:"config"`
}
