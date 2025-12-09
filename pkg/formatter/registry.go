package formatter

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

var (
	registryMu sync.RWMutex
	registry   = make(map[string]Factory)
)

// Register adds a formatter factory to the registry.
//
// It panics if the name is empty, the factory is nil, or the name is already registered.
func Register(name string, factory Factory) {
	if name == "" {
		panic("formatter name cannot be empty")
	}
	if factory == nil {
		panic("formatter factory cannot be nil")
	}

	registryMu.Lock()
	defer registryMu.Unlock()

	if _, exists := registry[name]; exists {
		panic(fmt.Sprintf("formatter %q is already registered", name))
	}

	registry[name] = factory
}

// Get returns a formatter for the provided name, falling back to DefaultFormat when empty.
func Get(name string) (Formatter, error) {
	formatName := name
	if formatName == "" {
		formatName = DefaultFormat
	}

	registryMu.RLock()
	factory, ok := registry[formatName]
	registryMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown format %q. Available formats: %s", formatName, strings.Join(Available(), ", "))
	}

	return factory(), nil
}

// Available returns the list of registered formatter names in sorted order.
func Available() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}
