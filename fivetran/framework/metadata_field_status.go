package framework

import (
	"github.com/fivetran/go-fivetran/metadata"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
)

const (
	FieldStatusDevelopment         = core.FieldStatusDevelopment
	FieldStatusPrivatePreview      = core.FieldStatusPrivatePreview
	FieldStatusGeneralAvailability = core.FieldStatusGeneralAvailability
	FieldStatusSunset              = core.FieldStatusSunset
)

func MetadataFieldStatus(prop *metadata.Property) string {
	return core.MetadataFieldStatus(prop)
}

func IsKnownMetadataFieldStatus(status string) bool {
	return core.IsKnownMetadataFieldStatus(status)
}

func ShouldWarnForMetadataFieldStatus(prop *metadata.Property) bool {
	return core.ShouldWarnForMetadataFieldStatus(prop)
}
