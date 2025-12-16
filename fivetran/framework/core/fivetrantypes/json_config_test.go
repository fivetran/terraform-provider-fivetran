package fivetrantypes

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestJsonConfigValue_Equal(t *testing.T) {
	tests := []struct {
		name     string
		value1   JsonConfigValue
		value2   JsonConfigValue
		expected bool
	}{
		{
			name:     "equal JSON strings",
			value1:   NewJsonConfigValue(`{"key": "value"}`),
			value2:   NewJsonConfigValue(`{"key": "value"}`),
			expected: true,
		},
		{
			name:     "different JSON strings",
			value1:   NewJsonConfigValue(`{"key": "value1"}`),
			value2:   NewJsonConfigValue(`{"key": "value2"}`),
			expected: false,
		},
		{
			name:     "both null",
			value1:   NewJsonConfigNull(),
			value2:   NewJsonConfigNull(),
			expected: true,
		},
		{
			name:     "both unknown",
			value1:   NewJsonConfigUnknown(),
			value2:   NewJsonConfigUnknown(),
			expected: true,
		},
		{
			name:     "null vs unknown",
			value1:   NewJsonConfigNull(),
			value2:   NewJsonConfigUnknown(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.value1.Equal(tt.value2)
			if result != tt.expected {
				t.Errorf("Equal() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestJsonConfigValue_StringSemanticEquals(t *testing.T) {
	tests := []struct {
		name     string
		value1   JsonConfigValue
		value2   JsonConfigValue
		expected bool
		wantErr  bool
	}{
		{
			name:     "identical JSON",
			value1:   NewJsonConfigValue(`{"key": "value"}`),
			value2:   NewJsonConfigValue(`{"key": "value"}`),
			expected: true,
			wantErr:  false,
		},
		{
			name:     "same JSON different whitespace",
			value1:   NewJsonConfigValue(`{"key":"value"}`),
			value2:   NewJsonConfigValue(`{"key": "value"}`),
			expected: true,
			wantErr:  false,
		},
		{
			name:     "same JSON different key order",
			value1:   NewJsonConfigValue(`{"a": 1, "b": 2}`),
			value2:   NewJsonConfigValue(`{"b": 2, "a": 1}`),
			expected: true,
			wantErr:  false,
		},
		{
			name:     "different JSON values",
			value1:   NewJsonConfigValue(`{"key": "value1"}`),
			value2:   NewJsonConfigValue(`{"key": "value2"}`),
			expected: false,
			wantErr:  false,
		},
		{
			name:     "invalid JSON in first value",
			value1:   NewJsonConfigValue(`{invalid json}`),
			value2:   NewJsonConfigValue(`{"key": "value"}`),
			expected: false,
			wantErr:  true,
		},
		{
			name:     "invalid JSON in second value",
			value1:   NewJsonConfigValue(`{"key": "value"}`),
			value2:   NewJsonConfigValue(`{invalid json}`),
			expected: false,
			wantErr:  true,
		},
		{
			name:     "numbers preserved",
			value1:   NewJsonConfigValue(`{"port": 5432}`),
			value2:   NewJsonConfigValue(`{"port": 5432}`),
			expected: true,
			wantErr:  false,
		},
		{
			name:     "nested objects equal",
			value1:   NewJsonConfigValue(`{"config": {"host": "localhost", "port": 5432}}`),
			value2:   NewJsonConfigValue(`{"config": {"port": 5432, "host": "localhost"}}`),
			expected: true,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, diags := tt.value1.StringSemanticEquals(context.Background(), tt.value2)

			if tt.wantErr {
				if !diags.HasError() {
					t.Errorf("StringSemanticEquals() expected error but got none")
				}
			} else {
				if diags.HasError() {
					t.Errorf("StringSemanticEquals() unexpected error: %v", diags)
				}
			}

			if result != tt.expected {
				t.Errorf("StringSemanticEquals() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestJsonConfigValue_StringSemanticEquals_WrongType(t *testing.T) {
	value1 := NewJsonConfigValue(`{"key": "value"}`)
	value2 := basetypes.NewStringValue(`{"key": "value"}`) // Wrong type

	result, diags := value1.StringSemanticEquals(context.Background(), value2)

	if !diags.HasError() {
		t.Error("Expected error for wrong type, got none")
	}

	if result != false {
		t.Errorf("Expected false for wrong type, got %v", result)
	}
}

func TestJsonConfigValue_Unmarshal(t *testing.T) {
	tests := []struct {
		name      string
		value     JsonConfigValue
		target    interface{}
		wantErr   bool
		expectNil bool
	}{
		{
			name:    "valid JSON to map",
			value:   NewJsonConfigValue(`{"key": "value", "port": 5432}`),
			target:  &map[string]interface{}{},
			wantErr: false,
		},
		{
			name:    "valid JSON to struct",
			value:   NewJsonConfigValue(`{"host": "localhost", "port": 5432}`),
			target:  &struct{ Host string; Port int }{},
			wantErr: false,
		},
		{
			name:      "null value",
			value:     NewJsonConfigNull(),
			target:    &map[string]interface{}{},
			wantErr:   true,
			expectNil: true,
		},
		{
			name:      "unknown value",
			value:     NewJsonConfigUnknown(),
			target:    &map[string]interface{}{},
			wantErr:   true,
			expectNil: true,
		},
		{
			name:    "invalid JSON",
			value:   NewJsonConfigValue(`{invalid json}`),
			target:  &map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := tt.value.Unmarshal(tt.target)

			if tt.wantErr {
				if !diags.HasError() {
					t.Error("Unmarshal() expected error but got none")
				}
			} else {
				if diags.HasError() {
					t.Errorf("Unmarshal() unexpected error: %v", diags)
				}

				// For valid cases, check that target was populated
				if !tt.expectNil {
					switch v := tt.target.(type) {
					case *map[string]interface{}:
						if len(*v) == 0 {
							t.Error("Unmarshal() target map is empty")
						}
					}
				}
			}
		})
	}
}

func TestJsonConfigValue_Constructors(t *testing.T) {
	t.Run("NewJsonConfigValue", func(t *testing.T) {
		value := NewJsonConfigValue(`{"key": "value"}`)
		if value.IsNull() {
			t.Error("NewJsonConfigValue() created null value")
		}
		if value.IsUnknown() {
			t.Error("NewJsonConfigValue() created unknown value")
		}
		if value.ValueString() != `{"key": "value"}` {
			t.Errorf("NewJsonConfigValue() value = %v, expected %v", value.ValueString(), `{"key": "value"}`)
		}
	})

	t.Run("NewJsonConfigNull", func(t *testing.T) {
		value := NewJsonConfigNull()
		if !value.IsNull() {
			t.Error("NewJsonConfigNull() did not create null value")
		}
	})

	t.Run("NewJsonConfigUnknown", func(t *testing.T) {
		value := NewJsonConfigUnknown()
		if !value.IsUnknown() {
			t.Error("NewJsonConfigUnknown() did not create unknown value")
		}
	})

	t.Run("NewJsonConfigPointerValue with non-nil", func(t *testing.T) {
		str := `{"key": "value"}`
		value := NewJsonConfigPointerValue(&str)
		if value.IsNull() {
			t.Error("NewJsonConfigPointerValue() created null value for non-nil pointer")
		}
		if value.ValueString() != str {
			t.Errorf("NewJsonConfigPointerValue() value = %v, expected %v", value.ValueString(), str)
		}
	})

	t.Run("NewJsonConfigPointerValue with nil", func(t *testing.T) {
		value := NewJsonConfigPointerValue(nil)
		if !value.IsNull() {
			t.Error("NewJsonConfigPointerValue() did not create null value for nil pointer")
		}
	})
}

func TestJsonConfigValue_Type(t *testing.T) {
	value := NewJsonConfigValue(`{"key": "value"}`)
	attrType := value.Type(context.Background())

	if _, ok := attrType.(JsonConfigType); !ok {
		t.Errorf("Type() returned %T, expected JsonConfigType", attrType)
	}
}
