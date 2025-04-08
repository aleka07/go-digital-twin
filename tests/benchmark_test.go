package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/aleka07/go-digital-twin/pkg/api"
	"github.com/aleka07/go-digital-twin/pkg/messaging_sim"
	"github.com/aleka07/go-digital-twin/pkg/registry"
	"github.com/aleka07/go-digital-twin/pkg/twin"
)

type contextKey string

// BenchmarkTwinCreation measures the performance of creating digital twins
func BenchmarkTwinCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := fmt.Sprintf("twin-%d", i)
		_ = twin.NewDigitalTwin(id, "sensor")
	}
}

// BenchmarkTwinAttributeOperations measures the performance of attribute operations
func BenchmarkTwinAttributeOperations(b *testing.B) {
	dt := twin.NewDigitalTwin("benchmark-twin", "device")

	b.Run("SetAttribute", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("attr-%d", i)
			dt.SetAttribute(key, i)
		}
	})

	b.Run("GetAttribute", func(b *testing.B) {
		// Pre-populate with attributes
		for i := 0; i < 1000; i++ {
			key := fmt.Sprintf("attr-%d", i)
			dt.SetAttribute(key, i)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("attr-%d", i%1000)
			_, _ = dt.GetAttribute(key)
		}
	})

	b.Run("RemoveAttribute", func(b *testing.B) {
		// Pre-populate with attributes
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("remove-attr-%d", i)
			dt.SetAttribute(key, i)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("remove-attr-%d", i)
			dt.RemoveAttribute(key)
		}
	})
}

// BenchmarkTwinFeatureOperations measures the performance of feature operations
func BenchmarkTwinFeatureOperations(b *testing.B) {
	dt := twin.NewDigitalTwin("benchmark-twin", "device")

	b.Run("AddFeature", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			featureID := fmt.Sprintf("feature-%d", i)
			feature := twin.NewFeatureState()
			feature.SetProperty("value", i)
			_ = dt.AddFeature(featureID, *feature)
		}
	})

	b.Run("GetFeature", func(b *testing.B) {
		// Pre-populate with features
		for i := 0; i < 1000; i++ {
			featureID := fmt.Sprintf("get-feature-%d", i)
			feature := twin.NewFeatureState()
			feature.SetProperty("value", i)
			dt.AddFeature(featureID, *feature)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			featureID := fmt.Sprintf("get-feature-%d", i%1000)
			_, _ = dt.GetFeature(featureID)
		}
	})

	b.Run("UpdateFeature", func(b *testing.B) {
		// Pre-populate with features
		for i := 0; i < 1000; i++ {
			featureID := fmt.Sprintf("update-feature-%d", i)
			feature := twin.NewFeatureState()
			feature.SetProperty("value", i)
			dt.AddFeature(featureID, *feature)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			featureID := fmt.Sprintf("update-feature-%d", i%1000)
			feature, _ := dt.GetFeature(featureID)
			feature.SetProperty("value", i)
			_ = dt.UpdateFeature(featureID, feature)
		}
	})
}

// BenchmarkFeaturePropertyOperations measures the performance of feature property operations
func BenchmarkFeaturePropertyOperations(b *testing.B) {
	feature := twin.NewFeatureState()

	b.Run("SetProperty", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("prop-%d", i)
			feature.SetProperty(key, i)
		}
	})

	b.Run("GetProperty", func(b *testing.B) {
		// Pre-populate with properties
		for i := 0; i < 1000; i++ {
			key := fmt.Sprintf("get-prop-%d", i)
			feature.SetProperty(key, i)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("get-prop-%d", i%1000)
			_, _ = feature.GetProperty(key)
		}
	})
}

// BenchmarkRegistryConcurrentAccess measures the performance of concurrent registry operations
func BenchmarkRegistryConcurrentAccess(b *testing.B) {
	registry := registry.NewRegistry()

	// Pre-populate with twins
	for i := 0; i < 1000; i++ {
		id := fmt.Sprintf("twin-%d", i)
		dt := twin.NewDigitalTwin(id, "sensor")
		registry.Create(dt)
	}

	b.Run("ConcurrentReads", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				id := fmt.Sprintf("twin-%d", i%1000)
				_, _ = registry.Get(id)
				i++
			}
		})
	})

	b.Run("ConcurrentWrites", func(b *testing.B) {
		var counter int32
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				i := atomic.AddInt32(&counter, 1)
				id := fmt.Sprintf("new-twin-%d", i)
				dt := twin.NewDigitalTwin(id, "sensor")
				registry.Create(dt)
			}
		})
	})

	b.Run("MixedReadWrite", func(b *testing.B) {
		var counter int32
		var wg sync.WaitGroup

		b.ResetTimer()
		// 75% reads, 25% writes
		for i := 0; i < b.N; i++ {
			wg.Add(1)
			if i%4 == 0 {
				// Write operation
				go func() {
					defer wg.Done()
					j := atomic.AddInt32(&counter, 1)
					id := fmt.Sprintf("mixed-twin-%d", j)
					dt := twin.NewDigitalTwin(id, "sensor")
					registry.Create(dt)
				}()
			} else {
				// Read operation
				go func(idx int) {
					defer wg.Done()
					id := fmt.Sprintf("twin-%d", idx%1000)
					_, _ = registry.Get(id)
				}(i)
			}
		}
		wg.Wait()
	})
}

// BenchmarkPubSubThroughput measures the performance of the messaging system
func BenchmarkPubSubThroughput(b *testing.B) {
	pubsub := messaging_sim.NewPubSub()

	b.Run("SinglePublisherMultipleSubscribers", func(b *testing.B) {
		const numSubscribers = 10
		var wg sync.WaitGroup

		// Create subscribers
		channels := make([]chan messaging_sim.Message, numSubscribers)
		for i := 0; i < numSubscribers; i++ {
			channels[i] = pubsub.Subscribe("benchmark-topic")

			// Start subscriber goroutine
			wg.Add(1)
			go func(ch chan messaging_sim.Message, id int) {
				defer wg.Done()
				count := 0
				for range ch {
					count++
					if count == b.N {
						return
					}
				}
			}(channels[i], i)
		}

		// Publisher
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			pubsub.Publish("benchmark-topic", i)
		}

		wg.Wait()
	})

	b.Run("MultiplePublishersOneSubscriber", func(b *testing.B) {
		ch := pubsub.Subscribe("benchmark-topic-2")

		// Start subscriber goroutine
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			count := 0
			for range ch {
				count++
				if count == b.N {
					return
				}
			}
		}()

		// Publishers
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				pubsub.Publish("benchmark-topic-2", i)
				i++
			}
		})

		wg.Wait()
	})
}

// BenchmarkAPIEndpoints measures the performance of API endpoints
func BenchmarkAPIEndpoints(b *testing.B) {
	server := api.NewServer(registry.NewRegistry(), messaging_sim.NewPubSub())

	// Pre-populate with a twin
	dt := twin.NewDigitalTwin("api-twin", "device")
	feature := twin.NewFeatureState()
	feature.SetProperty("value", 42)
	dt.AddFeature("test-feature", *feature)
	server.Registry.Create(dt)

	b.Run("GetTwin", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest("GET", "/twins/api-twin", nil)
			req = req.WithContext(context.WithValue(req.Context(), contextKey("twinID"), "api-twin"))
			w := httptest.NewRecorder()
			server.GetTwin(w, req)
		}
	})

	b.Run("GetFeature", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest("GET", "/twins/api-twin/features/test-feature", nil)
			ctx := context.WithValue(req.Context(), contextKey("twinID"), "api-twin")
			ctx = context.WithValue(ctx, contextKey("featureID"), "test-feature")
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()
			server.GetFeature(w, req)
		}
	})

	b.Run("CreateTwin", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			twinData := map[string]interface{}{
				"id":   fmt.Sprintf("bench-twin-%d", i),
				"type": "sensor",
			}
			jsonData, _ := json.Marshal(twinData)
			req := httptest.NewRequest("POST", "/twins", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			server.CreateTwin(w, req)
		}
	})
}
