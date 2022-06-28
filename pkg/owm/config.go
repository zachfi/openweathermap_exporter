package owm

import (
	"flag"
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	OtelEndpoint string `yaml:"otel_endpoint"`
	ListenAddr   string `yaml:"listen_addr"`

	APIKey    string     `mapstructure:"apikey"`
	Locations []Location `mapstructure:"locations"`
}

type Location struct {
	Name      string
	Latitude  float64
	Longitude float64
}

// LoadConfig receives a file path for a configuration to load.
func LoadConfig(file string) (Config, error) {
	filename, _ := filepath.Abs(file)

	config := Config{}
	err := loadYamlFile(filename, &config)
	if err != nil {
		return config, errors.Wrap(err, "failed to load yaml file")
	}

	return config, nil
}

// loadYamlFile unmarshals a YAML file into the received interface{} or returns an error.
func loadYamlFile(filename string, d interface{}) error {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(yamlFile, d)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) RegisterFlagsAndApplyDefaults(prefix string, f *flag.FlagSet) {
	f.StringVar(&c.OtelEndpoint, "otel.endpoint", "", "otel endpoint, eg: tempo:4317")
	f.StringVar(&c.ListenAddr, "listen.addr", ":9101", "address to listen on")
}
