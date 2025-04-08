package registry

import (
	"testing"

	"github.com/yourusername/go-digital-twin/pkg/twin"
)

func TestRegistryCreation(t *testing.T) {
	reg := NewRegistry()
	
	if reg.twins == nil {
		t.Error("Twins map should be initialized")
	}
	
	twins := reg.List()
	if len(twins) != 0 {
		t.Errorf("Expected empty registry, got %d twins", len(twins))
	}
}

func TestRegistryCRUD(t *testing.T) {
	reg := NewRegistry()
	
	// Create a test twin
	dt := twin.NewDigitalTwin("test-twin-1", "sensor")
	dt.SetAttribute("location", "living-room")
	
	// Test Create
	err := reg.Create(dt)
	if err != nil {
		t.Errorf("Failed to create twin: %v", err)
	}
	
	// Test Get
	retrievedTwin, err := reg.Get("test-twin-1")
	if err != nil {
		t.Errorf("Failed to get twin: %v", err)
	}
	
	if retrievedTwin.ID != "test-twin-1" {
		t.Errorf("Expected ID test-twin-1, got %s", retrievedTwin.ID)
	}
	
	if val, exists := retrievedTwin.GetAttribute("location"); !exists || val != "living-room" {
		t.Errorf("Expected location attribute to be 'living-room', got %v", val)
	}
	
	// Test List
	twins := reg.List()
	if len(twins) != 1 {
		t.Errorf("Expected 1 twin, got %d", len(twins))
	}
	
	// Test Update
	dt.SetAttribute("temperature", 22.5)
	err = reg.Update(dt)
	if err != nil {
		t.Errorf("Failed to update twin: %v", err)
	}
	
	retrievedTwin, err = reg.Get("test-twin-1")
	if err != nil {
		t.Errorf("Failed to get updated twin: %v", err)
	}
	
	if val, exists := retrievedTwin.GetAttribute("temperature"); !exists || val != 22.5 {
		t.Errorf("Expected temperature attribute to be 22.5, got %v", val)
	}
	
	// Test Delete
	err = reg.Delete("test-twin-1")
	if err != nil {
		t.Errorf("Failed to delete twin: %v", err)
	}
	
	_, err = reg.Get("test-twin-1")
	if err != ErrTwinNotFound {
		t.Errorf("Expected ErrTwinNotFound, got %v", err)
	}
	
	twins = reg.List()
	if len(twins) != 0 {
		t.Errorf("Expected empty registry after deletion, got %d twins", len(twins))
	}
}

func TestRegistryErrorCases(t *testing.T) {
	reg := NewRegistry()
	
	// Test Get with non-existent ID
	_, err := reg.Get("non-existent")
	if err != ErrTwinNotFound {
		t.Errorf("Expected ErrTwinNotFound, got %v", err)
	}
	
	// Test Update with non-existent ID
	dt := twin.NewDigitalTwin("non-existent", "sensor")
	err = reg.Update(dt)
	if err != ErrTwinNotFound {
		t.Errorf("Expected ErrTwinNotFound, got %v", err)
	}
	
	// Test Delete with non-existent ID
	err = reg.Delete("non-existent")
	if err != ErrTwinNotFound {
		t.Errorf("Expected ErrTwinNotFound, got %v", err)
	}
	
	// Test Create with duplicate ID
	dt1 := twin.NewDigitalTwin("duplicate", "sensor")
	err = reg.Create(dt1)
	if err != nil {
		t.Errorf("Failed to create first twin: %v", err)
	}
	
	dt2 := twin.NewDigitalTwin("duplicate", "actuator")
	err = reg.Create(dt2)
	if err != ErrTwinAlreadyExists {
		t.Errorf("Expected ErrTwinAlreadyExists, got %v", err)
	}
}

func TestRegistryFind(t *testing.T) {
	reg := NewRegistry()
	
	// Create test twins
	dt1 := twin.NewDigitalTwin("twin-1", "sensor")
	dt1.SetAttribute("location", "living-room")
	dt1.SetAttribute("manufacturer", "ACME")
	
	feature1 := twin.NewFeatureState()
	feature1.SetProperty("temperature", 22.5)
	dt1.AddFeature("temperature", *feature1)
	
	dt2 := twin.NewDigitalTwin("twin-2", "sensor")
	dt2.SetAttribute("location", "bedroom")
	dt2.SetAttribute("manufacturer", "ACME")
	
	feature2 := twin.NewFeatureState()
	feature2.SetProperty("temperature", 20.0)
	dt2.AddFeature("temperature", *feature2)
	
	dt3 := twin.NewDigitalTwin("twin-3", "actuator")
	dt3.SetAttribute("location", "kitchen")
	dt3.SetAttribute("manufacturer", "XYZ")
	
	feature3 := twin.NewFeatureState()
	feature3.SetProperty("state", "on")
	dt3.AddFeature("switch", *feature3)
	
	// Add twins to registry
	reg.Create(dt1)
	reg.Create(dt2)
	reg.Create(dt3)
	
	// Test FindByAttribute
	acmeTwins := reg.FindByAttribute("manufacturer", "ACME")
	if len(acmeTwins) != 2 {
		t.Errorf("Expected 2 twins with ACME manufacturer, got %d", len(acmeTwins))
	}
	
	kitchenTwins := reg.FindByAttribute("location", "kitchen")
	if len(kitchenTwins) != 1 {
		t.Errorf("Expected 1 twin in kitchen, got %d", len(kitchenTwins))
	}
	
	nonExistentTwins := reg.FindByAttribute("nonexistent", "value")
	if len(nonExistentTwins) != 0 {
		t.Errorf("Expected 0 twins with nonexistent attribute, got %d", len(nonExistentTwins))
	}
	
	// Test FindByFeature
	temperatureTwins := reg.FindByFeature("temperature")
	if len(temperatureTwins) != 2 {
		t.Errorf("Expected 2 twins with temperature feature, got %d", len(temperatureTwins))
	}
	
	switchTwins := reg.FindByFeature("switch")
	if len(switchTwins) != 1 {
		t.Errorf("Expected 1 twin with switch feature, got %d", len(switchTwins))
	}
	
	nonExistentFeatureTwins := reg.FindByFeature("nonexistent")
	if len(nonExistentFeatureTwins) != 0 {
		t.Errorf("Expected 0 twins with nonexistent feature, got %d", len(nonExistentFeatureTwins))
	}
}

func TestRegistryConcurrency(t *testing.T) {
	reg := NewRegistry()
	done := make(chan bool)
	
	// Concurrent creation
	for i := 0; i < 10; i++ {
		go func(idx int) {
			id := fmt.Sprintf("twin-%d", idx)
			dt := twin.NewDigitalTwin(id, "sensor")
			err := reg.Create(dt)
			if err != nil {
				t.Errorf("Failed to create twin %s: %v", id, err)
			}
			done <- true
		}(i)
	}
	
	// Wait for all creations
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Verify all twins were created
	twins := reg.List()
	if len(twins) != 10 {
		t.Errorf("Expected 10 twins, got %d", len(twins))
	}
	
	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func(idx int) {
			id := fmt.Sprintf("twin-%d", idx)
			_, err := reg.Get(id)
			if err != nil {
				t.Errorf("Failed to get twin %s: %v", id, err)
			}
			done <- true
		}(i)
	}
	
	// Wait for all reads
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Concurrent updates
	for i := 0; i < 10; i++ {
		go func(idx int) {
			id := fmt.Sprintf("twin-%d", idx)
			dt, err := reg.Get(id)
			if err != nil {
				t.Errorf("Failed to get twin %s for update: %v", id, err)
				done <- true
				return
			}
			
			dt.SetAttribute("value", idx)
			err = reg.Update(dt)
			if err != nil {
				t.Errorf("Failed to update twin %s: %v", id, err)
			}
			done <- true
		}(i)
	}
	
	// Wait for all updates
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Verify updates
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("twin-%d", i)
		dt, err := reg.Get(id)
		if err != nil {
			t.Errorf("Failed to get twin %s after update: %v", id, err)
			continue
		}
		
		val, exists := dt.GetAttribute("value")
		if !exists || val != i {
			t.Errorf("Expected twin %s to have value %d, got %v", id, i, val)
		}
	}
}
