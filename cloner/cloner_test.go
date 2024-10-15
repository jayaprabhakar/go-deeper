package cloner_test

import (
    "github.com/jayaprabhakar/go-deeper/cloner"
    "reflect"
    "testing"
)

// Helper function to check deep equality of values
func deepEqual(t *testing.T, got, want interface{}) {
    if !reflect.DeepEqual(got, want) {
        t.Errorf("got = %+v, want = %+v", got, want)
    }
}

// Test for cloning basic types (int, string, etc.)
func TestCloneBasicTypes(t *testing.T) {
    cm := cloner.NewCloneManager()

    // Test integers
    originalInt := 42
    clonedInt, err := cm.Clone(originalInt)
    if err != nil {
        t.Errorf("Clone failed: %v", err)
    }
    deepEqual(t, clonedInt, originalInt)

    // Test strings
    originalString := "hello"
    clonedString, err := cm.Clone(originalString)
    if err != nil {
        t.Errorf("Clone failed: %v", err)
    }
    deepEqual(t, clonedString, originalString)
}

// Test for cloning pointers
func TestClonePointer(t *testing.T) {
    cm := cloner.NewCloneManager()

    original := 42
    originalPtr := &original

    cloned, err := cm.Clone(originalPtr)
    if err != nil {
        t.Errorf("Clone failed: %v", err)
    }

    // Check that the value is the same
    deepEqual(t, cloned, originalPtr)

    // Ensure that the cloned pointer points to a different location
    if cloned == originalPtr {
        t.Errorf("Clone did not create a new pointer")
    }
}

// Test for cloning slices
func TestCloneSlice(t *testing.T) {
    cm := cloner.NewCloneManager()

    originalSlice := []int{1, 2, 3, 4}
    clonedSlice, err := cm.Clone(originalSlice)
    if err != nil {
        t.Errorf("Clone failed: %v", err)
    }

    deepEqual(t, clonedSlice, originalSlice)

    // Ensure modifying the original does not affect the clone
    originalSlice[0] = 100
    if reflect.DeepEqual(clonedSlice, originalSlice) {
        t.Errorf("Modifying the original affected the cloned slice")
    }
}

// Test for cloning arrays
func TestCloneArray(t *testing.T) {
    cm := cloner.NewCloneManager()

    originalArray := [3]int{1, 2, 3}
    clonedArray, err := cm.Clone(originalArray)
    if err != nil {
        t.Errorf("Clone failed: %v", err)
    }

    deepEqual(t, clonedArray, originalArray)

    // Ensure modifying the original does not affect the clone
    originalArray[0] = 100
    if reflect.DeepEqual(clonedArray, originalArray) {
        t.Errorf("Modifying the original affected the cloned array")
    }
}

// Test for cloning maps
func TestCloneMap(t *testing.T) {
    cm := cloner.NewCloneManager()

    originalMap := map[string]int{"a": 1, "b": 2}
    clonedMap, err := cm.Clone(originalMap)
    if err != nil {
        t.Errorf("Clone failed: %v", err)
    }

    deepEqual(t, clonedMap, originalMap)

    // Ensure modifying the original does not affect the clone
    originalMap["a"] = 100
    if reflect.DeepEqual(clonedMap, originalMap) {
        t.Errorf("Modifying the original affected the cloned map")
    }
}

// Test for cloning structs
type TestStruct struct {
    A int
    B *int
}

func TestCloneStruct(t *testing.T) {
    cm := cloner.NewCloneManager()

    original := TestStruct{A: 42, B: new(int)}
    *original.B = 100

    cloned, err := cm.Clone(original)
    if err != nil {
        t.Errorf("Clone failed: %v", err)
    }

    deepEqual(t, cloned, original)

    // Ensure the pointer is deeply cloned
    original.A = 0
    *original.B = 0
    if reflect.DeepEqual(cloned, original) {
        t.Errorf("Modifying the original affected the cloned struct")
    }
}

// Test for cloning interfaces
func TestCloneInterface(t *testing.T) {
    cm := cloner.NewCloneManager()

    var original interface{}
    original = 42

    cloned, err := cm.Clone(original)
    if err != nil {
        t.Errorf("Clone failed: %v", err)
    }

    deepEqual(t, cloned, original)

    // Test a more complex interface
    original = &TestStruct{A: 42, B: new(int)}
    *original.(*TestStruct).B = 100

    cloned, err = cm.Clone(original)
    if err != nil {
        t.Errorf("Clone failed: %v", err)
    }

    deepEqual(t, cloned, original)

    // Ensure the pointer inside the struct is deeply cloned
    *original.(*TestStruct).B = 0
    if reflect.DeepEqual(cloned, original) {
        t.Errorf("Modifying the original affected the cloned interface")
    }
}

// Test for `nil` handling
func TestCloneNilValues(t *testing.T) {
    cm := cloner.NewCloneManager()

    var nilPointer *int
    clonedPointer, err := cm.Clone(nilPointer)
    if err != nil {
        t.Errorf("Clone failed: %v", err)
    }
    if clonedPointer != nil {
        t.Errorf("Cloning a nil pointer should return nil")
    }

    var nilSlice []int
    clonedSlice, err := cm.Clone(nilSlice)
    if err != nil {
        t.Errorf("Clone failed: %v", err)
    }
    if clonedSlice != nil {
        t.Errorf("Cloning a nil slice should return nil")
    }

    var nilMap map[string]int
    clonedMap, err := cm.Clone(nilMap)
    if err != nil {
        t.Errorf("Clone failed: %v", err)
    }
    if clonedMap != nil {
        t.Errorf("Cloning a nil map should return nil")
    }

    var nilInterface interface{}
    clonedInterface, err := cm.Clone(nilInterface)
    if err != nil {
        t.Errorf("Clone failed: %v", err)
    }
    if clonedInterface != nil {
        t.Errorf("Cloning a nil interface should return nil")
    }
}

func TestClonePointerEquivalence(t *testing.T) {
    cm := cloner.NewCloneManager()

    // Test 1: Struct with two pointers to the same value
    a := 100
    original := struct {
        A *int
        B *int
    }{A: &a, B: &a} // Both A and B point to the same int

    cloned, err := cm.Clone(original)
    if err != nil {
        t.Fatalf("Clone failed: %v", err)
    }

    clonedStruct := cloned.(struct {
        A *int
        B *int
    })

    // Ensure that both pointers in the cloned struct point to the same value
    if clonedStruct.A != clonedStruct.B {
        t.Fatalf("Cloned pointers A and B do not point to the same value")
    }

    // Ensure the value they point to is correct
    if *clonedStruct.A != 100 {
        t.Errorf("Cloned value of A is incorrect: got %d, want 100", *clonedStruct.A)
    }
}

func TestCloneNestedPointers(t *testing.T) {
    cm := cloner.NewCloneManager()

    // Test 2: Nested Structs with pointers
    a := 200
    original := struct {
        Inner *struct {
            X *int
            Y *int
        }
    }{
        Inner: &struct {
            X *int
            Y *int
        }{
            X: &a,
            Y: &a,
        },
    }

    cloned, err := cm.Clone(original)
    if err != nil {
        t.Fatalf("Clone failed: %v", err)
    }

    clonedStruct := cloned.(struct {
        Inner *struct {
            X *int
            Y *int
        }
    })

    // Ensure that both pointers in the cloned nested struct point to the same value
    if clonedStruct.Inner.X != clonedStruct.Inner.Y {
        t.Fatalf("Cloned nested pointers X and Y do not point to the same value")
    }

    // Ensure the value they point to is correct
    if *clonedStruct.Inner.X != 200 {
        t.Errorf("Cloned nested value of X is incorrect: got %d, want 200", *clonedStruct.Inner.X)
    }
}

func TestCloneSliceOfPointers(t *testing.T) {
    cm := cloner.NewCloneManager()

    // Test 3: Slice containing pointers
    a := 300
    b := 400
    original := struct {
        Values []*int
    }{
        Values: []*int{&a, &b, &a}, // Third element points to the same value as first
    }

    cloned, err := cm.Clone(original)
    if err != nil {
        t.Fatalf("Clone failed: %v", err)
    }

    clonedStruct := cloned.(struct {
        Values []*int
    })

    // Ensure that the first and third pointers in the cloned slice point to the same value
    if clonedStruct.Values[0] != clonedStruct.Values[2] {
        t.Fatalf("Cloned slice pointers do not point to the same value")
    }

    // Ensure the values are correct
    if *clonedStruct.Values[0] != 300 {
        t.Errorf("Cloned slice value is incorrect: got %d, want 300", *clonedStruct.Values[0])
    }

    if *clonedStruct.Values[1] != 400 {
        t.Errorf("Cloned slice value is incorrect: got %d, want 400", *clonedStruct.Values[1])
    }
}
