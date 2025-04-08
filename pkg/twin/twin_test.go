package twin

import (
	"testing"
	"time"
)

func TestNewDigitalTwin(t *testing.T) {
	id := "test-twin-1"
	twinType := "sensor"
	
	dt := NewDigitalTwin(id, twinType)
	
	if dt.ID != id {
		t.Errorf("Expected ID %s, got %s", id, dt.ID)
	}
	
	if dt.Type != twinType {
		t.Errorf("Expected Type %s, got %s", twinType, dt.Type)
	}
	
	if dt.Attributes == nil {
		t.Error("Attributes map should be initialized")
	}
	
	if dt.Features == nil {
		t.Error("Features map should be initialized")
	}
	
	if dt.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
	
	if dt.ModifiedAt.IsZero() {
		t.Error("ModifiedAt should be set")
	}
}

func TestDigitalTwinAttributes(t *testing.T) {
	dt := NewDigitalTwin("test-twin-2", "device")
	
	// Test setting and getting attributes
	dt.SetAttribute("manufacturer", "ACME Corp")
	dt.SetAttribute("model", "X-1000")
	dt.SetAttribute("year", 2025)
	
	// Test GetAttribute
	if val, exists := dt.GetAttribute("manufacturer"); !exists || val != "ACME Corp" {
		t.Errorf("Expected manufacturer to be 'ACME Corp', got %v", val)
	}
	
	if val, exists := dt.GetAttribute("model"); !exists || val != "X-1000" {
		t.Errorf("Expected model to be 'X-1000', got %v", val)
	}
	
	if val, exists := dt.GetAttribute("year"); !exists || val != 2025 {
		t.Errorf("Expected year to be 2025, got %v", val)
	}
	
	// Test GetAllAttributes
	attrs := dt.GetAllAttributes()
	if len(attrs) != 3 {
		t.Errorf("Expected 3 attributes, got %d", len(attrs))
	}
	
	// Test RemoveAttribute
	dt.RemoveAttribute("model")
	if _, exists := dt.GetAttribute("model"); exists {
		t.Error("Expected model attribute to be removed")
	}
	
	attrs = dt.GetAllAttributes()
	if len(attrs) != 2 {
		t.Errorf("Expected 2 attributes after removal, got %d", len(attrs))
	}
}

func TestDigitalTwinFeatures(t *testing.T) {
	dt := NewDigitalTwin("test-twin-3", "thermostat")
	
	// Create a feature
	feature := NewFeatureState()
	feature.SetProperty("temperature", 22.5)
	feature.SetProperty("unit", "celsius")
	feature.SetDesiredProperty("temperature", 23.0)
	feature.SetDefinition([]string{"org.example:thermostat:1.0.0"})
	
	// Test AddFeature
	err := dt.AddFeature("temperature", *feature)
	if err != nil {
		t.Errorf("Failed to add feature: %v", err)
	}
	
	// Test GetFeature
	if retrievedFeature, exists := dt.GetFeature("temperature"); !exists {
		t.Error("Expected feature to exist")
	} else {
		if val, exists := retrievedFeature.GetProperty("temperature"); !exists || val != 22.5 {
			t.Errorf("Expected temperature property to be 22.5, got %v", val)
		}
		
		if val, exists := retrievedFeature.GetDesiredProperty("temperature"); !exists || val != 23.0 {
			t.Errorf("Expected desired temperature property to be 23.0, got %v", val)
		}
		
		defs := retrievedFeature.GetDefinition()
		if len(defs) != 1 || defs[0] != "org.example:thermostat:1.0.0" {
			t.Errorf("Expected definition to be ['org.example:thermostat:1.0.0'], got %v", defs)
		}
	}
	
	// Test GetAllFeatures
	features := dt.GetAllFeatures()
	if len(features) != 1 {
		t.Errorf("Expected 1 feature, got %d", len(features))
	}
	
	// Test UpdateFeature
	updatedFeature := NewFeatureState()
	updatedFeature.SetProperty("temperature", 24.0)
	updatedFeature.SetProperty("humidity", 45)
	
	err = dt.UpdateFeature("temperature", *updatedFeature)
	if err != nil {
		t.Errorf("Failed to update feature: %v", err)
	}
	
	if retrievedFeature, exists := dt.GetFeature("temperature"); !exists {
		t.Error("Expected feature to exist after update")
	} else {
		if val, exists := retrievedFeature.GetProperty("temperature"); !exists || val != 24.0 {
			t.Errorf("Expected updated temperature property to be 24.0, got %v", val)
		}
		
		if val, exists := retrievedFeature.GetProperty("humidity"); !exists || val != 45 {
			t.Errorf("Expected humidity property to be 45, got %v", val)
		}
	}
	
	// Test error cases
	err = dt.AddFeature("temperature", *feature)
	if err != ErrFeatureAlreadyExists {
		t.Errorf("Expected ErrFeatureAlreadyExists, got %v", err)
	}
	
	err = dt.UpdateFeature("nonexistent", *feature)
	if err != ErrFeatureNotFound {
		t.Errorf("Expected ErrFeatureNotFound, got %v", err)
	}
	
	// Test RemoveFeature
	err = dt.RemoveFeature("temperature")
	if err != nil {
		t.Errorf("Failed to remove feature: %v", err)
	}
	
	if _, exists := dt.GetFeature("temperature"); exists {
		t.Error("Expected feature to be removed")
	}
	
	err = dt.RemoveFeature("nonexistent")
	if err != ErrFeatureNotFound {
		t.Errorf("Expected ErrFeatureNotFound when removing nonexistent feature, got %v", err)
	}
}

func TestDigitalTwinDefinition(t *testing.T) {
	dt := NewDigitalTwin("test-twin-4", "device")
	
	// Test setting and getting definition
	definition := "org.example:device:2.0.0"
	dt.SetDefinition(definition)
	
	if dt.GetDefinition() != definition {
		t.Errorf("Expected definition to be %s, got %s", definition, dt.GetDefinition())
	}
}

func TestDigitalTwinConcurrency(t *testing.T) {
	dt := NewDigitalTwin("test-twin-5", "concurrent-device")
	
	// Test concurrent attribute access
	done := make(chan bool)
	
	// Concurrent writers
	for i := 0; i < 10; i++ {
		go func(idx int) {
			key := fmt.Sprintf("attr-%d", idx)
			dt.SetAttribute(key, idx)
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
			key := fmt.Sprintf("attr-%d", idx)
			val, exists := dt.GetAttribute(key)
			if !exists || val != idx {
				t.Errorf("Expected attribute %s to be %d, got %v", key, idx, val)
			}
			done <- true
		}(i)
	}
	
	// Wait for all readers
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Verify all attributes are present
	attrs := dt.GetAllAttributes()
	if len(attrs) != 10 {
		t.Errorf("Expected 10 attributes, got %d", len(attrs))
	}
}
