package main

import (
	"testing"
)

// ExampleTest - Simple example test structure
func TestExample(t *testing.T) {
	// Arrange - Set up test data
	expectedValue := "test"

	// Act - Perform the action
	actualValue := "test"

	// Assert - Verify the result
	if actualValue != expectedValue {
		t.Errorf("Expected %s, got %s", expectedValue, actualValue)
	}
}

// TestWithSubtests - Example using subtests
func TestWithSubtests(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"positive number", 5, 5},
		{"zero", 0, 0},
		{"negative number", -5, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

// Example - This appears in go test -v output
// Example:
// $ go test -v -run Example
// Output:
// === RUN   Example
// --- PASS: Example (0.00s)
func Example() {
	// Example code here
	// println("example")
	// Output:
}

// BenchmarkExample - Benchmark test template
func BenchmarkExample(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// Code to benchmark
		_ = "test"
	}
}

// TestError - Test error handling
func TestError(t *testing.T) {
	// Test when error should be nil
	var err error
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Test when error should occur
	if err == nil {
		t.Log("no error as expected")
	}
}

// TestParallel - Test that can run in parallel
func TestParallel(t *testing.T) {
	t.Parallel()
	// This test will run in parallel with other parallel tests
	result := 1 + 1
	if result != 2 {
		t.Error("math is broken")
	}
}

// TestSkip - Test that can be skipped
func TestSkip(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	// Expensive test code
}

/*
To run these tests:

# Run all tests
go test -v

# Run specific test
go test -v -run TestExample

# Run with coverage
go test -v -cover -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run with race detector
go test -v -race

# Run benchmarks
go test -v -bench=. -benchmem

# Run parallel tests
go test -v -count=1

# Run only short tests
go test -short -v
*/
