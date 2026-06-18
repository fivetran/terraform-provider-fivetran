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

func IsGenerallyAvailableMetadataField(prop *metadata.Property) bool {
	status := MetadataFieldStatus(prop)
	return status == "" || status == FieldStatusGeneralAvailability
}

func IsTerraformSupportedMetadataField(prop *metadata.Property) bool {
	switch MetadataFieldStatus(prop) {
	case "", FieldStatusGeneralAvailability, FieldStatusSunset:
		return true
	case FieldStatusPrivatePreview, FieldStatusDevelopment:
		return false
	default:
		return false
	}
}

func IsPrivatePreviewMetadataField(prop *metadata.Property) bool {
	return MetadataFieldStatus(prop) == FieldStatusPrivatePreview
}

func IsDevelopmentMetadataField(prop *metadata.Property) bool {
	return MetadataFieldStatus(prop) == FieldStatusDevelopment
}

func IsSunsetMetadataField(prop *metadata.Property) bool {
	return MetadataFieldStatus(prop) == FieldStatusSunset
}
