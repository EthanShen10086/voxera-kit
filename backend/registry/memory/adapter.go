// Package memory provides an in-memory implementation of the registry.ServiceRegistry
// and registry.ConfigCenter interfaces.
package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/EthanShen10086/voxera-kit/registry"
)

type serviceWatcher struct {
	callback func([]*registry.ServiceInstance)
}

type configWatcher struct {
	callback func(*registry.ConfigValue)
}

// Adapter is an in-memory service registry and configuration center.
type Adapter struct {
	mu              sync.RWMutex
	instances       map[string]*registry.ServiceInstance
	configs         map[string]*registry.ConfigValue
	serviceWatchers map[string][]serviceWatcher
	configWatchers  map[string][]configWatcher
}

// New creates a new in-memory registry adapter.
func New() *Adapter {
	return &Adapter{
		instances:       make(map[string]*registry.ServiceInstance),
		configs:         make(map[string]*registry.ConfigValue),
		serviceWatchers: make(map[string][]serviceWatcher),
		configWatchers:  make(map[string][]configWatcher),
	}
}

func (a *Adapter) Register(_ context.Context, instance *registry.ServiceInstance) error {
	if instance.ID == "" {
		return fmt.Errorf("registry: instance ID is required")
	}

	a.mu.Lock()
	instance.RegisteredAt = time.Now()
	if instance.Status == 0 {
		instance.Status = registry.Up
	}
	a.instances[instance.ID] = instance
	watchers := a.serviceWatchers[instance.Name]
	a.mu.Unlock()

	a.notifyServiceWatchers(instance.Name, watchers)
	return nil
}

func (a *Adapter) Deregister(_ context.Context, instanceID string) error {
	a.mu.Lock()
	instance, exists := a.instances[instanceID]
	if !exists {
		a.mu.Unlock()
		return fmt.Errorf("registry: instance %q not found", instanceID)
	}
	serviceName := instance.Name
	delete(a.instances, instanceID)
	watchers := a.serviceWatchers[serviceName]
	a.mu.Unlock()

	a.notifyServiceWatchers(serviceName, watchers)
	return nil
}

func (a *Adapter) Discover(_ context.Context, serviceName string) ([]*registry.ServiceInstance, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var result []*registry.ServiceInstance
	for _, inst := range a.instances {
		if inst.Name == serviceName && inst.Status == registry.Up {
			result = append(result, inst)
		}
	}
	return result, nil
}

func (a *Adapter) Watch(_ context.Context, serviceName string, callback func([]*registry.ServiceInstance)) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.serviceWatchers[serviceName] = append(a.serviceWatchers[serviceName], serviceWatcher{callback: callback})
	return nil
}

func (a *Adapter) Heartbeat(_ context.Context, instanceID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	inst, exists := a.instances[instanceID]
	if !exists {
		return fmt.Errorf("registry: instance %q not found", instanceID)
	}
	inst.Status = registry.Up
	return nil
}

func (a *Adapter) Get(_ context.Context, key string) (*registry.ConfigValue, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	cv, exists := a.configs[key]
	if !exists {
		return nil, fmt.Errorf("registry: config key %q not found", key)
	}
	return cv, nil
}

func (a *Adapter) Set(_ context.Context, key, value string) error {
	a.mu.Lock()
	cv, exists := a.configs[key]
	if exists {
		cv.Value = value
		cv.Version++
		cv.UpdatedAt = time.Now()
	} else {
		cv = &registry.ConfigValue{
			Key:       key,
			Value:     value,
			Version:   1,
			UpdatedAt: time.Now(),
		}
		a.configs[key] = cv
	}
	watchers := a.configWatchers[key]
	snapshot := *cv
	a.mu.Unlock()

	for _, w := range watchers {
		w.callback(&snapshot)
	}
	return nil
}

func (a *Adapter) Delete(_ context.Context, key string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.configs[key]; !exists {
		return fmt.Errorf("registry: config key %q not found", key)
	}
	delete(a.configs, key)
	return nil
}

// Watch on ConfigCenter registers a callback for changes to the given configuration key.
func (a *Adapter) WatchConfig(_ context.Context, key string, callback func(*registry.ConfigValue)) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.configWatchers[key] = append(a.configWatchers[key], configWatcher{callback: callback})
	return nil
}

func (a *Adapter) List(_ context.Context, prefix string) ([]*registry.ConfigValue, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var result []*registry.ConfigValue
	for k, cv := range a.configs {
		if strings.HasPrefix(k, prefix) {
			result = append(result, cv)
		}
	}
	return result, nil
}

func (a *Adapter) notifyServiceWatchers(serviceName string, watchers []serviceWatcher) {
	if len(watchers) == 0 {
		return
	}

	a.mu.RLock()
	var instances []*registry.ServiceInstance
	for _, inst := range a.instances {
		if inst.Name == serviceName {
			instances = append(instances, inst)
		}
	}
	a.mu.RUnlock()

	for _, w := range watchers {
		w.callback(instances)
	}
}
