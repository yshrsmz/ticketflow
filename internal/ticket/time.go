package ticket

import (
	"time"

	"gopkg.in/yaml.v3"
)

// RFC3339Time is a wrapper around time.Time that marshals to RFC3339 format without subseconds
type RFC3339Time struct {
	time.Time
}

// MarshalYAML implements yaml.Marshaler interface
func (t RFC3339Time) MarshalYAML() (interface{}, error) {
	if t.IsZero() {
		return nil, nil
	}
	return t.Format(time.RFC3339), nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface
func (t *RFC3339Time) UnmarshalYAML(node *yaml.Node) error {
	if node.Value == "" || node.Value == "null" {
		*t = RFC3339Time{}
		return nil
	}

	parsed, err := time.Parse(time.RFC3339, node.Value)
	if err != nil {
		// Try parsing with nanoseconds for backward compatibility
		parsed, err = time.Parse(time.RFC3339Nano, node.Value)
		if err != nil {
			return err
		}
	}

	*t = RFC3339Time{parsed}
	return nil
}

// NewRFC3339Time creates a new RFC3339Time from time.Time
func NewRFC3339Time(t time.Time) RFC3339Time {
	return RFC3339Time{t}
}

// ToTimePtr converts RFC3339Time to *time.Time.
// This method is useful when you need a pointer to a time.Time object
// for compatibility with APIs or functions that require nullable time values.
func (t RFC3339Time) ToTimePtr() *time.Time {
	if t.IsZero() {
		return nil
	}
	tt := t.Time
	return &tt
}

// RFC3339TimePtr represents a nullable RFC3339Time
type RFC3339TimePtr struct {
	Time *time.Time
}

// MarshalYAML implements yaml.Marshaler interface
func (t RFC3339TimePtr) MarshalYAML() (interface{}, error) {
	if t.Time == nil || t.Time.IsZero() {
		return nil, nil
	}
	return t.Time.Format(time.RFC3339), nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface
func (t *RFC3339TimePtr) UnmarshalYAML(node *yaml.Node) error {
	if node.Value == "" || node.Value == "null" {
		t.Time = nil
		return nil
	}

	parsed, err := time.Parse(time.RFC3339, node.Value)
	if err != nil {
		// Try parsing with nanoseconds for backward compatibility
		parsed, err = time.Parse(time.RFC3339Nano, node.Value)
		if err != nil {
			return err
		}
	}

	t.Time = &parsed
	return nil
}

// NewRFC3339TimePtr creates a new RFC3339TimePtr from *time.Time
func NewRFC3339TimePtr(t *time.Time) RFC3339TimePtr {
	return RFC3339TimePtr{Time: t}
}

