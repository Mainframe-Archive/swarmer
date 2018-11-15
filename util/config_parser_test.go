package util

import "testing"

func TestGetConfigParser(t *testing.T) {
	parser := GetConfigParser()

	var i interface{} = parser
	_, ok := i.(IConfigParser)

	if !ok {
		t.Error("GetConfigParser doesn't return an implementation of IConfigParser")
	}
}

func TestConfigParser_ParseYamlConfig(t *testing.T) {
	parser := GetConfigParser()

	_, err := parser.ParseYamlConfig("../swarmer.yml")

	if err != nil {
		t.Errorf("Error parsing yml file %s", err.Error())
	}
}