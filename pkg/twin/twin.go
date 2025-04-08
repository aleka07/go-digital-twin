package twin

import (
	"errors"
	"sync"
	"time"
)

// Common errors
var (
	ErrFeatureNotFound    = errors.New("feature not found")
	ErrFeatureAlreadyExists = errors.New("feature already exists")
	ErrPropertyNotFound   = errors.New("property not found")
	ErrInvalidValue       = errors.New("invalid value")
)

// DigitalTwin represents a digital representation of a physical entity
type DigitalTwin struct {
	ID         string                  // Unique identifier
	Type       string                  // Type of the twin
	Definition string                  // Optional definition reference
	Attributes map[string]interface{}  // General attributes
	Features   map[string]FeatureState // Features of the twin
	mutex      sync.RWMutex            // For thread safety
	CreatedAt  time.Time               // Creation timestamp
	ModifiedAt time.Time               // Last modification timestamp
}

// NewDigitalTwin creates a new digital twin with the given ID and type
func NewDigitalTwin(id, twinType string) *DigitalTwin {
	now := time.Now()
	return &DigitalTwin{
		ID:         id,
		Type:       twinType,
		Attributes: make(map[string]interface{}),
		Features:   make(map[string]FeatureState),
		CreatedAt:  now,
		ModifiedAt: now,
	}
}

// SetDefinition sets the definition of the digital twin
func (dt *DigitalTwin) SetDefinition(definition string) {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()
	
	dt.Definition = definition
	dt.ModifiedAt = time.Now()
}

// GetDefinition returns the definition of the digital twin
func (dt *DigitalTwin) GetDefinition() string {
	dt.mutex.RLock()
	defer dt.mutex.RUnlock()
	
	return dt.Definition
}

// GetAttribute returns the value of an attribute
func (dt *DigitalTwin) GetAttribute(key string) (interface{}, bool) {
	dt.mutex.RLock()
	defer dt.mutex.RUnlock()
	
	val, exists := dt.Attributes[key]
	return val, exists
}

// SetAttribute sets the value of an attribute
func (dt *DigitalTwin) SetAttribute(key string, value interface{}) {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()
	
	dt.Attributes[key] = value
	dt.ModifiedAt = time.Now()
}

// RemoveAttribute removes an attribute
func (dt *DigitalTwin) RemoveAttribute(key string) {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()
	
	delete(dt.Attributes, key)
	dt.ModifiedAt = time.Now()
}

// GetAllAttributes returns a copy of all attributes
func (dt *DigitalTwin) GetAllAttributes() map[string]interface{} {
	dt.mutex.RLock()
	defer dt.mutex.RUnlock()
	
	attributes := make(map[string]interface{}, len(dt.Attributes))
	for k, v := range dt.Attributes {
		attributes[k] = v
	}
	return attributes
}

// GetFeature returns a feature by ID
func (dt *DigitalTwin) GetFeature(id string) (FeatureState, bool) {
	dt.mutex.RLock()
	defer dt.mutex.RUnlock()
	
	feature, exists := dt.Features[id]
	return feature, exists
}

// AddFeature adds a new feature
func (dt *DigitalTwin) AddFeature(id string, feature FeatureState) error {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()
	
	if _, exists := dt.Features[id]; exists {
		return ErrFeatureAlreadyExists
	}
	
	dt.Features[id] = feature
	dt.ModifiedAt = time.Now()
	return nil
}

// UpdateFeature updates an existing feature
func (dt *DigitalTwin) UpdateFeature(id string, feature FeatureState) error {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()
	
	if _, exists := dt.Features[id]; !exists {
		return ErrFeatureNotFound
	}
	
	dt.Features[id] = feature
	dt.ModifiedAt = time.Now()
	return nil
}

// RemoveFeature removes a feature
func (dt *DigitalTwin) RemoveFeature(id string) error {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()
	
	if _, exists := dt.Features[id]; !exists {
		return ErrFeatureNotFound
	}
	
	delete(dt.Features, id)
	dt.ModifiedAt = time.Now()
	return nil
}

// GetAllFeatures returns a copy of all features
func (dt *DigitalTwin) GetAllFeatures() map[string]FeatureState {
	dt.mutex.RLock()
	defer dt.mutex.RUnlock()
	
	features := make(map[string]FeatureState, len(dt.Features))
	for k, v := range dt.Features {
		features[k] = v
	}
	return features
}
