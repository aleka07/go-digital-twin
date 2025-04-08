package twin

import (
	"sync"
	"time"
)

// FeatureState represents the state of a feature in a digital twin
type FeatureState struct {
	Properties    map[string]interface{} // Current properties
	DesiredProps  map[string]interface{} // Desired properties (target state)
	Definition    []string               // Feature definition identifiers
	LastModified  time.Time              // Last modification timestamp
	mutex         sync.RWMutex           // For thread safety
}

// NewFeatureState creates a new feature state
func NewFeatureState() *FeatureState {
	return &FeatureState{
		Properties:   make(map[string]interface{}),
		DesiredProps: make(map[string]interface{}),
		Definition:   []string{},
		LastModified: time.Now(),
	}
}

// GetProperty returns the value of a property
func (fs *FeatureState) GetProperty(key string) (interface{}, bool) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	
	val, exists := fs.Properties[key]
	return val, exists
}

// SetProperty sets the value of a property
func (fs *FeatureState) SetProperty(key string, value interface{}) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	
	fs.Properties[key] = value
	fs.LastModified = time.Now()
}

// RemoveProperty removes a property
func (fs *FeatureState) RemoveProperty(key string) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	
	delete(fs.Properties, key)
	fs.LastModified = time.Now()
}

// GetAllProperties returns a copy of all properties
func (fs *FeatureState) GetAllProperties() map[string]interface{} {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	
	properties := make(map[string]interface{}, len(fs.Properties))
	for k, v := range fs.Properties {
		properties[k] = v
	}
	return properties
}

// GetDesiredProperty returns the value of a desired property
func (fs *FeatureState) GetDesiredProperty(key string) (interface{}, bool) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	
	val, exists := fs.DesiredProps[key]
	return val, exists
}

// SetDesiredProperty sets the value of a desired property
func (fs *FeatureState) SetDesiredProperty(key string, value interface{}) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	
	fs.DesiredProps[key] = value
	fs.LastModified = time.Now()
}

// RemoveDesiredProperty removes a desired property
func (fs *FeatureState) RemoveDesiredProperty(key string) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	
	delete(fs.DesiredProps, key)
	fs.LastModified = time.Now()
}

// GetAllDesiredProperties returns a copy of all desired properties
func (fs *FeatureState) GetAllDesiredProperties() map[string]interface{} {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	
	desiredProps := make(map[string]interface{}, len(fs.DesiredProps))
	for k, v := range fs.DesiredProps {
		desiredProps[k] = v
	}
	return desiredProps
}

// SetDefinition sets the definition identifiers for the feature
func (fs *FeatureState) SetDefinition(definitions []string) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	
	fs.Definition = make([]string, len(definitions))
	copy(fs.Definition, definitions)
	fs.LastModified = time.Now()
}

// GetDefinition returns a copy of the definition identifiers
func (fs *FeatureState) GetDefinition() []string {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	
	definitions := make([]string, len(fs.Definition))
	copy(definitions, fs.Definition)
	return definitions
}
