package cloner

import (
    "errors"
    "fmt"
    "reflect"
    "strings"
    "sync"
)

var (
    stats      = make(map[string]int)
    statsMutex sync.Mutex // Mutex for concurrent access
)

// UpdateStats increments the count for the given type in the stats map.
func UpdateStats(typeName string) {
    statsMutex.Lock()
    defer statsMutex.Unlock()
    stats[typeName]++
}

func FormatStats() string {
    statsMutex.Lock()
    defer statsMutex.Unlock()
    b := strings.Builder{}
    for k, v := range stats {
        b.WriteString(fmt.Sprintf("%s: %d\n", k, v))
    }
    return b.String()
}

// Cloneable interface defines objects that can clone themselves.
type Cloneable interface {
    Clone(manager *CloneManager) (interface{}, error)
}

// Cloner defines custom cloners for external types.
type Cloner interface {
    Clone(value interface{}, manager *CloneManager) (interface{}, error)
}

// CloneManager manages the cloning process and tracks visited references.
type CloneManager struct {
    visited map[uintptr]interface{}
    cloners map[reflect.Type]Cloner
}

// NewCloneManager creates a new CloneManager instance.
func NewCloneManager() *CloneManager {
    return &CloneManager{
        visited: make(map[uintptr]interface{}),
        cloners: make(map[reflect.Type]Cloner),
    }
}

// RegisterCloner registers a custom Cloner for a specific type.
func (cm *CloneManager) RegisterCloner(t reflect.Type, cloner Cloner) {
    cm.cloners[t] = cloner
}

// Clone performs a deep clone of the given object.
func (cm *CloneManager) Clone(src interface{}) (interface{}, error) {
    return cm.deepClone(reflect.ValueOf(src))
}

// Clone performs a deep clone of the given object and returns it as the same type.
func Clone[T any](cm *CloneManager, src T) (T, error) {
    // Initialize the result as a zero value of type T
    var result T

    // Handle nil case for pointer types
    if reflect.ValueOf(src).IsNil() {
        return result, nil // Return zero value for nil pointers
    }

    // Deep clone the value
    clonedValue, err := cm.Clone(src)
    if err != nil {
        return result, err
    }

    // Assert the cloned value back to type T
    clonedValueTyped, ok := clonedValue.(T)
    if !ok {
        return result, errors.New("failed to cast cloned value to the original type")
    }

    return clonedValueTyped, nil
}

// deepClone handles recursive cloning and checks for registered Cloner or Cloneable interfaces.
func (cm *CloneManager) deepClone(src reflect.Value) (interface{}, error) {
    if !src.IsValid() {
        return nil, nil
    }

    // Check if the value implements Cloneable
    if src.CanInterface() {
        if cloneable, ok := src.Interface().(Cloneable); ok {
            // Delegate to the Cloneable method
            return cloneable.Clone(cm)
        }
    }

    // Check for registered Cloner
    if cloner, found := cm.cloners[src.Type()]; found {
        return cloner.Clone(src.Interface(), cm)
    }

    // Perform default deep clone logic (same as in the previous example)
    // Clone for Ptr, Slice, Array, Map, Struct, etc.
    switch src.Kind() {
    case reflect.Ptr:
        return cm.clonePtr(src)
    case reflect.Slice:
        return cm.cloneSlice(src)
    case reflect.Array:
        return cm.cloneArray(src)
    case reflect.Map:
        return cm.cloneMap(src)
    case reflect.Struct:
        return cm.cloneStruct(src)
    case reflect.Interface:
        return cm.cloneInterface(src)
    case reflect.Chan:
        return nil, errors.New("channels cannot be cloned")
    case reflect.Func:
        return nil, errors.New(fmt.Sprintf("functions cannot be cloned: %v", src))
        //return src.Interface(), nil // Functions are reference types but immutable
    default:
        return src.Interface(), nil // Primitive types can be copied directly
    }
}

// clonePtr clones a pointer value.
func (cm *CloneManager) clonePtr(src reflect.Value) (interface{}, error) {
    if src.IsNil() {
        return nil, nil
    }
    ptr := src.Pointer()
    if cloned, ok := cm.visited[ptr]; ok {
        return cloned, nil
    }

    // Recursively clone the pointed value
    cloned, err := cm.deepClone(src.Elem())
    if err != nil {
        return nil, err
    }
    UpdateStats(src.Kind().String())

    clonePtr := reflect.New(src.Elem().Type())
    clonePtr.Elem().Set(reflect.ValueOf(cloned))
    cm.visited[ptr] = clonePtr.Interface()
    return clonePtr.Interface(), nil
}

// cloneSlice clones a slice value.
func (cm *CloneManager) cloneSlice(src reflect.Value) (interface{}, error) {
    if src.IsNil() {
        return nil, nil
    }

    // Check if we've already cloned this slice
    ptr := src.Pointer()
    if cloned, found := cm.visited[ptr]; found {
        return cloned, nil
    }

    // Create a new slice of the same type and length
    clone := reflect.MakeSlice(src.Type(), src.Len(), src.Cap())
    cm.visited[ptr] = clone.Interface()

    // Iterate through the slice and deep clone each element
    for i := 0; i < src.Len(); i++ {
        clonedElem, err := cm.deepClone(src.Index(i))
        if err != nil {
            return nil, err
        }
        clone.Index(i).Set(reflect.ValueOf(clonedElem))
    }
    UpdateStats(src.Kind().String())
    return clone.Interface(), nil
}

// cloneArray clones an array value.
func (cm *CloneManager) cloneArray(src reflect.Value) (interface{}, error) {
    // Create a new array of the same type and length
    clone := reflect.New(src.Type()).Elem()

    // Clone each element in the array
    for i := 0; i < src.Len(); i++ {
        elem := src.Index(i)
        clonedElem, err := cm.deepClone(elem)
        if err != nil {
            return nil, err
        }
        clone.Index(i).Set(reflect.ValueOf(clonedElem))
    }
    UpdateStats(src.Kind().String())
    return clone.Interface(), nil
}

// cloneMap clones a map value.
func (cm *CloneManager) cloneMap(src reflect.Value) (interface{}, error) {
    if src.IsNil() {
        return nil, nil
    }

    // Use the map's underlying pointer as the key
    ptr := src.Pointer()

    // Check if we've already cloned this map
    if cloned, found := cm.visited[ptr]; found {
        return cloned, nil
    }

    // Create a new map of the same type
    clone := reflect.MakeMapWithSize(src.Type(), src.Len())
    cm.visited[ptr] = clone.Interface()

    // Deep clone each key-value pair in the map
    for _, key := range src.MapKeys() {
        clonedKey, err := cm.deepClone(key)
        if err != nil {
            return nil, err
        }

        clonedValue, err := cm.deepClone(src.MapIndex(key))
        if err != nil {
            return nil, err
        }

        clone.SetMapIndex(reflect.ValueOf(clonedKey), reflect.ValueOf(clonedValue))
    }
    UpdateStats(src.Kind().String())
    return clone.Interface(), nil
}

// cloneStruct clones a struct value.
func (cm *CloneManager) cloneStruct(src reflect.Value) (interface{}, error) {
    // Create a new struct of the same type
    clone := reflect.New(src.Type()).Elem()

    // Clone each field of the struct
    for i := 0; i < src.NumField(); i++ {
        field := src.Field(i)
        clonedFieldRef := clone.Field(i)
        if clonedFieldRef.CanSet() {
            clonedField, err := cm.deepClone(field)
            if err != nil {
                return nil, err
            }
            //clonedFieldRef.Set(reflect.ValueOf(clonedField))
            // Ensure the cloned value is not zero
            if !clonedFieldRef.IsValid() {
                return nil, fmt.Errorf("cannot set invalid field at index %d", i)
            }

            // Set the cloned field only if the value is valid
            if clonedField != nil {
                clonedFieldRef.Set(reflect.ValueOf(clonedField))
            }
        }
    }
    UpdateStats(src.Kind().String() + " " + src.Type().String())
    return clone.Interface(), nil
}

func (cm *CloneManager) cloneInterface(src reflect.Value) (interface{}, error) {
    // Get the underlying value
    underlyingValue := src.Elem()

    // Check for nil underlying value
    if !underlyingValue.IsValid() {
        return nil, nil // Return nil for nil underlying value
    }
    if src.IsNil() {
        return nil, nil
    }
    // Clone the underlying value
    clonedValue, err := cm.deepClone(underlyingValue)
    if err != nil {
        return nil, err
    }
    UpdateStats(src.Kind().String() + " " + src.Type().String())
    // Return as an interface type
    return reflect.ValueOf(clonedValue).Convert(src.Type()).Interface(), nil
}
