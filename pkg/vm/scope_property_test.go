package vm

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Property-based tests for Scope management.
// These tests verify the correctness properties defined in the design document.

// TestProperty16_GlobalVariableScope tests that variables declared at top level
// are stored in global scope.
// **Validates: Requirements 9.1**
// Feature: execution-engine, Property 16: グローバル変数のスコープ
func TestProperty16_GlobalVariableScope(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("variables set in global scope exist in global scope", prop.ForAll(
		func(varName string, value int) bool {
			// Create a global scope (no parent)
			globalScope := NewScope(nil)

			// Set variable in global scope
			globalScope.Set(varName, value)

			// Verify variable exists in global scope
			retrievedValue, exists := globalScope.Get(varName)
			if !exists {
				return false
			}

			// Verify value is correct
			return retrievedValue == value
		},
		gen.Identifier(),
		gen.Int(),
	))

	properties.Property("global scope has no parent", prop.ForAll(
		func(varName string, value int) bool {
			globalScope := NewScope(nil)
			globalScope.Set(varName, value)

			// Global scope should have no parent
			return globalScope.Parent() == nil
		},
		gen.Identifier(),
		gen.Int(),
	))

	properties.Property("variable in global scope is accessible from child scope", prop.ForAll(
		func(varName string, value int) bool {
			// Create global scope and set variable
			globalScope := NewScope(nil)
			globalScope.Set(varName, value)

			// Create child scope
			childScope := NewScope(globalScope)

			// Variable should be accessible from child
			retrievedValue, exists := childScope.Get(varName)
			if !exists {
				return false
			}

			return retrievedValue == value
		},
		gen.Identifier(),
		gen.Int(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty17_ScopeRoundTrip tests that local scope is created when function
// is called and destroyed when function returns.
// **Validates: Requirements 9.3, 9.4**
// Feature: execution-engine, Property 17: スコープのラウンドトリップ
func TestProperty17_ScopeRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("local scope is created with parent reference", prop.ForAll(
		func(globalVarName string, localVarName string, globalValue int, localValue int) bool {
			// Create global scope
			globalScope := NewScope(nil)
			globalScope.Set(globalVarName, globalValue)

			// Simulate function call: create local scope
			localScope := NewScope(globalScope)

			// Local scope should have parent
			if localScope.Parent() != globalScope {
				return false
			}

			// Set local variable
			localScope.SetLocal(localVarName, localValue)

			// Local variable should exist in local scope
			_, existsLocal := localScope.GetLocal(localVarName)
			if !existsLocal {
				return false
			}

			return true
		},
		gen.Identifier(),
		gen.Identifier(),
		gen.Int(),
		gen.Int(),
	))

	properties.Property("local scope destruction does not affect global scope", prop.ForAll(
		func(varName string, globalValue int, localValue int) bool {
			// Create global scope
			globalScope := NewScope(nil)
			globalScope.Set(varName, globalValue)

			// Simulate function call
			localScope := NewScope(globalScope)
			localScope.SetLocal(varName, localValue) // Shadow global variable

			// Verify local scope has local value
			localVal, _ := localScope.GetLocal(varName)
			if localVal != localValue {
				return false
			}

			// Simulate function return: local scope goes out of scope
			localScope = nil
			_ = localScope // Prevent unused variable warning

			// Global scope should still have original value
			globalVal, exists := globalScope.Get(varName)
			if !exists {
				return false
			}

			return globalVal == globalValue
		},
		gen.Identifier(),
		gen.Int(),
		gen.Int(),
	))

	properties.Property("nested function calls create independent scopes", prop.ForAll(
		func(depth int, varName string) bool {
			// Limit depth to reasonable value
			if depth > 10 {
				depth = 10
			}

			// Create global scope
			globalScope := NewScope(nil)
			globalScope.Set(varName, 0)

			// Create nested scopes (simulating nested function calls)
			scopes := make([]*Scope, depth+1)
			scopes[0] = globalScope

			for i := 1; i <= depth; i++ {
				scopes[i] = NewScope(scopes[i-1])
				scopes[i].SetLocal(varName, i) // Each scope shadows with its depth value
			}

			// Verify each scope has its own local value
			for i := 1; i <= depth; i++ {
				localVal, exists := scopes[i].GetLocal(varName)
				if !exists {
					return false
				}
				if localVal != i {
					return false
				}
			}

			// Verify global scope still has original value
			globalVal, _ := globalScope.GetLocal(varName)
			return globalVal == 0
		},
		gen.IntRange(1, 10),
		gen.Identifier(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty18_VariableResolutionPriority tests that local scope takes
// precedence over global scope when resolving variables.
// **Validates: Requirements 9.5**
// Feature: execution-engine, Property 18: 変数解決の優先順位
func TestProperty18_VariableResolutionPriority(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("local variable shadows global variable with same name", prop.ForAll(
		func(varName string, globalValue int, localValue int) bool {
			// Ensure values are different to test shadowing
			if globalValue == localValue {
				localValue = globalValue + 1
			}

			// Create global scope with variable
			globalScope := NewScope(nil)
			globalScope.Set(varName, globalValue)

			// Create local scope with same variable name
			localScope := NewScope(globalScope)
			localScope.SetLocal(varName, localValue)

			// Get should return local value (shadowing)
			retrievedValue, exists := localScope.Get(varName)
			if !exists {
				return false
			}

			return retrievedValue == localValue
		},
		gen.Identifier(),
		gen.Int(),
		gen.Int(),
	))

	properties.Property("global variable is accessible when no local shadow exists", prop.ForAll(
		func(globalVarName string, localVarName string, globalValue int, localValue int) bool {
			// Ensure different variable names
			if globalVarName == localVarName {
				localVarName = localVarName + "_local"
			}

			// Create global scope with variable
			globalScope := NewScope(nil)
			globalScope.Set(globalVarName, globalValue)

			// Create local scope with different variable
			localScope := NewScope(globalScope)
			localScope.SetLocal(localVarName, localValue)

			// Get global variable from local scope should work
			retrievedValue, exists := localScope.Get(globalVarName)
			if !exists {
				return false
			}

			return retrievedValue == globalValue
		},
		gen.Identifier(),
		gen.Identifier(),
		gen.Int(),
		gen.Int(),
	))

	properties.Property("deeply nested scope resolves to nearest shadow", prop.ForAll(
		func(varName string, values []int) bool {
			if len(values) < 2 {
				return true
			}
			// Limit to reasonable depth
			if len(values) > 10 {
				values = values[:10]
			}

			// Create scope chain
			scopes := make([]*Scope, len(values))
			scopes[0] = NewScope(nil)
			scopes[0].Set(varName, values[0])

			for i := 1; i < len(values); i++ {
				scopes[i] = NewScope(scopes[i-1])
				scopes[i].SetLocal(varName, values[i])
			}

			// Get from deepest scope should return deepest value
			deepestScope := scopes[len(scopes)-1]
			retrievedValue, exists := deepestScope.Get(varName)
			if !exists {
				return false
			}

			return retrievedValue == values[len(values)-1]
		},
		gen.Identifier(),
		gen.SliceOfN(5, gen.Int()),
	))

	properties.Property("GetLocal only returns local scope value", prop.ForAll(
		func(varName string, globalValue int, localValue int) bool {
			// Create global scope with variable
			globalScope := NewScope(nil)
			globalScope.Set(varName, globalValue)

			// Create local scope without the variable
			localScope := NewScope(globalScope)

			// GetLocal should not find the global variable
			_, existsLocal := localScope.GetLocal(varName)
			if existsLocal {
				return false // Should not find global variable via GetLocal
			}

			// Now set local variable
			localScope.SetLocal(varName, localValue)

			// GetLocal should find local variable
			retrievedValue, existsLocal := localScope.GetLocal(varName)
			if !existsLocal {
				return false
			}

			return retrievedValue == localValue
		},
		gen.Identifier(),
		gen.Int(),
		gen.Int(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
