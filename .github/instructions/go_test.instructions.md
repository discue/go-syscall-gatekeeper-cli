---
applyTo: "**/*_test.go"
---
# Go Test Instructions
This file provides instructions for writing and running tests in Go. It is intended to be used as a guide for developers working on Go projects.

## Writing Tests
1. **Test File Naming**: Test files should be named with the `_test.go` suffix. For example, if you have a file named `math.go`, the corresponding test file should be named `math_test.go`.
2. **Test Function Naming**: Test functions should start with the word `Test` followed by the name of the function being tested. For example, to test a function named `Add`, the test function should be named `TestAdd`.
3. **Test Function Signature**: Test functions should have the following signature:
   ```go
   func TestFunctionName(t *testing.T) {
       // test code
   }
   ```
4. **Using the `testify` Package**: You must use the `testify` package for assertions. You can install it using:
   Example usage:
   ```go
   import (
       "testing"
       a "github.com/stretchr/testify/assert"
   )

   func TestAdd(t *testing.T) {
       assert := a.New(t)
       result := Add(2, 3)
       assert.Equal(t, 5, result, "they should be equal")
   }
   ```

5. **Table-Driven Tests**: For functions with multiple test cases, use table-driven tests to organize your test cases. Example:
```go
func TestAdd(t *testing.T) {
        tests := []struct {
            name     string
            a, b     int
            expected int
        }{
            {"positive numbers", 2, 3, 5},
            {"negative numbers", -2, -3, -5},
            {"mixed numbers", -2, 3, 1},
        }

        for _, tt := range tests {
            t.Run(tt.name, func(t *testing.T) {
            result := Add(tt.a, tt.b)
            assert.Equal(t, tt.expected, result)
            })
        }
}
```
6. **Tenant and User Isolation Tests**:
When writing tests for functions that involve tenant and user isolation, ensure that you create separate test cases for each tenant and user context. This helps verify that the function behaves correctly under different isolation scenarios.
Example:
```go
func TestGetUserData(t *testing.T) {
    tests := []struct {
        name       string
        tenantID   string
        userID     string
        expected   UserData
        expectError bool
    }{
        {"tenant1 user1", "tenant1", "user1", UserData{...}, false},
        {"tenant2 user2", "tenant2", "user2", UserData{...}, false},
        {"tenant1 user2", "tenant1", "user2", UserData{...}, true},
        {"tenant2 user1", "tenant2", "user1", UserData{...}, true},
        {"tenant2 user1", "tenant3", "user1", UserData{...}, false},
    }

    // create test data for each tenant and user context
    for _, tt := range tests {
        createTestData(tt.tenantID, tt.userID, tt.expected)
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := GetUserData(tt.tenantID, tt.userID)
            if tt.expectError {
                assert.Error(t, err)
                return
            }
            assert.Equal(t, tt.expected, result)
        })
    }
}
```
7. **Happy Paths and Edge Cases**:
Create test cases for both happy paths (normal cases) and edge cases (boundary conditions, invalid inputs, etc.) to ensure comprehensive test coverage. Example:
```go
func TestDivide(t *testing.T) {
    tests := []struct {
        name       string
        numerator   int
        denominator int
        expected    int
        expectError bool
    }{
        {"normal case", 6, 2, 3, false},
        {"division by zero", 6, 0, 0, true},
        {"negative numbers", -6, -2, 3, false},
        {"mixed numbers", -6, 2, -3, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Divide(tt.numerator, tt.denominator)
            if tt.expectError {
                assert.Error(t, err)
                return
            }
            assert.Equal(t, tt.expected, result)
        })
    }
}
```
8. **Comprehensive Coverage**: 
Ensure that tests cover the following scenarios:
   - Normal cases (happy paths)
   - Edge cases 
     - Empty Inputs
     - Invalid Inputs (test lower and upper bounds)
     - Invalid Combination of Inputs
   - Tenant and user isolation scenarios (if applicable, verify that data is correctly isolated between tenants and users, e.g., tenant1 cannot access tenant2's data, user1 cannot access user2's data, etc.)