package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Validation struct {
	StatusCode string `yaml:"status_code"`
}
type TestCase struct {
	Name       string            `yaml:"name"`
	URL        string            `yaml:"url"`
	Method     string            `yaml:"method"`
	Headers    map[string]string `yaml:"headers"`
	Body       string            `yaml:"body"`
	Validation Validation        `yaml:"validation"`
}

func LoadTestCases(filePath string) ([]TestCase, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var testCases []TestCase
	err = yaml.Unmarshal(data, &testCases)
	if err != nil {
		return nil, err
	}

	return testCases, nil
}
