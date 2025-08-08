// Package goresttest provides REST API testing functionality
package goresttest

// TestRunner provides the main interface for running REST API tests
type TestRunner struct {
	executor *TestExecutor
}

// NewTestRunner creates a new TestRunner with the specified base URL
func NewTestRunner(baseURL string) *TestRunner {
	return &TestRunner{
		executor: NewTestExecutor(baseURL),
	}
}

// RunTestSuite executes a test suite and returns the results
func (tr *TestRunner) RunTestSuite(suite *TestSuite) ([]*TestResult, error) {
	return tr.executor.ExecuteTestSuite(suite)
}

// RunTest executes a single test and returns the result
func (tr *TestRunner) RunTest(test Test, variables map[string]string) (*TestResult, error) {
	if variables != nil {
		tr.executor.globalVariables = variables
	}
	return tr.executor.executeTest(test)
}