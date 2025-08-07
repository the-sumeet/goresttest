package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var (
		configFile   = flag.String("config", "", "Path to YAML test configuration file")
		outputFormat = flag.String("output", "console", "Output format: console, json, html")
		outputFile   = flag.String("file", "", "Output file path (for json/html formats)")
		parallel     = flag.Bool("parallel", false, "Run tests in parallel")
		maxWorkers   = flag.Int("workers", 10, "Maximum number of parallel workers")
		verbose      = flag.Bool("verbose", false, "Verbose output")
		version      = flag.Bool("version", false, "Show version information")
	)
	
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "GoRestTest - API Testing Framework\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -config tests.yaml\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -config tests.yaml -parallel -workers 5\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -config tests.yaml -output html -file report.html\n", os.Args[0])
	}
	
	flag.Parse()
	
	if *version {
		fmt.Println("GoRestTest v1.0.0")
		fmt.Println("API Testing Framework inspired by pyresttest")
		return
	}
	
	if *configFile == "" {
		fmt.Fprintf(os.Stderr, "Error: -config flag is required\n\n")
		flag.Usage()
		os.Exit(1)
	}
	
	if _, err := os.Stat(*configFile); os.IsNotExist(err) {
		log.Fatalf("Config file does not exist: %s", *configFile)
	}
	
	suite, err := parseTestSuite(*configFile)
	if err != nil {
		log.Fatalf("Failed to parse test suite: %v", err)
	}
	
	if *parallel {
		suite.Parallel = true
		if *maxWorkers > 0 {
			suite.MaxWorkers = *maxWorkers
		}
	}
	
	if *verbose {
		fmt.Printf("Loaded test suite: %s\n", suite.Name)
		fmt.Printf("Base URL: %s\n", suite.BaseURL)
		fmt.Printf("Tests: %d\n", len(suite.Tests))
		fmt.Printf("Parallel: %t\n", suite.Parallel)
		if suite.Parallel {
			fmt.Printf("Max Workers: %d\n", suite.MaxWorkers)
		}
		fmt.Println()
	}
	
	var testResults []*TestResult
	
	if len(suite.Tests) > 0 {
		executor := NewTestExecutor(suite.BaseURL)
		testResults, err = executor.ExecuteTestSuite(suite)
		if err != nil {
			log.Fatalf("Failed to execute tests: %v", err)
		}
	}
	
	reporter := NewReporter()
	
	switch strings.ToLower(*outputFormat) {
	case "console":
		reporter.PrintConsoleReport(testResults)
		
	case "json":
		filename := *outputFile
		if filename == "" {
			filename = "test-report.json"
		}
		
		if err := reporter.GenerateJSONReport(testResults, filename); err != nil {
			log.Fatalf("Failed to generate JSON report: %v", err)
		}
		
		fmt.Printf("JSON report generated: %s\n", filename)
		
	case "html":
		filename := *outputFile
		if filename == "" {
			filename = "test-report.html"
		}
		
		if err := reporter.GenerateHTMLReport(testResults, filename); err != nil {
			log.Fatalf("Failed to generate HTML report: %v", err)
		}
		
		fmt.Printf("HTML report generated: %s\n", filename)
		absPath, _ := filepath.Abs(filename)
		fmt.Printf("Open in browser: file://%s\n", absPath)
		
	default:
		log.Fatalf("Unsupported output format: %s", *outputFormat)
	}
	
	exitCode := 0
	for _, result := range testResults {
		if !result.Success {
			exitCode = 1
			break
		}
	}
	
	os.Exit(exitCode)
}