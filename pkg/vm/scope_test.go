package vm

import (
	"testing"
)

// TestNewScope tests the Scope constructor.
func TestNewScope(t *testing.T) {
	t.Run("creates scope without parent", func(t *testing.T) {
		scope := NewScope(nil)
		if scope == nil {
			t.Fatal("expected scope to be created")
		}
		if scope.Parent() != nil {
			t.Error("expected parent to be nil")
		}
		if scope.Size() != 0 {
			t.Errorf("expected size 0, got %d", scope.Size())
		}
	})

	t.Run("creates scope with parent", func(t *testing.T) {
		parent := NewScope(nil)
		child := NewScope(parent)
		if child.Parent() != parent {
			t.Error("expected parent to be set")
		}
	})
}

// TestScopeGetSet tests Get and Set methods.
func TestScopeGetSet(t *testing.T) {
	t.Run("sets and gets variable", func(t *testing.T) {
		scope := NewScope(nil)
		scope.Set("x", 42)

		val, ok := scope.Get("x")
		if !ok {
			t.Error("expected variable to exist")
		}
		if val != 42 {
			t.Errorf("expected 42, got %v", val)
		}
	})

	t.Run("returns false for non-existent variable", func(t *testing.T) {
		scope := NewScope(nil)
		_, ok := scope.Get("nonexistent")
		if ok {
			t.Error("expected variable to not exist")
		}
	})

	t.Run("updates existing variable", func(t *testing.T) {
		scope := NewScope(nil)
		scope.Set("x", 42)
		scope.Set("x", 100)

		val, _ := scope.Get("x")
		if val != 100 {
			t.Errorf("expected 100, got %v", val)
		}
	})

	t.Run("gets variable from parent scope", func(t *testing.T) {
		parent := NewScope(nil)
		parent.Set("x", 42)

		child := NewScope(parent)
		val, ok := child.Get("x")
		if !ok {
			t.Error("expected variable to exist in parent")
		}
		if val != 42 {
			t.Errorf("expected 42, got %v", val)
		}
	})

	t.Run("child variable shadows parent", func(t *testing.T) {
		parent := NewScope(nil)
		parent.Set("x", 42)

		child := NewScope(parent)
		child.SetLocal("x", 100)

		val, _ := child.Get("x")
		if val != 100 {
			t.Errorf("expected 100 (child value), got %v", val)
		}

		// Parent should still have original value
		parentVal, _ := parent.Get("x")
		if parentVal != 42 {
			t.Errorf("expected parent to have 42, got %v", parentVal)
		}
	})

	t.Run("updates variable in parent scope", func(t *testing.T) {
		parent := NewScope(nil)
		parent.Set("x", 42)

		child := NewScope(parent)
		child.Set("x", 100) // Should update parent's x

		// Parent should have updated value
		parentVal, _ := parent.Get("x")
		if parentVal != 100 {
			t.Errorf("expected parent to have 100, got %v", parentVal)
		}
	})
}

// TestScopeGetSetLocal tests GetLocal and SetLocal methods.
func TestScopeGetSetLocal(t *testing.T) {
	t.Run("sets and gets local variable", func(t *testing.T) {
		scope := NewScope(nil)
		scope.SetLocal("x", 42)

		val, ok := scope.GetLocal("x")
		if !ok {
			t.Error("expected variable to exist locally")
		}
		if val != 42 {
			t.Errorf("expected 42, got %v", val)
		}
	})

	t.Run("GetLocal does not search parent", func(t *testing.T) {
		parent := NewScope(nil)
		parent.Set("x", 42)

		child := NewScope(parent)
		_, ok := child.GetLocal("x")
		if ok {
			t.Error("expected GetLocal to not find parent variable")
		}
	})

	t.Run("SetLocal creates in current scope only", func(t *testing.T) {
		parent := NewScope(nil)
		parent.Set("x", 42)

		child := NewScope(parent)
		child.SetLocal("x", 100)

		// Child should have its own x
		childVal, _ := child.GetLocal("x")
		if childVal != 100 {
			t.Errorf("expected child to have 100, got %v", childVal)
		}

		// Parent should still have original x
		parentVal, _ := parent.GetLocal("x")
		if parentVal != 42 {
			t.Errorf("expected parent to have 42, got %v", parentVal)
		}
	})
}

// TestScopeDelete tests the Delete method.
func TestScopeDelete(t *testing.T) {
	t.Run("deletes existing variable", func(t *testing.T) {
		scope := NewScope(nil)
		scope.Set("x", 42)

		deleted := scope.Delete("x")
		if !deleted {
			t.Error("expected Delete to return true")
		}

		if scope.Has("x") {
			t.Error("expected variable to be deleted")
		}
	})

	t.Run("returns false for non-existent variable", func(t *testing.T) {
		scope := NewScope(nil)
		deleted := scope.Delete("nonexistent")
		if deleted {
			t.Error("expected Delete to return false")
		}
	})
}

// TestScopeHas tests Has and HasLocal methods.
func TestScopeHas(t *testing.T) {
	t.Run("Has returns true for existing variable", func(t *testing.T) {
		scope := NewScope(nil)
		scope.Set("x", 42)

		if !scope.Has("x") {
			t.Error("expected Has to return true")
		}
	})

	t.Run("Has returns false for non-existent variable", func(t *testing.T) {
		scope := NewScope(nil)
		if scope.Has("nonexistent") {
			t.Error("expected Has to return false")
		}
	})

	t.Run("Has searches parent scope", func(t *testing.T) {
		parent := NewScope(nil)
		parent.Set("x", 42)

		child := NewScope(parent)
		if !child.Has("x") {
			t.Error("expected Has to find variable in parent")
		}
	})

	t.Run("HasLocal does not search parent", func(t *testing.T) {
		parent := NewScope(nil)
		parent.Set("x", 42)

		child := NewScope(parent)
		if child.HasLocal("x") {
			t.Error("expected HasLocal to not find parent variable")
		}
	})
}

// TestScopeKeys tests Keys and AllKeys methods.
func TestScopeKeys(t *testing.T) {
	t.Run("Keys returns local variables only", func(t *testing.T) {
		parent := NewScope(nil)
		parent.Set("a", 1)

		child := NewScope(parent)
		child.SetLocal("b", 2)
		child.SetLocal("c", 3)

		keys := child.Keys()
		if len(keys) != 2 {
			t.Errorf("expected 2 keys, got %d", len(keys))
		}

		// Check that keys contains b and c
		keyMap := make(map[string]bool)
		for _, k := range keys {
			keyMap[k] = true
		}
		if !keyMap["b"] || !keyMap["c"] {
			t.Error("expected keys to contain 'b' and 'c'")
		}
	})

	t.Run("AllKeys returns all variables including parent", func(t *testing.T) {
		parent := NewScope(nil)
		parent.Set("a", 1)

		child := NewScope(parent)
		child.SetLocal("b", 2)

		keys := child.AllKeys()
		if len(keys) != 2 {
			t.Errorf("expected 2 keys, got %d", len(keys))
		}

		// Check that keys contains a and b
		keyMap := make(map[string]bool)
		for _, k := range keys {
			keyMap[k] = true
		}
		if !keyMap["a"] || !keyMap["b"] {
			t.Error("expected keys to contain 'a' and 'b'")
		}
	})
}

// TestScopeClear tests the Clear method.
func TestScopeClear(t *testing.T) {
	t.Run("clears all variables", func(t *testing.T) {
		scope := NewScope(nil)
		scope.Set("a", 1)
		scope.Set("b", 2)

		scope.Clear()

		if scope.Size() != 0 {
			t.Errorf("expected size 0 after clear, got %d", scope.Size())
		}
		if scope.Has("a") || scope.Has("b") {
			t.Error("expected variables to be cleared")
		}
	})
}

// TestScopeSize tests the Size method.
func TestScopeSize(t *testing.T) {
	t.Run("returns correct size", func(t *testing.T) {
		scope := NewScope(nil)
		if scope.Size() != 0 {
			t.Errorf("expected size 0, got %d", scope.Size())
		}

		scope.Set("a", 1)
		if scope.Size() != 1 {
			t.Errorf("expected size 1, got %d", scope.Size())
		}

		scope.Set("b", 2)
		if scope.Size() != 2 {
			t.Errorf("expected size 2, got %d", scope.Size())
		}

		scope.Delete("a")
		if scope.Size() != 1 {
			t.Errorf("expected size 1 after delete, got %d", scope.Size())
		}
	})
}

// TestNestedScopes tests nested scope behavior for function calls.
// Requirement 9.8: System supports nested function calls with independent local scopes.
func TestNestedScopes(t *testing.T) {
	t.Run("supports deeply nested scopes", func(t *testing.T) {
		// Simulate: global -> func1 -> func2 -> func3
		global := NewScope(nil)
		global.Set("x", 1)

		func1Scope := NewScope(global)
		func1Scope.SetLocal("x", 10)
		func1Scope.SetLocal("y", 20)

		func2Scope := NewScope(func1Scope)
		func2Scope.SetLocal("x", 100)
		func2Scope.SetLocal("z", 30)

		func3Scope := NewScope(func2Scope)
		func3Scope.SetLocal("w", 40)

		// func3 should see its own w, func2's z, func1's y, and func2's x (shadowed)
		w, _ := func3Scope.Get("w")
		if w != 40 {
			t.Errorf("expected w=40, got %v", w)
		}

		z, _ := func3Scope.Get("z")
		if z != 30 {
			t.Errorf("expected z=30, got %v", z)
		}

		y, _ := func3Scope.Get("y")
		if y != 20 {
			t.Errorf("expected y=20, got %v", y)
		}

		x, _ := func3Scope.Get("x")
		if x != 100 {
			t.Errorf("expected x=100 (from func2), got %v", x)
		}

		// Each scope should have independent local variables
		if func3Scope.HasLocal("z") {
			t.Error("func3 should not have z locally")
		}
		if func2Scope.HasLocal("y") {
			t.Error("func2 should not have y locally")
		}
	})

	t.Run("scope destruction does not affect parent", func(t *testing.T) {
		global := NewScope(nil)
		global.Set("x", 1)

		// Simulate function call
		funcScope := NewScope(global)
		funcScope.SetLocal("x", 10)
		funcScope.SetLocal("y", 20)

		// Verify local scope has its own values
		localX, _ := funcScope.Get("x")
		if localX != 10 {
			t.Errorf("expected local x=10, got %v", localX)
		}

		// Simulate function return (scope goes out of scope)
		funcScope = nil

		// Global should still have its value
		globalX, _ := global.Get("x")
		if globalX != 1 {
			t.Errorf("expected global x=1, got %v", globalX)
		}
	})

	t.Run("recursive function calls have independent scopes", func(t *testing.T) {
		// Simulate recursive calls: global -> recurse(3) -> recurse(2) -> recurse(1)
		global := NewScope(nil)
		global.Set("result", 0)

		// First call: n=3
		scope1 := NewScope(global)
		scope1.SetLocal("n", 3)

		// Second call: n=2
		scope2 := NewScope(scope1)
		scope2.SetLocal("n", 2)

		// Third call: n=1
		scope3 := NewScope(scope2)
		scope3.SetLocal("n", 1)

		// Each scope should have its own n
		n1, _ := scope1.GetLocal("n")
		n2, _ := scope2.GetLocal("n")
		n3, _ := scope3.GetLocal("n")

		if n1 != 3 {
			t.Errorf("expected scope1 n=3, got %v", n1)
		}
		if n2 != 2 {
			t.Errorf("expected scope2 n=2, got %v", n2)
		}
		if n3 != 1 {
			t.Errorf("expected scope3 n=1, got %v", n3)
		}

		// Get from innermost scope should return local value
		n, _ := scope3.Get("n")
		if n != 1 {
			t.Errorf("expected Get from scope3 to return 1, got %v", n)
		}
	})
}

// TestScopeVariableTypes tests storing different types of values.
func TestScopeVariableTypes(t *testing.T) {
	scope := NewScope(nil)

	t.Run("stores int", func(t *testing.T) {
		scope.Set("int", 42)
		val, _ := scope.Get("int")
		if val != 42 {
			t.Errorf("expected 42, got %v", val)
		}
	})

	t.Run("stores int64", func(t *testing.T) {
		scope.Set("int64", int64(9223372036854775807))
		val, _ := scope.Get("int64")
		if val != int64(9223372036854775807) {
			t.Errorf("expected max int64, got %v", val)
		}
	})

	t.Run("stores float64", func(t *testing.T) {
		scope.Set("float", 3.14159)
		val, _ := scope.Get("float")
		if val != 3.14159 {
			t.Errorf("expected 3.14159, got %v", val)
		}
	})

	t.Run("stores string", func(t *testing.T) {
		scope.Set("string", "hello")
		val, _ := scope.Get("string")
		if val != "hello" {
			t.Errorf("expected 'hello', got %v", val)
		}
	})

	t.Run("stores bool", func(t *testing.T) {
		scope.Set("bool", true)
		val, _ := scope.Get("bool")
		if val != true {
			t.Errorf("expected true, got %v", val)
		}
	})

	t.Run("stores slice", func(t *testing.T) {
		arr := []int{1, 2, 3}
		scope.Set("slice", arr)
		val, _ := scope.Get("slice")
		if valArr, ok := val.([]int); !ok || len(valArr) != 3 {
			t.Errorf("expected slice [1,2,3], got %v", val)
		}
	})

	t.Run("stores nil", func(t *testing.T) {
		scope.Set("nil", nil)
		val, ok := scope.Get("nil")
		if !ok {
			t.Error("expected variable to exist")
		}
		if val != nil {
			t.Errorf("expected nil, got %v", val)
		}
	})
}
