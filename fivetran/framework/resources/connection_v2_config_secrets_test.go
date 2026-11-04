package resources

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/fivetran/go-fivetran/metadata"
)

func TestConnectionV2ConfigSecretsPatchFiltersNestedFields(t *testing.T) {
	t.Parallel()

	credentials := &metadata.Property{Type: "object", Properties: map[string]*metadata.Property{
		"client_id": {Type: "string", Immutable: true},
		"status":    {Type: "string", Readonly: true},
		"token":     {Type: "string"},
	}}
	for _, tc := range []struct {
		name    string
		value   interface{}
		prop    *metadata.Property
		want    interface{}
		omitted bool
	}{
		{
			name:  "nested object",
			value: map[string]interface{}{"client_id": "client", "status": "ready", "token": "rotated"},
			prop:  credentials,
			want:  map[string]interface{}{"token": "rotated"},
		},
		{
			name:  "array preserves item positions",
			value: []interface{}{map[string]interface{}{"client_id": "first"}, map[string]interface{}{"client_id": "second", "status": "ready", "token": "rotated"}},
			prop:  &metadata.Property{Type: "array", Items: credentials},
			want:  []interface{}{map[string]interface{}{}, map[string]interface{}{"token": "rotated"}},
		},
		{
			name:  "object inside nested array",
			value: map[string]interface{}{"accounts": []interface{}{map[string]interface{}{"client_id": "client", "token": "rotated"}}},
			prop:  &metadata.Property{Type: "object", Properties: map[string]*metadata.Property{"accounts": {Type: "array", Items: credentials}}},
			want:  map[string]interface{}{"accounts": []interface{}{map[string]interface{}{"token": "rotated"}}},
		},
		{
			name:    "object with only excluded fields",
			value:   map[string]interface{}{"client_id": "client", "status": "ready"},
			prop:    credentials,
			omitted: true,
		},
		{
			name:    "array with only excluded fields",
			value:   []interface{}{map[string]interface{}{"client_id": "client"}},
			prop:    &metadata.Property{Type: "array", Items: credentials},
			omitted: true,
		},
		{
			name:    "immutable parent excludes mutable children",
			value:   map[string]interface{}{"token": "rotated"},
			prop:    &metadata.Property{Type: "object", Immutable: true, Properties: credentials.Properties},
			omitted: true,
		},
		{
			name:    "readonly parent excludes mutable children",
			value:   map[string]interface{}{"token": "rotated"},
			prop:    &metadata.Property{Type: "object", Readonly: true, Properties: credentials.Properties},
			omitted: true,
		},
		{name: "explicit null", value: nil, prop: credentials, want: nil},
		{name: "explicit empty object", value: map[string]interface{}{}, prop: credentials, want: map[string]interface{}{}},
		{name: "explicit empty array", value: []interface{}{}, prop: &metadata.Property{Type: "array", Items: credentials}, want: []interface{}{}},
		{name: "empty string", value: "", prop: &metadata.Property{Type: "string"}, want: ""},
		{name: "missing metadata", value: map[string]interface{}{"token": "rotated"}, want: map[string]interface{}{"token": "rotated"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			for _, versionChanged := range []bool{false, true} {
				secrets := map[string]interface{}{"credentials": tc.value}
				before, err := json.Marshal(secrets)
				if err != nil {
					t.Fatal(err)
				}
				var patch, stateConfig map[string]interface{}
				if !versionChanged {
					// Migration can start with a null PATCH that would clear the removed field.
					patch = map[string]interface{}{"credentials": nil}
					stateConfig = map[string]interface{}{"credentials": tc.value}
				}
				slot := &metadata.Property{Properties: map[string]*metadata.Property{"credentials": tc.prop}}
				patch = applyConnectionV2ConfigSecretsToPatch(patch, secrets, stateConfig, slot, versionChanged)
				want := map[string]interface{}{}
				if !tc.omitted {
					want["credentials"] = tc.want
				}
				if len(patch) != len(want) || !reflect.DeepEqual(patch["credentials"], want["credentials"]) {
					t.Errorf("versionChanged=%v: patch = %#v, want %#v", versionChanged, patch, want)
				}
				after, err := json.Marshal(secrets)
				if err != nil {
					t.Fatal(err)
				}
				if string(before) != string(after) {
					t.Error("filtering modified the input secrets")
				}
			}
		})
	}
}
