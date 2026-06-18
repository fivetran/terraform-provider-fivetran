package framework

import (
	"testing"

	"github.com/fivetran/go-fivetran/metadata"
)

func TestMetadataFieldStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		prop *metadata.Property
		want string
	}{
		{
			name: "nil property",
			prop: nil,
			want: "",
		},
		{
			name: "missing status",
			prop: &metadata.Property{},
			want: "",
		},
		{
			name: "private preview",
			prop: &metadata.Property{FieldStatus: FieldStatusPrivatePreview},
			want: FieldStatusPrivatePreview,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := MetadataFieldStatus(tt.prop); got != tt.want {
				t.Fatalf("MetadataFieldStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMetadataFieldStatusHelpers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		prop      *metadata.Property
		wantKnown bool
		wantWarn  bool
	}{
		{
			name:      "missing status is treated as known for backward compatibility",
			prop:      &metadata.Property{},
			wantKnown: true,
		},
		{
			name:      "general availability",
			prop:      &metadata.Property{FieldStatus: FieldStatusGeneralAvailability},
			wantKnown: true,
		},
		{
			name:      "private preview",
			prop:      &metadata.Property{FieldStatus: FieldStatusPrivatePreview},
			wantKnown: true,
			wantWarn:  true,
		},
		{
			name:      "development",
			prop:      &metadata.Property{FieldStatus: FieldStatusDevelopment},
			wantKnown: true,
			wantWarn:  true,
		},
		{
			name:      "sunset",
			prop:      &metadata.Property{FieldStatus: FieldStatusSunset},
			wantKnown: true,
			wantWarn:  true,
		},
		{
			name:      "unknown status is not known",
			prop:      &metadata.Property{FieldStatus: "some_future_status"},
			wantKnown: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := IsKnownMetadataFieldStatus(MetadataFieldStatus(tt.prop)); got != tt.wantKnown {
				t.Errorf("IsKnownMetadataFieldStatus() = %v, want %v", got, tt.wantKnown)
			}
			if got := ShouldWarnForMetadataFieldStatus(tt.prop); got != tt.wantWarn {
				t.Errorf("ShouldWarnForMetadataFieldStatus() = %v, want %v", got, tt.wantWarn)
			}
		})
	}
}
