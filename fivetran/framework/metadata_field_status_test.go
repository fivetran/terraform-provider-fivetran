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
		name        string
		prop        *metadata.Property
		wantGA      bool
		wantSupport bool
		wantWarn    bool
		wantPrivate bool
		wantDevelop bool
		wantSunset  bool
	}{
		{
			name:        "missing status is treated as supported for backward compatibility",
			prop:        &metadata.Property{},
			wantGA:      true,
			wantSupport: true,
		},
		{
			name:        "general availability",
			prop:        &metadata.Property{FieldStatus: FieldStatusGeneralAvailability},
			wantGA:      true,
			wantSupport: true,
		},
		{
			name:        "private preview",
			prop:        &metadata.Property{FieldStatus: FieldStatusPrivatePreview},
			wantSupport: true,
			wantWarn:    true,
			wantPrivate: true,
		},
		{
			name:        "development",
			prop:        &metadata.Property{FieldStatus: FieldStatusDevelopment},
			wantSupport: true,
			wantWarn:    true,
			wantDevelop: true,
		},
		{
			name:        "sunset",
			prop:        &metadata.Property{FieldStatus: FieldStatusSunset},
			wantSupport: true,
			wantWarn:    true,
			wantSunset:  true,
		},
		{
			name:        "unknown status is unsupported",
			prop:        &metadata.Property{FieldStatus: "some_future_status"},
			wantSupport: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := IsGenerallyAvailableMetadataField(tt.prop); got != tt.wantGA {
				t.Errorf("IsGenerallyAvailableMetadataField() = %v, want %v", got, tt.wantGA)
			}
			if got := IsTerraformSupportedMetadataField(tt.prop); got != tt.wantSupport {
				t.Errorf("IsTerraformSupportedMetadataField() = %v, want %v", got, tt.wantSupport)
			}
			if got := ShouldWarnForMetadataFieldStatus(tt.prop); got != tt.wantWarn {
				t.Errorf("ShouldWarnForMetadataFieldStatus() = %v, want %v", got, tt.wantWarn)
			}
			if got := IsPrivatePreviewMetadataField(tt.prop); got != tt.wantPrivate {
				t.Errorf("IsPrivatePreviewMetadataField() = %v, want %v", got, tt.wantPrivate)
			}
			if got := IsDevelopmentMetadataField(tt.prop); got != tt.wantDevelop {
				t.Errorf("IsDevelopmentMetadataField() = %v, want %v", got, tt.wantDevelop)
			}
			if got := IsSunsetMetadataField(tt.prop); got != tt.wantSunset {
				t.Errorf("IsSunsetMetadataField() = %v, want %v", got, tt.wantSunset)
			}
		})
	}
}
