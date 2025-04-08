package api

import (
	"encoding/json"
	"net/http"

	"github.com/aleka07/go-digital-twin/pkg/registry"
	"github.com/aleka07/go-digital-twin/pkg/twin"
	"github.com/go-chi/chi/v5"
)

// Twin management handlers

// CreateTwin handles POST /twins
func (s *Server) CreateTwin(w http.ResponseWriter, r *http.Request) {
	s.wg.Add(1)
	defer s.wg.Done()

	var req struct {
		ID         string                 `json:"id"`
		Type       string                 `json:"type"`
		Definition string                 `json:"definition,omitempty"`
		Attributes map[string]interface{} `json:"attributes,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Validate required fields
	if req.ID == "" || req.Type == "" {
		respondError(w, http.StatusBadRequest, "ID and Type are required")
		return
	}

	// Create the digital twin
	dt := twin.NewDigitalTwin(req.ID, req.Type)

	// Set optional fields
	if req.Definition != "" {
		dt.SetDefinition(req.Definition)
	}

	for k, v := range req.Attributes {
		dt.SetAttribute(k, v)
	}

	// Add to registry
	if err := s.Registry.Create(dt); err != nil {
		if err == registry.ErrTwinAlreadyExists {
			respondError(w, http.StatusConflict, "Digital twin already exists")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to create digital twin: "+err.Error())
		}
		return
	}

	// Publish event
	s.PubSub.Publish("twin.created", map[string]string{"id": dt.ID})

	// Return the created twin
	respondJSON(w, http.StatusCreated, dt)
}

// GetTwin handles GET /twins/{twinID}
func (s *Server) GetTwin(w http.ResponseWriter, r *http.Request) {
	s.wg.Add(1)
	defer s.wg.Done()

	twinID := chi.URLParam(r, "twinID")
	if twinID == "" {
		respondError(w, http.StatusBadRequest, "Twin ID is required")
		return
	}

	dt, err := s.Registry.Get(twinID)
	if err != nil {
		if err == registry.ErrTwinNotFound {
			respondError(w, http.StatusNotFound, "Digital twin not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to get digital twin: "+err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, dt)
}

// UpdateTwin handles PUT /twins/{twinID}
func (s *Server) UpdateTwin(w http.ResponseWriter, r *http.Request) {
	s.wg.Add(1)
	defer s.wg.Done()

	twinID := chi.URLParam(r, "twinID")
	if twinID == "" {
		respondError(w, http.StatusBadRequest, "Twin ID is required")
		return
	}

	// Get existing twin
	dt, err := s.Registry.Get(twinID)
	if err != nil {
		if err == registry.ErrTwinNotFound {
			respondError(w, http.StatusNotFound, "Digital twin not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to get digital twin: "+err.Error())
		}
		return
	}

	// Parse update request
	var req struct {
		Type       string                 `json:"type,omitempty"`
		Definition string                 `json:"definition,omitempty"`
		Attributes map[string]interface{} `json:"attributes,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Update fields
	if req.Type != "" {
		dt.Type = req.Type
	}

	if req.Definition != "" {
		dt.SetDefinition(req.Definition)
	}

	if req.Attributes != nil {
		for k, v := range req.Attributes {
			dt.SetAttribute(k, v)
		}
	}

	// Update in registry
	if err := s.Registry.Update(dt); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update digital twin: "+err.Error())
		return
	}

	// Publish event
	s.PubSub.Publish("twin.updated", map[string]string{"id": dt.ID})

	respondJSON(w, http.StatusOK, dt)
}

// DeleteTwin handles DELETE /twins/{twinID}
func (s *Server) DeleteTwin(w http.ResponseWriter, r *http.Request) {
	s.wg.Add(1)
	defer s.wg.Done()

	twinID := chi.URLParam(r, "twinID")
	if twinID == "" {
		respondError(w, http.StatusBadRequest, "Twin ID is required")
		return
	}

	if err := s.Registry.Delete(twinID); err != nil {
		if err == registry.ErrTwinNotFound {
			respondError(w, http.StatusNotFound, "Digital twin not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to delete digital twin: "+err.Error())
		}
		return
	}

	// Publish event
	s.PubSub.Publish("twin.deleted", map[string]string{"id": twinID})

	respondJSON(w, http.StatusOK, map[string]string{"message": "Digital twin deleted"})
}

// ListTwins handles GET /twins
func (s *Server) ListTwins(w http.ResponseWriter, r *http.Request) {
	s.wg.Add(1)
	defer s.wg.Done()

	twins := s.Registry.List()
	respondJSON(w, http.StatusOK, twins)
}

// Feature management handlers

// GetFeatures handles GET /twins/{twinID}/features
func (s *Server) GetFeatures(w http.ResponseWriter, r *http.Request) {
	s.wg.Add(1)
	defer s.wg.Done()

	twinID := chi.URLParam(r, "twinID")
	if twinID == "" {
		respondError(w, http.StatusBadRequest, "Twin ID is required")
		return
	}

	dt, err := s.Registry.Get(twinID)
	if err != nil {
		if err == registry.ErrTwinNotFound {
			respondError(w, http.StatusNotFound, "Digital twin not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to get digital twin: "+err.Error())
		}
		return
	}

	features := dt.GetAllFeatures()
	respondJSON(w, http.StatusOK, features)
}

// GetFeature handles GET /twins/{twinID}/features/{featureID}
func (s *Server) GetFeature(w http.ResponseWriter, r *http.Request) {
	s.wg.Add(1)
	defer s.wg.Done()

	twinID := chi.URLParam(r, "twinID")
	featureID := chi.URLParam(r, "featureID")

	if twinID == "" || featureID == "" {
		respondError(w, http.StatusBadRequest, "Twin ID and Feature ID are required")
		return
	}

	dt, err := s.Registry.Get(twinID)
	if err != nil {
		if err == registry.ErrTwinNotFound {
			respondError(w, http.StatusNotFound, "Digital twin not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to get digital twin: "+err.Error())
		}
		return
	}

	feature, exists := dt.GetFeature(featureID)
	if !exists {
		respondError(w, http.StatusNotFound, "Feature not found")
		return
	}

	respondJSON(w, http.StatusOK, &feature)
}

// UpdateFeature handles PUT /twins/{twinID}/features/{featureID}
func (s *Server) UpdateFeature(w http.ResponseWriter, r *http.Request) {
	s.wg.Add(1)
	defer s.wg.Done()

	twinID := chi.URLParam(r, "twinID")
	featureID := chi.URLParam(r, "featureID")

	if twinID == "" || featureID == "" {
		respondError(w, http.StatusBadRequest, "Twin ID and Feature ID are required")
		return
	}

	dt, err := s.Registry.Get(twinID)
	if err != nil {
		if err == registry.ErrTwinNotFound {
			respondError(w, http.StatusNotFound, "Digital twin not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to get digital twin: "+err.Error())
		}
		return
	}

	var req struct {
		Properties   map[string]interface{} `json:"properties,omitempty"`
		DesiredProps map[string]interface{} `json:"desiredProperties,omitempty"`
		Definition   []string               `json:"definition,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Check if feature exists
	feature, exists := dt.GetFeature(featureID)

	// If feature doesn't exist, create a new one
	if !exists {
		feature = *twin.NewFeatureState()
	}

	// Update feature fields
	if req.Properties != nil {
		for k, v := range req.Properties {
			feature.SetProperty(k, v)
		}
	}

	if req.DesiredProps != nil {
		for k, v := range req.DesiredProps {
			feature.SetDesiredProperty(k, v)
		}
	}

	if req.Definition != nil {
		feature.SetDefinition(req.Definition)
	}

	// Update or add the feature
	var updateErr error
	if exists {
		updateErr = dt.UpdateFeature(featureID, feature)
	} else {
		updateErr = dt.AddFeature(featureID, feature)
	}

	if updateErr != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update feature: "+updateErr.Error())
		return
	}

	// Update the twin in the registry
	if err := s.Registry.Update(dt); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update digital twin: "+err.Error())
		return
	}

	// Publish event
	s.PubSub.Publish("feature.updated", map[string]string{
		"twinId":    twinID,
		"featureId": featureID,
	})

	respondJSON(w, http.StatusOK, feature)
}

// DeleteFeature handles DELETE /twins/{twinID}/features/{featureID}
func (s *Server) DeleteFeature(w http.ResponseWriter, r *http.Request) {
	s.wg.Add(1)
	defer s.wg.Done()

	twinID := chi.URLParam(r, "twinID")
	featureID := chi.URLParam(r, "featureID")

	if twinID == "" || featureID == "" {
		respondError(w, http.StatusBadRequest, "Twin ID and Feature ID are required")
		return
	}

	dt, err := s.Registry.Get(twinID)
	if err != nil {
		if err == registry.ErrTwinNotFound {
			respondError(w, http.StatusNotFound, "Digital twin not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to get digital twin: "+err.Error())
		}
		return
	}

	if err := dt.RemoveFeature(featureID); err != nil {
		if err == twin.ErrFeatureNotFound {
			respondError(w, http.StatusNotFound, "Feature not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to delete feature: "+err.Error())
		}
		return
	}

	// Update the twin in the registry
	if err := s.Registry.Update(dt); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update digital twin: "+err.Error())
		return
	}

	// Publish event
	s.PubSub.Publish("feature.deleted", map[string]string{
		"twinId":    twinID,
		"featureId": featureID,
	})

	respondJSON(w, http.StatusOK, map[string]string{"message": "Feature deleted"})
}

// Property management handlers

// GetProperties handles GET /twins/{twinID}/features/{featureID}/properties
func (s *Server) GetProperties(w http.ResponseWriter, r *http.Request) {
	s.wg.Add(1)
	defer s.wg.Done()

	twinID := chi.URLParam(r, "twinID")
	featureID := chi.URLParam(r, "featureID")

	if twinID == "" || featureID == "" {
		respondError(w, http.StatusBadRequest, "Twin ID and Feature ID are required")
		return
	}

	dt, err := s.Registry.Get(twinID)
	if err != nil {
		if err == registry.ErrTwinNotFound {
			respondError(w, http.StatusNotFound, "Digital twin not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to get digital twin: "+err.Error())
		}
		return
	}

	feature, exists := dt.GetFeature(featureID)
	if !exists {
		respondError(w, http.StatusNotFound, "Feature not found")
		return
	}

	properties := feature.GetAllProperties()
	respondJSON(w, http.StatusOK, properties)
}

// UpdateProperties handles PUT /twins/{twinID}/features/{featureID}/properties
func (s *Server) UpdateProperties(w http.ResponseWriter, r *http.Request) {
	s.wg.Add(1)
	defer s.wg.Done()

	twinID := chi.URLParam(r, "twinID")
	featureID := chi.URLParam(r, "featureID")

	if twinID == "" || featureID == "" {
		respondError(w, http.StatusBadRequest, "Twin ID and Feature ID are required")
		return
	}

	dt, err := s.Registry.Get(twinID)
	if err != nil {
		if err == registry.ErrTwinNotFound {
			respondError(w, http.StatusNotFound, "Digital twin not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to get digital twin: "+err.Error())
		}
		return
	}

	feature, exists := dt.GetFeature(featureID)
	if !exists {
		respondError(w, http.StatusNotFound, "Feature not found")
		return
	}

	var properties map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&properties); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Update properties
	for k, v := range properties {
		feature.SetProperty(k, v)
	}

	// Update the feature
	if err := dt.UpdateFeature(featureID, feature); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update feature: "+err.Error())
		return
	}

	// Update the twin in the registry
	if err := s.Registry.Update(dt); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update digital twin: "+err.Error())
		return
	}

	// Publish event
	s.PubSub.Publish("properties.updated", map[string]string{
		"twinId":    twinID,
		"featureId": featureID,
	})

	respondJSON(w, http.StatusOK, feature.GetAllProperties())
}

// GetProperty handles GET /twins/{twinID}/features/{featureID}/properties/{propKey}
func (s *Server) GetProperty(w http.ResponseWriter, r *http.Request) {
	s.wg.Add(1)
	defer s.wg.Done()

	twinID := chi.URLParam(r, "twinID")
	featureID := chi.URLParam(r, "featureID")
	propKey := chi.URLParam(r, "propKey")

	if twinID == "" || featureID == "" || propKey == "" {
		respondError(w, http.StatusBadRequest, "Twin ID, Feature ID, and Property Key are required")
		return
	}

	dt, err := s.Registry.Get(twinID)
	if err != nil {
		if err == registry.ErrTwinNotFound {
			respondError(w, http.StatusNotFound, "Digital twin not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to get digital twin: "+err.Error())
		}
		return
	}

	feature, exists := dt.GetFeature(featureID)
	if !exists {
		respondError(w, http.StatusNotFound, "Feature not found")
		return
	}

	propValue, exists := feature.GetProperty(propKey)
	if !exists {
		respondError(w, http.StatusNotFound, "Property not found")
		return
	}

	respondJSON(w, http.StatusOK, propValue)
}

// UpdateProperty handles PUT /twins/{twinID}/features/{featureID}/properties/{propKey}
func (s *Server) UpdateProperty(w http.ResponseWriter, r *http.Request) {
	s.wg.Add(1)
	defer s.wg.Done()

	twinID := chi.URLParam(r, "twinID")
	featureID := chi.URLParam(r, "featureID")
	propKey := chi.URLParam(r, "propKey")

	if twinID == "" || featureID == "" || propKey == "" {
		respondError(w, http.StatusBadRequest, "Twin ID, Feature ID, and Property Key are required")
		return
	}

	dt, err := s.Registry.Get(twinID)
	if err != nil {
		if err == registry.ErrTwinNotFound {
			respondError(w, http.StatusNotFound, "Digital twin not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to get digital twin: "+err.Error())
		}
		return
	}

	feature, exists := dt.GetFeature(featureID)
	if !exists {
		respondError(w, http.StatusNotFound, "Feature not found")
		return
	}

	var propValue interface{}
	if err := json.NewDecoder(r.Body).Decode(&propValue); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Update property
	feature.SetProperty(propKey, propValue)

	// Update the feature
	if err := dt.UpdateFeature(featureID, feature); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update feature: "+err.Error())
		return
	}

	// Update the twin in the registry
	if err := s.Registry.Update(dt); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update digital twin: "+err.Error())
		return
	}

	// Publish event
	s.PubSub.Publish("property.updated", map[string]interface{}{
		"twinId":      twinID,
		"featureId":   featureID,
		"propertyKey": propKey,
		"value":       propValue,
	})

	respondJSON(w, http.StatusOK, propValue)
}

// DeleteProperty handles DELETE /twins/{twinID}/features/{featureID}/properties/{propKey}
func (s *Server) DeleteProperty(w http.ResponseWriter, r *http.Request) {
	s.wg.Add(1)
	defer s.wg.Done()

	twinID := chi.URLParam(r, "twinID")
	featureID := chi.URLParam(r, "featureID")
	propKey := chi.URLParam(r, "propKey")

	if twinID == "" || featureID == "" || propKey == "" {
		respondError(w, http.StatusBadRequest, "Twin ID, Feature ID, and Property Key are required")
		return
	}

	dt, err := s.Registry.Get(twinID)
	if err != nil {
		if err == registry.ErrTwinNotFound {
			respondError(w, http.StatusNotFound, "Digital twin not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to get digital twin: "+err.Error())
		}
		return
	}

	feature, exists := dt.GetFeature(featureID)
	if !exists {
		respondError(w, http.StatusNotFound, "Feature not found")
		return
	}

	// Check if property exists
	_, exists = feature.GetProperty(propKey)
	if !exists {
		respondError(w, http.StatusNotFound, "Property not found")
		return
	}

	// Remove property
	feature.RemoveProperty(propKey)

	// Update the feature
	if err := dt.UpdateFeature(featureID, feature); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update feature: "+err.Error())
		return
	}

	// Update the twin in the registry
	if err := s.Registry.Update(dt); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update digital twin: "+err.Error())
		return
	}

	// Publish event
	s.PubSub.Publish("property.deleted", map[string]string{
		"twinId":      twinID,
		"featureId":   featureID,
		"propertyKey": propKey,
	})

	respondJSON(w, http.StatusOK, map[string]string{"message": "Property deleted"})
}
