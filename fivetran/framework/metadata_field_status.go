package framework

import "github.com/fivetran/go-fivetran/metadata"

const (
	FieldStatusDevelopment         = "development"
	FieldStatusPrivatePreview      = "private_preview"
	FieldStatusGeneralAvailability = "general_availability"
	FieldStatusSunset              = "sunset"
)

func MetadataFieldStatus(prop *metadata.Property) string {
	if prop == nil {
		return ""
	}
	return prop.FieldStatus
}

func IsKnownMetadataFieldStatus(status string) bool {
	switch status {
	case "", FieldStatusGeneralAvailability, FieldStatusPrivatePreview, FieldStatusDevelopment, FieldStatusSunset:
		return true
	default:
		return false
	}
}

func ShouldWarnForMetadataFieldStatus(prop *metadata.Property) bool {
	switch MetadataFieldStatus(prop) {
	case FieldStatusPrivatePreview, FieldStatusDevelopment, FieldStatusSunset:
		return true
	default:
		return false
	}
}
