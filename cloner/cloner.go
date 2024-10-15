package cloner

import (
    "errors"
    "fmt"
    "reflect"
)

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

    // Create a new slice of the same type and length
    clone := reflect.MakeSlice(src.Type(), src.Len(), src.Cap())

    // Clone each element in the slice
    for i := 0; i < src.Len(); i++ {
        elem := src.Index(i)
        clonedElem, err := cm.deepClone(elem)
        if err != nil {
            return nil, err
        }
        clone.Index(i).Set(reflect.ValueOf(clonedElem))
    }

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

    return clone.Interface(), nil
}

// cloneMap clones a map value.
func (cm *CloneManager) cloneMap(src reflect.Value) (interface{}, error) {
    if src.IsNil() {
        return nil, nil
    }

    // Create a new map of the same type
    clone := reflect.MakeMapWithSize(src.Type(), src.Len())

    // Clone each key-value pair
    for _, key := range src.MapKeys() {
        clonedKey, err := cm.deepClone(key)
        if err != nil {
            return nil, err
        }

        value := src.MapIndex(key)
        clonedValue, err := cm.deepClone(value)
        if err != nil {
            return nil, err
        }

        clone.SetMapIndex(reflect.ValueOf(clonedKey), reflect.ValueOf(clonedValue))
    }

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
            clonedFieldRef.Set(reflect.ValueOf(clonedField))
        }
    }

    return clone.Interface(), nil
}
