package twin

import (
	"fmt"
	"testing"
)

func TestFeatureStateCreation(t *testing.T) {
	fs := NewFeatureState()
	
	if fs.Properties == nil {
		t.Error("Properties map should be initialized")
	}
	
	if fs.DesiredProps == nil {
		t.Error("DesiredProps map should be initialized")
	}
	
	if fs.Definition == nil {
		t.Error("Definition slice should be initialized")
	}
	
	if fs.LastModified.IsZero() {
		t.Error("LastModified should be set")
	}
}

func TestFeatureStateProperties(t *testing.T) {
	fs := NewFeatureState()
	
	// Test setting and getting properties
	fs.SetProperty("power", true)
	fs.SetProperty("brightness", 75)
	fs.SetProperty("color", "blue")
	
	// Test GetProperty
	if val, exists := fs.GetProperty("power"); !exists || val != true {
		t.Errorf("Expected power to be true, got %v", val)
	}
	
	if val, exists := fs.GetProperty("brightness"); !exists || val != 75 {
		t.Errorf("Expected brightness to be 75, got %v", val)
	}
	
	if val, exists := fs.GetProperty("color"); !exists || val != "blue" {
		t.Errorf("Expected color to be 'blue', got %v", val)
	}
	
	// Test GetAllProperties
	props := fs.GetAllProperties()
	if len(props) != 3 {
		t.Errorf("Expected 3 properties, got %d", len(props))
	}
	
	// Test RemoveProperty
	fs.RemoveProperty("brightness")
	if _, exists := fs.GetProperty("brightness"); exists {
		t.Error("Expected brightness property to be removed")
	}
	
	props = fs.GetAllProperties()
	if len(props) != 2 {
		t.Errorf("Expected 2 properties after removal, got %d", len(props))
	}
}

func TestFeatureStateDesiredProperties(t *testing.T) {
	fs := NewFeatureState()
	
	// Test setting and getting desired properties
	fs.SetDesiredProperty("power", false)
	fs.SetDesiredProperty("brightness", 100)
	
	// Test GetDesiredProperty
	if val, exists := fs.GetDesiredProperty("power"); !exists || val != false {
		t.Errorf("Expected desired power to be false, got %v", val)
	}
	
	if val, exists := fs.GetDesiredProperty("brightness"); !exists || val != 100 {
		t.Errorf("Expected desired brightness to be 100, got %v", val)
	}
	
	// Test GetAllDesiredProperties
	desiredProps := fs.GetAllDesiredProperties()
	if len(desiredProps) != 2 {
		t.Errorf("Expected 2 desired properties, got %d", len(desiredProps))
	}
	
	// Test RemoveDesiredProperty
	fs.RemoveDesiredProperty("brightness")
	if _, exists := fs.GetDesiredProperty("brightness"); exists {
		t.Error("Expected desired brightness property to be removed")
	}
	
	desiredProps = fs.GetAllDesiredProperties()
	if len(desiredProps) != 1 {
		t.Errorf("Expected 1 desired property after removal, got %d", len(desiredProps))
	}
}

func TestFeatureStateDefinition(t *testing.T) {
	fs := NewFeatureState()
	
	// Test setting and getting definition
	definitions := []string{"org.example:light:1.0.0", "org.example:dimmable:1.0.0"}
	fs.SetDefinition(definitions)
	
	// Test GetDefinition
	defs := fs.GetDefinition()
	if len(defs) != 2 {
		t.Errorf("Expected 2 definitions, got %d", len(defs))
	}
	
	if defs[0] != "org.example:light:1.0.0" || defs[1] != "org.example:dimmable:1.0.0" {
		t.Errorf("Expected definitions to be ['org.example:light:1.0.0', 'org.example:dimmable:1.0.0'], got %v", defs)
	}
	
	// Test that definition is a copy, not a reference
	definitions[0] = "modified"
	defs = fs.GetDefinition()
	if defs[0] == "modified" {
		t.Error("Definition should be a copy, not a reference")
	}
}

func TestFeatureStateConcurrency(t *testing.T) {
	fs := NewFeatureState()
	
	// Test concurrent property access
	done := make(chan bool)
	
	// Concurrent writers
	for i := 0; i < 10; i++ {
		go func(idx int) {
			key := fmt.Sprintf("prop-%d", idx)
			fs.SetProperty(key, idx)
			done <- true
		}(i)
	}
	
	// Wait for all writers
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Concurrent readers
	for i := 0; i < 10; i++ {
		go func(idx int) {
			key := fmt.Sprintf("prop-%d", idx)
			val, exists := fs.GetProperty(key)
			if !exists || val != idx {
				t.Errorf("Expected property %s to be %d, got %v", key, idx, val)
			}
			done <- true
		}(i)
	}
	
	// Wait for all readers
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Verify all properties are present
	props := fs.GetAllProperties()
	if len(props) != 10 {
		t.Errorf("Expected 10 properties, got %d", len(props))
	}
}
