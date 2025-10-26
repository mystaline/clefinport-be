package entity

import (
	"reflect"

	"github.com/mystaline/clefinport-be/pkg/functions"
)

// Container represents a high-level pseudo map-like function.
type Container[K any, V any] interface {
	Get(key K, fallback V) V
	Set(key K, value V)
	Delete(key K)
	Exist(key K) bool
}

// M is a map-like struct used by Map as a high-level map-like data types.
type M[K any, V any] struct {
	Key   K
	Value V
}

// Map is the implementation of Container as a high-level map-like data type.
type Map[K any, V any] struct {
	_m []M[K, V]
}

// remove removes an element from a slice of the provided index.
func remove[T any](s []T, i int) []T {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func MakeMap[K any, V any]() Map[K, V] {
	return Map[K, V]{
		_m: []M[K, V]{},
	}
}

// Exist checks whether a key exists in a map.
func (m *Map[K, V]) Exist(key K) bool {
	exists := functions.Any(m._m, func(m M[K, V], _ int) bool {
		return reflect.DeepEqual(m.Key, key)
	})

	return exists
}

// Get gets the value from the provided key, use fallback if the key doesn't exist.
func (m *Map[K, V]) Get(key K, fallback V) V {
	if !m.Exist(key) {
		return fallback
	}

	v := functions.Find(m._m, func(m M[K, V], _ int) bool {
		return reflect.DeepEqual(m.Key, key)
	})

	return v.Value
}

// Set sets the value from the provided key, if the key exists previously, delete the existing value to be overwritten.
func (m *Map[K, V]) Set(key K, value V) {
	m.Delete(key)

	m._m = append(m._m, M[K, V]{Key: key, Value: value})
}

// Delete deletes the key-value pair from the map, doesn't return anything whether the key exists or not.
func (m *Map[K, V]) Delete(key K) {
	functions.ForEach(m._m, func(_m M[K, V], index int) {
		if reflect.DeepEqual(_m.Key, key) {
			m._m = remove(m._m, index)
		}
	})
}
