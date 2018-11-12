package util

import (
	"github.com/MainframeHQ/swarmer/models"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// IConfigParser is the interface to use for parsing file configurations.
type IConfigParser interface {
	ParseYamlConfig(path string) (models.Config, error)
}

// ConfigParser is the struct for this implementation of IConfigParser.
type ConfigParser struct {
}

// GetConfigParser returns a pointer to an implementation of IConfigParser.
func GetConfigParser() *ConfigParser {
	var c = ConfigParser{}

	return &c
}

// ParseYamlConfig takes a string path to the yaml config file and returns a config model.
func (c *ConfigParser) ParseYamlConfig(path string) (models.Config, error) {
	var config models.Config

	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(yamlFile, &config)

	return config, err
}
