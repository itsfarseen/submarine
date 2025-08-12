package polkadot_scale_schema

import (
	"fmt"
	"submarine/rust_types"
	"submarine/rust_types/sanitizer"
	s "submarine/scale"
	"sync"
)

type Registry struct {
	types map[string][]TypeEntry
}

func NewRegistry() *Registry {
	return &Registry{
		types: make(map[string][]TypeEntry),
	}
}

type LazyType struct {
	raw    string
	parsed *s.Type
	mutex  sync.Mutex
}

func NewLazyType(raw string) *LazyType {
	return &LazyType{
		raw: raw,
	}
}

func (lt *LazyType) ToScaleType() (*s.Type, error) {
	lt.mutex.Lock()
	defer lt.mutex.Unlock()

	if lt.parsed != nil {
		return lt.parsed, nil
	}

	rustType := sanitizer.ParseAndSanitize(lt.raw)
	scaleType, errSpan := rust_types.ToScaleSchema(&rustType)
	if errSpan != nil {
		return nil, fmt.Errorf("failed to convert to scale type: %v", errSpan)
	}

	lt.parsed = &scaleType
	return lt.parsed, nil
}

type TypeEntry struct {
	Module string
	Type   *LazyType
}

func (r *Registry) Lookup(moduleName, typeName string) (*LazyType, error) {
	entries, exists := r.types[typeName]
	if !exists {
		return nil, fmt.Errorf("type %s not found", typeName)
	}

	if len(entries) == 1 {
		return entries[0].Type, nil
	}

	for _, entry := range entries {
		if entry.Module == moduleName {
			return entry.Type, nil
		}
	}

	return nil, fmt.Errorf("type %s not found in module %s", typeName, moduleName)
}

func (r *Registry) GetAllTypes(typeName string) ([]TypeEntry, bool) {
	entries, exists := r.types[typeName]
	return entries, exists
}

func (r *Registry) GetModuleTypes(moduleName string) map[string]*LazyType {
	result := make(map[string]*LazyType)
	for typeName, entries := range r.types {
		for _, entry := range entries {
			if entry.Module == moduleName {
				result[typeName] = entry.Type
			}
		}
	}
	return result
}
