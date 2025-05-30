package main

import (
	"reflect"
	"testing"
)

func TestMultiFlagSet(t *testing.T) {
	var m multiFlag

	err := m.Set("chrome firefox")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := multiFlag{"chrome", "firefox"}
	if !reflect.DeepEqual(m, expected) {
		t.Errorf("Expected %v, got %v", expected, m)
	}

	err = m.Set("webkit")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected = multiFlag{"chrome", "firefox", "webkit"}
	if !reflect.DeepEqual(m, expected) {
		t.Errorf("Expected %v after second set, got %v", expected, m)
	}
}

func TestMultiFlagString(t *testing.T) {
	m := multiFlag{"a", "b", "c"}
	result := m.String()
	expected := "a,b,c"
	if result != expected {
		t.Errorf("Expected string %q, got %q", expected, result)
	}
}
