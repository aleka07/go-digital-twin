package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aleka07/go-digital-twin/pkg/messaging_sim"
	"github.com/aleka07/go-digital-twin/pkg/registry"
	"github.com/aleka07/go-digital-twin/pkg/twin"
)

func setupTestServer() *Server {
	reg := registry.NewRegistry()
	pubsub := messaging_sim.NewPubSub()
	return NewServer(reg, pubsub)
}

func TestCreateTwin(t *testing.T) {
	server := setupTestServer()

	// Create a test twin
	twinData := map[string]interface{}{
		"id":         "test-twin-1",
		"type":       "sensor",
		"definition": "org.example:sensor:1.0.0",
		"attributes": map[string]interface{}{
			"location": "living-room",
			"model":    "X-1000",
		},
	}

	jsonData, _ := json.Marshal(twinData)
	req := httptest.NewRequest("POST", "/twins", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	server.CreateTwin(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	// Decode response
	var createdTwin map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&createdTwin)

	// Verify twin was created correctly
	if createdTwin["id"] != "test-twin-1" {
		t.Errorf("Expected twin ID test-twin-1, got %v", createdTwin["id"])
	}

	if createdTwin["type"] != "sensor" {
		t.Errorf("Expected twin type sensor, got %v", createdTwin["type"])
	}

	// Test creating a twin with missing required fields
	invalidTwin := map[string]interface{}{
		"attributes": map[string]interface{}{
			"location": "bedroom",
		},
	}

	jsonData, _ = json.Marshal(invalidTwin)
	req = httptest.NewRequest("POST", "/twins", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	w = httptest.NewRecorder()
	server.CreateTwin(w, req)

	resp = w.Result()
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestGetTwin(t *testing.T) {
	server := setupTestServer()

	// Create a test twin
	dt := twin.NewDigitalTwin("test-twin-2", "sensor")
	dt.SetAttribute("location", "kitchen")
	server.Registry.Create(dt)

	// Test getting the twin
	req := httptest.NewRequest("GET", "/twins/test-twin-2", nil)
	req.Header.Set("Content-Type", "application/json")
	
	// Set URL parameter
	req = req.WithContext(setURLParam(req.Context(), "twinID", "test-twin-2"))
	
	w := httptest.NewRecorder()
	server.GetTwin(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Decode response
	var retrievedTwin map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&retrievedTwin)

	// Verify twin was retrieved correctly
	if retrievedTwin["id"] != "test-twin-2" {
		t.Errorf("Expected twin ID test-twin-2, got %v", retrievedTwin["id"])
	}

	// Test getting a non-existent twin
	req = httptest.NewRequest("GET", "/twins/non-existent", nil)
	req = req.WithContext(setURLParam(req.Context(), "twinID", "non-existent"))
	
	w = httptest.NewRecorder()
	server.GetTwin(w, req)

	resp = w.Result()
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestUpdateTwin(t *testing.T) {
	server := setupTestServer()

	// Create a test twin
	dt := twin.NewDigitalTwin("test-twin-3", "sensor")
	dt.SetAttribute("location", "bedroom")
	server.Registry.Create(dt)

	// Update the twin
	updateData := map[string]interface{}{
		"type": "advanced-sensor",
		"attributes": map[string]interface{}{
			"location": "master-bedroom",
			"status":   "active",
		},
	}

	jsonData, _ := json.Marshal(updateData)
	req := httptest.NewRequest("PUT", "/twins/test-twin-3", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	// Set URL parameter
	req = req.WithContext(setURLParam(req.Context(), "twinID", "test-twin-3"))
	
	w := httptest.NewRecorder()
	server.UpdateTwin(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Verify twin was updated
	updatedTwin, _ := server.Registry.Get("test-twin-3")
	
	if updatedTwin.Type != "advanced-sensor" {
		t.Errorf("Expected twin type advanced-sensor, got %s", updatedTwin.Type)
	}
	
	if val, exists := updatedTwin.GetAttribute("location"); !exists || val != "master-bedroom" {
		t.Errorf("Expected location attribute to be master-bedroom, got %v", val)
	}
	
	if val, exists := updatedTwin.GetAttribute("status"); !exists || val != "active" {
		t.Errorf("Expected status attribute to be active, got %v", val)
	}
}

func TestDeleteTwin(t *testing.T) {
	server := setupTestServer()

	// Create a test twin
	dt := twin.NewDigitalTwin("test-twin-4", "sensor")
	server.Registry.Create(dt)

	// Delete the twin
	req := httptest.NewRequest("DELETE", "/twins/test-twin-4", nil)
	
	// Set URL parameter
	req = req.WithContext(setURLParam(req.Context(), "twinID", "test-twin-4"))
	
	w := httptest.NewRecorder()
	server.DeleteTwin(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Verify twin was deleted
	_, err := server.Registry.Get("test-twin-4")
	if err != registry.ErrTwinNotFound {
		t.Errorf("Expected ErrTwinNotFound, got %v", err)
	}
}

func TestFeatureManagement(t *testing.T) {
	server := setupTestServer()

	// Create a test twin with features
	dt := twin.NewDigitalTwin("test-twin-5", "device")
	
	tempFeature := twin.NewFeatureState()
	tempFeature.SetProperty("value", 22.5)
	tempFeature.SetProperty("unit", "celsius")
	
	dt.AddFeature("temperature", *tempFeature)
	server.Registry.Create(dt)

	// Test getting features
	req := httptest.NewRequest("GET", "/twins/test-twin-5/features", nil)
	req = req.WithContext(setURLParam(req.Context(), "twinID", "test-twin-5"))
	
	w := httptest.NewRecorder()
	server.GetFeatures(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Test getting a specific feature
	req = httptest.NewRequest("GET", "/twins/test-twin-5/features/temperature", nil)
	req = req.WithContext(setURLParam(req.Context(), "twinID", "test-twin-5"))
	req = req.WithContext(setURLParam(req.Context(), "featureID", "temperature"))
	
	w = httptest.NewRecorder()
	server.GetFeature(w, req)

	resp = w.Result()
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Decode response
	var feature map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&feature)

	// Verify feature properties
	properties, ok := feature["Properties"].(map[string]interface{})
	if !ok {
		t.Error("Expected Properties to be a map")
	} else {
		if properties["value"] != 22.5 {
			t.Errorf("Expected temperature value to be 22.5, got %v", properties["value"])
		}
	}

	// Test updating a feature
	updateData := map[string]interface{}{
		"properties": map[string]interface{}{
			"value": 23.0,
			"status": "normal",
		},
	}

	jsonData, _ := json.Marshal(updateData)
	req = httptest.NewRequest("PUT", "/twins/test-twin-5/features/temperature", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(setURLParam(req.Context(), "twinID", "test-twin-5"))
	req = req.WithContext(setURLParam(req.Context(), "featureID", "temperature"))
	
	w = httptest.NewRecorder()
	server.UpdateFeature(w, req)

	resp = w.Result()
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Verify feature was updated
	updatedTwin, _ := server.Registry.Get("test-twin-5")
	updatedFeature, _ := updatedTwin.GetFeature("temperature")
	
	if val, exists := updatedFeature.GetProperty("value"); !exists || val != 23.0 {
		t.Errorf("Expected value property to be 23.0, got %v", val)
	}
	
	if val, exists := updatedFeature.GetProperty("status"); !exists || val != "normal" {
		t.Errorf("Expected status property to be normal, got %v", val)
	}
}

func TestPropertyManagement(t *testing.T) {
	server := setupTestServer()

	// Create a test twin with features and properties
	dt := twin.NewDigitalTwin("test-twin-6", "device")
	
	lightFeature := twin.NewFeatureState()
	lightFeature.SetProperty("state", "on")
	lightFeature.SetProperty("brightness", 80)
	lightFeature.SetProperty("color", "white")
	
	dt.AddFeature("light", *lightFeature)
	server.Registry.Create(dt)

	// Test getting properties
	req := httptest.NewRequest("GET", "/twins/test-twin-6/features/light/properties", nil)
	req = req.WithContext(setURLParam(req.Context(), "twinID", "test-twin-6"))
	req = req.WithContext(setURLParam(req.Context(), "featureID", "light"))
	
	w := httptest.NewRecorder()
	server.GetProperties(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Test getting a specific property
	req = httptest.NewRequest("GET", "/twins/test-twin-6/features/light/properties/brightness", nil)
	req = req.WithContext(setURLParam(req.Context(), "twinID", "test-twin-6"))
	req = req.WithContext(setURLParam(req.Context(), "featureID", "light"))
	req = req.WithContext(setURLParam(req.Context(), "propKey", "brightness"))
	
	w = httptest.NewRecorder()
	server.GetProperty(w, req)

	resp = w.Result()
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Decode response
	var brightness float64
	json.NewDecoder(resp.Body).Decode(&brightness)

	// Verify property value
	if brightness != 80 {
		t.Errorf("Expected brightness to be 80, got %v", brightness)
	}

	// Test updating a property
	jsonData, _ := json.Marshal(90)
	req = httptest.NewRequest("PUT", "/twins/test-twin-6/features/light/properties/brightness", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(setURLParam(req.Context(), "twinID", "test-twin-6"))
	req = req.WithContext(setURLParam(req.Context(), "featureID", "light"))
	req = req.WithContext(setURLParam(req.Context(), "propKey", "brightness"))
	
	w = httptest.NewRecorder()
	server.UpdateProperty(w, req)

	resp = w.Result()
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Verify property was updated
	updatedTwin, _ := server.Registry.Get("test-twin-6")
	updatedFeature, _ := updatedTwin.GetFeature("light")
	
	if val, exists := updatedFeature.GetProperty("brightness"); !exists || val != 90.0 {
		t.Errorf("Expected brightness property to be 90, got %v", val)
	}

	// Test deleting a property
	req = httptest.NewRequest("DELETE", "/twins/test-twin-6/features/light/properties/color", nil)
	req = req.WithContext(setURLParam(req.Context(), "twinID", "test-twin-6"))
	req = req.WithContext(setURLParam(req.Context(), "featureID", "light"))
	req = req.WithContext(setURLParam(req.Context(), "propKey", "color"))
	
	w = httptest.NewRecorder()
	server.DeleteProperty(w, req)

	resp = w.Result()
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Verify property was deleted
	updatedTwin, _ = server.Registry.Get("test-twin-6")
	updatedFeature, _ = updatedTwin.GetFeature("light")
	
	if _, exists := updatedFeature.GetProperty("color"); exists {
		t.Error("Expected color property to be deleted")
	}
}

// Helper function to set URL parameters in the request context
// In a real application, this would be handled by the router
func setURLParam(ctx context.Context, key, value string) context.Context {
	return context.WithValue(ctx, contextKey(key), value)
}

type contextKey string
