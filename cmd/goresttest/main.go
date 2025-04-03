package main

import (
	"fmt"
	"goresttest/internal/config"
	"goresttest/internal/report"
	"goresttest/internal/runner"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go-resttest <test-file.yaml>")
		os.Exit(1)
	}

	testFile := os.Args[1]
	testCases, err := config.LoadTestCases(testFile)
	if err != nil {
		fmt.Printf("Error loading test cases: %v\n", err)
		os.Exit(1)
	}

	var results []runner.TestResult
	for _, tc := range testCases {
		results = append(results, runner.RunTest(tc))
	}

	report.PrintResults(results)

	// Exit with non-zero if any tests failed
	for _, result := range results {
		if !result.Passed {
			os.Exit(1)
		}
	}
}
