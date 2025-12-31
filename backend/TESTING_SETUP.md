# Go Testing Configuration

## Backend Test Setup

### Test Packages Structure
```
backend/
├── main.go
├── main_test.go              # Main package tests
├── controllers/
│   ├── controllers.go
│   └── controllers_test.go
├── database/
│   ├── db.go
│   └── db_test.go
├── services/
│   ├── services.go
│   └── services_test.go
├── models/
│   ├── models.go
│   └── models_test.go
└── validators/
    ├── validators.go
    └── validators_test.go
```

### Required Testing Tools
```bash
# Install testing tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install gotest.tools/gotestsum@latest
go install gomodule/demo@latest
```

### Test Configuration
- Framework: Go's built-in `testing` package
- Assertions: Use `testing.T` with manual assertions
- Mocking: Use interfaces and dependency injection
- Coverage: Target >80% coverage

### Running Tests
```bash
# Run all tests with verbose output
go test -v ./...

# Run with coverage
go test -cover -coverprofile=coverage.out ./...

# Generate HTML coverage report
go tool cover -html=coverage.out

# Run with race detector
go test -race ./...

# Run specific test
go test -run TestFunctionName ./package

# Set timeout
go test -timeout 30s ./...
```

### Test Naming Convention
- Test files: `*_test.go`
- Test functions: `TestFunctionName(t *testing.T)`
- Sub-tests: `t.Run("subtest", func(t *testing.T) { ... })`

### Example Test Structure
```go
package database

import (
	"testing"
)

func TestConnection(t *testing.T) {
	t.Run("successful connection", func(t *testing.T) {
		// Arrange
		db := setupTestDB()
		defer db.Close()

		// Act
		err := db.Ping()

		// Assert
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("failed connection", func(t *testing.T) {
		// Arrange
		invalidDB := setupInvalidDB()

		// Act
		err := invalidDB.Ping()

		// Assert
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

// Helper function
func setupTestDB() *DB {
	// Setup code
	return &DB{}
}

func setupInvalidDB() *DB {
	// Setup code for invalid connection
	return &DB{}
}
```

### Mocking Database Connections
```go
// Use interfaces for dependency injection
type Repository interface {
	GetVenta(id string) (*Venta, error)
	SaveVenta(venta *Venta) error
}

// Mock implementation
type MockRepository struct {
	GetVentaFunc func(id string) (*Venta, error)
	SaveVentaFunc func(venta *Venta) error
}

func (m *MockRepository) GetVenta(id string) (*Venta, error) {
	return m.GetVentaFunc(id)
}

func (m *MockRepository) SaveVenta(venta *Venta) error {
	return m.SaveVentaFunc(venta)
}
```

### Table-Driven Tests
```go
func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		valid   bool
	}{
		{"valid email", "test@example.com", true},
		{"invalid email", "invalid", false},
		{"empty email", "", false},
		{"email with spaces", "test @example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateEmail(tt.email)
			if result != tt.valid {
				t.Errorf("ValidateEmail(%q) = %v, want %v", tt.email, result, tt.valid)
			}
		})
	}
}
```

### Benchmarking
```go
func BenchmarkFunctionName(b *testing.B) {
	// Setup (not timed)
	data := setupData()

	b.ResetTimer()
	// Benchmark code
	for i := 0; i < b.N; i++ {
		ExpensiveFunction(data)
	}
}

// Run: go test -bench=. -benchmem ./...
```

### Integration Tests
```go
// +build integration

package main

import (
	"testing"
)

func TestDatabaseIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Integration test code
	db := connectToTestDatabase()
	defer db.Close()

	// Test code...
}

// Run: go test -v -tags=integration ./...
```

### Coverage Goals
- Overall: >80%
- Critical paths (auth, payment, validation): 100%
- Utils: >75%
- Controllers: >85%

### Continuous Integration
Tests run automatically on:
- Every push to `main` or `develop`
- All pull requests
- Manual workflow trigger

Coverage reports uploaded to Codecov.io
