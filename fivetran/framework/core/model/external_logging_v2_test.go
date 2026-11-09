package model

import (
	"reflect"
	"testing"

	"github.com/fivetran/go-fivetran/metadata"
)

func TestExternalLoggingV2ImportConfigMask(t *testing.T) {
	config := map[string]interface{}{
		"api_key":     "******",
		"token":       "plain",
		"site":        "datadoghq.com",
		"external_id": "ext",
	}
	meta := &metadata.LogServiceMetadata{Config: metadata.Property{
		Properties: map[string]*metadata.Property{
			"api_key":     {Type: "string", Format: "password"},
			"token":       {Type: "string", Format: "password"},
			"site":        {Type: "string"},
			"external_id": {Type: "string", Readonly: true},
		},
	}}

	tests := []struct {
		name string
		meta *metadata.LogServiceMetadata
		want map[string]interface{}
	}{
		{
			name: "drops masked, password and readonly fields",
			meta: meta,
			want: map[string]interface{}{"site": "datadoghq.com"},
		},
		{
			name: "without metadata only masked values are dropped",
			want: map[string]interface{}{"token": "plain", "site": "datadoghq.com", "external_id": "ext"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExternalLoggingV2ImportConfigMask(config, tt.meta); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mask = %v, want %v", got, tt.want)
			}
		})
	}
	if len(config) != 4 {
		t.Errorf("input config must not be modified, got %v", config)
	}
}
