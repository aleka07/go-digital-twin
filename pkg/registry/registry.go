package registry

import (
	"errors"
	"sync"

	"github.com/aleka07/go-digital-twin/pkg/twin"
)

// Common errors
var (
	ErrTwinNotFound      = errors.New("digital twin not found")
	ErrTwinAlreadyExists = errors.New("digital twin already exists")
)

// Registry provides thread-safe storage for digital twins
type Registry struct {
	twins map[string]*twin.DigitalTwin
	mutex sync.RWMutex
}

// NewRegistry creates a new registry
func NewRegistry() *Registry {
	return &Registry{
		twins: make(map[string]*twin.DigitalTwin),
	}
}

// Create adds a new digital twin to the registry
func (r *Registry) Create(dt *twin.DigitalTwin) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.twins[dt.ID]; exists {
		return ErrTwinAlreadyExists
	}

	r.twins[dt.ID] = dt
	return nil
}

// Get retrieves a digital twin by ID
func (r *Registry) Get(id string) (*twin.DigitalTwin, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	dt, exists := r.twins[id]
	if !exists {
		return nil, ErrTwinNotFound
	}

	return dt, nil
}

// Update updates an existing digital twin
func (r *Registry) Update(dt *twin.DigitalTwin) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.twins[dt.ID]; !exists {
		return ErrTwinNotFound
	}

	r.twins[dt.ID] = dt
	return nil
}

// Delete removes a digital twin from the registry
func (r *Registry) Delete(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.twins[id]; !exists {
		return ErrTwinNotFound
	}

	delete(r.twins, id)
	return nil
}

// List returns all digital twins in the registry
func (r *Registry) List() []*twin.DigitalTwin {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	twins := make([]*twin.DigitalTwin, 0, len(r.twins))
	for _, dt := range r.twins {
		twins = append(twins, dt)
	}

	return twins
}

// FindByAttribute returns twins that have a specific attribute value
func (r *Registry) FindByAttribute(key string, value interface{}) []*twin.DigitalTwin {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []*twin.DigitalTwin

	for _, dt := range r.twins {
		if attrValue, exists := dt.GetAttribute(key); exists && attrValue == value {
			result = append(result, dt)
		}
	}

	return result
}

// FindByFeature returns twins that have a specific feature
func (r *Registry) FindByFeature(featureID string) []*twin.DigitalTwin {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []*twin.DigitalTwin

	for _, dt := range r.twins {
		if _, exists := dt.GetFeature(featureID); exists {
			result = append(result, dt)
		}
	}

	return result
}
