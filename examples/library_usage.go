package main

import (
	"fmt"
	"log"

	"github.com/the-sumeet/goresttest"
)

func main() {
	// Example 1: Parse and run a test suite from file
	fmt.Println("=== Example 1: Running test suite from file ===")
	suite, err := goresttest.ParseTestSuite("api_tests.yml")
	if err != nil {
		log.Printf("Error parsing test suite: %v", err)
	} else {
		runner := goresttest.NewTestRunner(suite.BaseURL)
		results, err := runner.RunTestSuite(suite)
		if err != nil {
			log.Printf("Error running test suite: %v", err)
		} else {
			reporter := goresttest.NewReporter()
			reporter.PrintConsoleReport(results)
		}
	}

	fmt.Println("\n=== Example 2: Running individual tests programmatically ===")

	// Example 2: Create and run individual tests programmatically
	runner := goresttest.NewTestRunner("https://jsonplaceholder.typicode.com")

	test := goresttest.Test{
		Name:   "Get Post by ID",
		Method: "GET",
		URL:    "/posts/1",
		Headers: map[string]string{
			"Accept": "application/json",
		},
		Assertions: []goresttest.Assertion{
			{
				Type:     "status_code",
				Expected: 200,
			},
			{
				Type:     "json_path",
				Path:     "$.userId",
				Expected: 1,
			},
			{
				Type:     "json_path",
				Path:     "$.title",
				Operator: "contains",
				Expected: "sunt",
			},
		},
	}

	result, err := runner.RunTest(test, nil)
	if err != nil {
		log.Printf("Error running test: %v", err)
	} else {
		fmt.Printf("Test %s: %s\n", result.Name, getTestStatus(result.Success))
		if !result.Success {
			fmt.Printf("  Error: %s\n", result.Error)
		}
	}

	fmt.Println("\n=== Example 3: Using variable extraction and chaining ===")

	// Example 3: Variable extraction and test chaining
	createTest := goresttest.Test{
		Name:   "Create New Post",
		Method: "POST",
		URL:    "/posts",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `{"title": "Test Post", "body": "This is a test", "userId": 1}`,
		Assertions: []goresttest.Assertion{
			{
				Type:     "status_code",
				Expected: 201,
			},
		},
		Extract: map[string]string{
			"post_id": "json:$.id",
		},
	}

	createResult, err := runner.RunTest(createTest, nil)
	if err != nil {
		log.Printf("Error creating post: %v", err)
	} else {
		fmt.Printf("Test %s: %s\n", createResult.Name, getTestStatus(createResult.Success))
		if createResult.Success {
			fmt.Printf("  Extracted post_id: %s\n", createResult.Variables["post_id"])

			// Use extracted variable in next test
			getTest := goresttest.Test{
				Name:   "Get Created Post",
				Method: "GET",
				URL:    "/posts/${post_id}",
				Assertions: []goresttest.Assertion{
					{
						Type:     "status_code",
						Expected: 200,
					},
				},
			}

			getResult, err := runner.RunTest(getTest, createResult.Variables)
			if err != nil {
				log.Printf("Error getting created post: %v", err)
			} else {
				fmt.Printf("Test %s: %s\n", getResult.Name, getTestStatus(getResult.Success))
			}
		}
	}
}

func getTestStatus(success bool) string {
	if success {
		return "✓ PASS"
	}
	return "✗ FAIL"
}
