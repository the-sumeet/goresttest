package report

import (
	"fmt"
	"goresttest/internal/runner"
)

func PrintResults(results []runner.TestResult) {
	passed := 0
	for _, result := range results {
		if result.Passed {
			passed++
			fmt.Printf("[PASS] %s\n", result.TestCase.Name)
		} else {
			fmt.Printf("[FAIL] %s: %v\n", result.TestCase.Name, result.Error)
		}
	}

	fmt.Printf("\nSummary: %d/%d tests passed\n", passed, len(results))
}
