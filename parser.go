package goresttest

import (
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseTestSuite parses a YAML test configuration file and returns a TestSuite
func ParseTestSuite(filename string) (*TestSuite, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var suite TestSuite
	if err := yaml.Unmarshal(data, &suite); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if suite.MaxWorkers <= 0 {
		suite.MaxWorkers = 10
	}

	return &suite, nil
}

// ParseTestSuiteFromString parses a YAML string and returns a TestSuite
func ParseTestSuiteFromString(yamlContent string) (*TestSuite, error) {
	var suite TestSuite
	if err := yaml.Unmarshal([]byte(yamlContent), &suite); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if suite.MaxWorkers <= 0 {
		suite.MaxWorkers = 10
	}

	return &suite, nil
}

// InterpolateVariables replaces variable placeholders in text with their values
func InterpolateVariables(text string, variables map[string]string) string {
	result := text
	for key, value := range variables {
		placeholder := fmt.Sprintf("${%s}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}