package main

import "time"

type TestSuite struct {
	Name        string            `yaml:"name"`
	BaseURL     string            `yaml:"base_url"`
	Variables   map[string]string `yaml:"variables"`
	Tests       []Test            `yaml:"tests"`
	Parallel    bool              `yaml:"parallel"`
	MaxWorkers  int               `yaml:"max_workers"`
}

type Test struct {
	Name        string            `yaml:"name"`
	Method      string            `yaml:"method"`
	URL         string            `yaml:"url"`
	Headers     map[string]string `yaml:"headers"`
	Body        string            `yaml:"body"`
	BodyFile    string            `yaml:"body_file"`
	Timeout     time.Duration     `yaml:"timeout"`
	Assertions  []Assertion       `yaml:"assertions"`
	Extract     map[string]string `yaml:"extract"`
	DependsOn   []string          `yaml:"depends_on"`
}

type Assertion struct {
	Type     string      `yaml:"type"`
	Path     string      `yaml:"path"`
	Expected interface{} `yaml:"expected"`
	Operator string      `yaml:"operator"`
}

type TestResult struct {
	Name       string
	Success    bool
	StatusCode int
	Duration   time.Duration
	Response   string
	Headers    map[string][]string
	Error      string
	Variables  map[string]string
}