package fivetrantypes

import (
	"testing"
)

func TestJsonEqual(t *testing.T) {
	tests := []struct {
		name     string
		json1    string
		json2    string
		expected bool
		wantErr  bool
	}{
		{
			name:     "identical simple JSON",
			json1:    `{"key": "value"}`,
			json2:    `{"key": "value"}`,
			expected: true,
			wantErr:  false,
		},
		{
			name:     "same JSON different whitespace",
			json1:    `{"key":"value"}`,
			json2:    `{"key": "value"}`,
			expected: true,
			wantErr:  false,
		},
		{
			name:     "same JSON different key order",
			json1:    `{"a": 1, "b": 2}`,
			json2:    `{"b": 2, "a": 1}`,
			expected: true,
			wantErr:  false,
		},
		{
			name:     "different values",
			json1:    `{"key": "value1"}`,
			json2:    `{"key": "value2"}`,
			expected: false,
			wantErr:  false,
		},
		{
			name:     "different keys",
			json1:    `{"key1": "value"}`,
			json2:    `{"key2": "value"}`,
			expected: false,
			wantErr:  false,
		},
		{
			name:     "nested objects equal",
			json1:    `{"outer": {"inner": "value"}}`,
			json2:    `{"outer": {"inner": "value"}}`,
			expected: true,
			wantErr:  false,
		},
		{
			name:     "nested objects different order",
			json1:    `{"a": {"x": 1, "y": 2}, "b": 3}`,
			json2:    `{"b": 3, "a": {"y": 2, "x": 1}}`,
			expected: true,
			wantErr:  false,
		},
		{
			name:     "arrays equal",
			json1:    `{"items": [1, 2, 3]}`,
			json2:    `{"items": [1, 2, 3]}`,
			expected: true,
			wantErr:  false,
		},
		{
			name:     "arrays different order",
			json1:    `{"items": [1, 2, 3]}`,
			json2:    `{"items": [3, 2, 1]}`,
			expected: false,
			wantErr:  false,
		},
		{
			name:     "numbers preserved",
			json1:    `{"port": 5432, "timeout": 30.5}`,
			json2:    `{"timeout": 30.5, "port": 5432}`,
			expected: true,
			wantErr:  false,
		},
		{
			name:     "booleans",
			json1:    `{"enabled": true, "disabled": false}`,
			json2:    `{"disabled": false, "enabled": true}`,
			expected: true,
			wantErr:  false,
		},
		{
			name:     "null values",
			json1:    `{"key": null}`,
			json2:    `{"key": null}`,
			expected: true,
			wantErr:  false,
		},
		{
			name:     "empty objects",
			json1:    `{}`,
			json2:    `{}`,
			expected: true,
			wantErr:  false,
		},
		{
			name:     "empty arrays",
			json1:    `{"items": []}`,
			json2:    `{"items": []}`,
			expected: true,
			wantErr:  false,
		},
		{
			name:     "invalid JSON in first string",
			json1:    `{invalid}`,
			json2:    `{"key": "value"}`,
			expected: false,
			wantErr:  true,
		},
		{
			name:     "invalid JSON in second string",
			json1:    `{"key": "value"}`,
			json2:    `{invalid}`,
			expected: false,
			wantErr:  true,
		},
		{
			name:     "unclosed brace",
			json1:    `{"key": "value"`,
			json2:    `{"key": "value"}`,
			expected: false,
			wantErr:  true,
		},
		{
			name:     "trailing comma",
			json1:    `{"key": "value",}`,
			json2:    `{"key": "value"}`,
			expected: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := jsonEqual(tt.json1, tt.json2)

			if tt.wantErr {
				if err == nil {
					t.Error("jsonEqual() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("jsonEqual() unexpected error: %v", err)
				}
			}

			if result != tt.expected {
				t.Errorf("jsonEqual() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestNormalizeJSONString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "simple object",
			input:    `{"key": "value"}`,
			expected: `{"key":"value"}`,
			wantErr:  false,
		},
		{
			name:     "whitespace removed",
			input:    `  {  "key"  :  "value"  }  `,
			expected: `{"key":"value"}`,
			wantErr:  false,
		},
		{
			name:     "keys sorted alphabetically",
			input:    `{"z": 1, "a": 2, "m": 3}`,
			expected: `{"a":2,"m":3,"z":1}`,
			wantErr:  false,
		},
		{
			name:     "nested object keys sorted",
			input:    `{"outer": {"z": 1, "a": 2}}`,
			expected: `{"outer":{"a":2,"z":1}}`,
			wantErr:  false,
		},
		{
			name:     "numbers preserved",
			input:    `{"int": 42, "float": 3.14}`,
			expected: `{"float":3.14,"int":42}`,
			wantErr:  false,
		},
		{
			name:     "booleans preserved",
			input:    `{"true": true, "false": false}`,
			expected: `{"false":false,"true":true}`,
			wantErr:  false,
		},
		{
			name:     "null preserved",
			input:    `{"value": null}`,
			expected: `{"value":null}`,
			wantErr:  false,
		},
		{
			name:     "array preserved in order",
			input:    `{"items": [3, 1, 2]}`,
			expected: `{"items":[3,1,2]}`,
			wantErr:  false,
		},
		{
			name:     "empty object",
			input:    `{}`,
			expected: `{}`,
			wantErr:  false,
		},
		{
			name:     "empty array",
			input:    `[]`,
			expected: `[]`,
			wantErr:  false,
		},
		{
			name:     "invalid JSON",
			input:    `{invalid}`,
			expected: ``,
			wantErr:  true,
		},
		{
			name:     "unclosed brace",
			input:    `{"key": "value"`,
			expected: ``,
			wantErr:  true,
		},
		{
			name:     "trailing comma",
			input:    `{"key": "value",}`,
			expected: ``,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := normalizeJSONString(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("normalizeJSONString() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("normalizeJSONString() unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("normalizeJSONString() = %q, expected %q", result, tt.expected)
				}
			}
		})
	}
}

func TestNormalizeJSONString_ComplexConnectorConfig(t *testing.T) {
	input := `{
		"host": "db.example.com",
		"port": 5432,
		"database": "production",
		"user": "fivetran_user",
		"update_method": "QUERY_BASED"
	}`

	result, err := normalizeJSONString(input)
	if err != nil {
		t.Fatalf("normalizeJSONString() unexpected error: %v", err)
	}

	result2, err := normalizeJSONString(result)
	if err != nil {
		t.Fatalf("normalizeJSONString() second pass failed: %v", err)
	}

	if result != result2 {
		t.Errorf("normalizeJSONString() is not idempotent: %q != %q", result, result2)
	}
}
