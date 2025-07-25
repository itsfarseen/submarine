package parser_test

import (
	"strings"
	. "submarine/codegen/parser"
	"testing"
)

func TestParseSimpleString(t *testing.T) {
	input := "hello world"
	reader := strings.NewReader(input)

	result, err := ParseYAML(reader)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result != "hello world" {
		t.Errorf("Expected 'hello world', got %v", result)
	}
}

func TestParseSimpleObject(t *testing.T) {
	input := `name: John
age: 30`
	reader := strings.NewReader(input)

	result, err := ParseYAML(reader)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	obj, ok := result.(*YamlObject)
	if !ok {
		t.Fatalf("Expected YamlObject, got %T", result)
	}

	if len(obj.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(obj.Fields))
	}

	if obj.Fields[0].Key != "name" || obj.Fields[0].Value != "John" {
		t.Errorf("Expected name: John, got %s: %v", obj.Fields[0].Key, obj.Fields[0].Value)
	}

	if obj.Fields[1].Key != "age" || obj.Fields[1].Value != "30" {
		t.Errorf("Expected age: 30, got %s: %v", obj.Fields[1].Key, obj.Fields[1].Value)
	}
}

func TestParseSimpleArray(t *testing.T) {
	input := `- apple
- banana
- cherry`
	reader := strings.NewReader(input)

	result, err := ParseYAML(reader)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	arr, ok := result.(*YamlArray)
	if !ok {
		t.Fatalf("Expected YamlArray, got %T", result)
	}

	if len(arr.Items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(arr.Items))
	}

	expected := []string{"apple", "banana", "cherry"}
	for i, item := range arr.Items {
		if item != expected[i] {
			t.Errorf("Expected %s at index %d, got %v", expected[i], i, item)
		}
	}
}

func TestParseNestedObject(t *testing.T) {
	input := `person:
  name: John
  age: 30
  address:
    street: 123 Main St
    city: Anytown`
	reader := strings.NewReader(input)

	result, err := ParseYAML(reader)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	obj, ok := result.(*YamlObject)
	if !ok {
		t.Fatalf("Expected YamlObject, got %T", result)
	}

	if len(obj.Fields) != 1 {
		t.Errorf("Expected 1 field, got %d", len(obj.Fields))
	}

	personField := obj.Fields[0]
	if personField.Key != "person" {
		t.Errorf("Expected key 'person', got %s", personField.Key)
	}

	personObj, ok := personField.Value.(*YamlObject)
	if !ok {
		t.Fatalf("Expected nested YamlObject, got %T", personField.Value)
	}

	if len(personObj.Fields) != 3 {
		t.Errorf("Expected 3 nested fields, got %d", len(personObj.Fields))
	}
}

func TestParseArrayOfObjects(t *testing.T) {
	input := `- name: John
  age: 30
- name: Jane
  age: 25`
	reader := strings.NewReader(input)

	result, err := ParseYAML(reader)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	arr, ok := result.(*YamlArray)
	if !ok {
		t.Fatalf("Expected YamlArray, got %T", result)
	}

	if len(arr.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(arr.Items))
	}

	// Check first object
	obj1, ok := arr.Items[0].(*YamlObject)
	if !ok {
		t.Fatalf("Expected first item to be YamlObject, got %T", arr.Items[0])
	}

	if len(obj1.Fields) != 2 {
		t.Errorf("Expected 2 fields in first object, got %d", len(obj1.Fields))
	}

	if obj1.Fields[0].Key != "name" || obj1.Fields[0].Value != "John" {
		t.Errorf("Expected name: John in first object, got %s: %v", obj1.Fields[0].Key, obj1.Fields[0].Value)
	}
}

func TestParseObjectWithArray(t *testing.T) {
	input := `fruits:
  - apple
  - banana
  - cherry
vegetables:
  - carrot
  - lettuce`
	reader := strings.NewReader(input)

	result, err := ParseYAML(reader)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	obj, ok := result.(*YamlObject)
	if !ok {
		t.Fatalf("Expected YamlObject, got %T", result)
	}

	if len(obj.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(obj.Fields))
	}

	// Check fruits array
	fruitsField := obj.Fields[0]
	if fruitsField.Key != "fruits" {
		t.Errorf("Expected key 'fruits', got %s", fruitsField.Key)
	}

	fruitsArr, ok := fruitsField.Value.(*YamlArray)
	if !ok {
		t.Fatalf("Expected fruits value to be YamlArray, got %T", fruitsField.Value)
	}

	if len(fruitsArr.Items) != 3 {
		t.Errorf("Expected 3 fruits, got %d", len(fruitsArr.Items))
	}
}

func TestParseInlineArrayItem(t *testing.T) {
	input := `- name: John
- age: 30`
	reader := strings.NewReader(input)

	result, err := ParseYAML(reader)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	arr, ok := result.(*YamlArray)
	if !ok {
		t.Fatalf("Expected YamlArray, got %T", result)
	}

	if len(arr.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(arr.Items))
	}

	// Check first inline object
	obj1, ok := arr.Items[0].(*YamlObject)
	if !ok {
		t.Fatalf("Expected first item to be YamlObject, got %T", arr.Items[0])
	}

	if len(obj1.Fields) != 1 {
		t.Errorf("Expected 1 field in first object, got %d", len(obj1.Fields))
	}

	if obj1.Fields[0].Key != "name" || obj1.Fields[0].Value != "John" {
		t.Errorf("Expected name: John, got %s: %v", obj1.Fields[0].Key, obj1.Fields[0].Value)
	}
}

func TestParseEmptyInput(t *testing.T) {
	input := ""
	reader := strings.NewReader(input)

	result, err := ParseYAML(reader)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil result for empty input, got %v", result)
	}
}

func TestParseWithComments(t *testing.T) {
	input := `# This is a comment
name: John  # inline comment
# Another comment
age: 30`
	reader := strings.NewReader(input)

	result, err := ParseYAML(reader)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	obj, ok := result.(*YamlObject)
	if !ok {
		t.Fatalf("Expected YamlObject, got %T", result)
	}

	if len(obj.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(obj.Fields))
	}
}

func TestCountIndent(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"no indent", 0},
		{"  two spaces", 2},
		{"    four spaces", 4},
		{"\ttab should not count", 0},
		{"  spaces then text", 2},
	}

	for _, test := range tests {
		result := CountIndent(test.input)
		if result != test.expected {
			t.Errorf("CountIndent(%q) = %d, expected %d", test.input, result, test.expected)
		}
	}
}
