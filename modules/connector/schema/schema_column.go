package schema

import (
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/connectors"
	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type _column struct {
	_element
	hashed *bool
}

func (c *_column) setHashed(value *bool) {
	if value != nil && (c.hashed == nil || *value != *c.hashed) {
		c.hashed = value
		c.updated = true
	} else {
		c.hashed = nil
	}
}

func (c *_column) setHashedToDefault() {
	hashedDefault := false
	c.setHashed(&hashedDefault)
}

func (c _column) prepareRequest() *connectors.ConnectorSchemaConfigColumn {
	result := fivetran.NewConnectorSchemaConfigColumn()
	if c.enabledPatched && c.isPatchAllowed() {
		result.Enabled(c.enabled)
	}
	if c.hashed != nil {
		result.Hashed(*c.hashed)
	}
	if c.isPrimaryKey != nil {
		result.IsPrimaryKey(*c.isPrimaryKey)
	}
	return result
}

func (c _column) prepareCreateRequest() *connectors.ConnectorSchemaConfigColumn {
	result := fivetran.NewConnectorSchemaConfigColumn()
	result.Enabled(c.enabled)
	if c.hashed != nil {
		result.Hashed(*c.hashed)
	}
	if c.isPrimaryKey != nil {
		result.IsPrimaryKey(*c.isPrimaryKey)
	}
	return result
}

func (c *_column) override(local *_column, sch string) error {
	if local != nil {
		if local.enabled != c.enabled {
			if c.isPatchAllowed() {
				c.setEnabled(local.enabled)
			} else {
				return fmt.Errorf("Attempt to patch locked column %s. The column is not allowed to change `enabled` value, reason: %v.", c.name, c.getLockReason())
			}
		}
		c.setHashed(local.hashed)
	} else {
		// patch silently if possible
		c.setEnabled(sch != BLOCK_ALL)
		// do not manage hashed for disabled columns - it doesn't make any sense
		if c.enabled {
			if c.hashed != nil && *(c.hashed) {
				c.setHashedToDefault()
			}
		} else {
			// don't pass it in request
			c.setHashed(nil)
		}
	}
	return nil
}

func (c *_column) readFromResourceData(source map[string]interface{}, sch string) {
	// Set hashed only if it is configured
	if hashed, ok := source[HASHED]; ok {
		value := getBoolValue(hashed)
		c.hashed = &value
	}
	if enabled, ok := source[ENABLED]; ok {
		c.enabled = getBoolValue(enabled)
	} else {
		c.enabled = (c.hashed != nil) || sch != BLOCK_ALL
	}

	if isPrimaryKey, ok := source[IS_PRIMARY_KEY]; ok {
		value := getBoolValue(isPrimaryKey)
		c.isPrimaryKey = &value
	}
	c.name = source[NAME].(string)
}

func (c *_column) readFromResponse(name string, response *connectors.ConnectorSchemaConfigColumnResponse) {
	c.name = name
	c.enabled = *response.Enabled
	c.hashed = response.Hashed
	c.isPrimaryKey = response.IsPrimaryKey
	c.patchAllowed = response.EnabledPatchSettings.Allowed
	if !c.isPatchAllowed() {
		lockReason := "Reason unknown. Please report this error to provider developers."
		if response.EnabledPatchSettings.ReasonCode != nil || response.EnabledPatchSettings.Reason != nil {
			code := "unknown"
			reason := "unknown"
			if response.EnabledPatchSettings.ReasonCode != nil {
				code = *response.EnabledPatchSettings.ReasonCode
			}
			if response.EnabledPatchSettings.Reason != nil {
				reason = *response.EnabledPatchSettings.Reason
			}
			lockReason = fmt.Sprintf("code: %v | reason: %v", code, reason)
		}
		c.lockReason = &lockReason
	}
}

func (c _column) toStateObject(sch string, local *_column, diag *diag.Diagnostics, schema, table string) (map[string]interface{}, bool) {
	result := make(map[string]interface{})

	result[ENABLED] = helpers.BoolToStr(c.enabled)

	// In case if table patch is not allowed we have to preserve local value in state to avoid conflict
	if local != nil {
		if c.patchAllowed != nil && !*c.patchAllowed && c.enabled != local.enabled {
			lockReason := "Unknown"
			if c.lockReason != nil {
				lockReason = *c.lockReason
			}
			diag.AddWarning(
				"Schema might be missconfigured.",
				fmt.Sprintf(
					"Column `%v` in table `%v` of schema `%v`, defined in your config, doesn't allowed to be enabled or disabled:\n"+
						"Reason: %v\n"+
						"Configured `enabled = %v` value ignored and not applied. Effective value: %v", c.name, table, schema, lockReason, local.enabled, c.enabled),
			)
			result[ENABLED] = helpers.BoolToStr(local.enabled)
		}
	}

	result[NAME] = c.name

	if local != nil && local.hashed != nil && c.hashed != nil {
		result[HASHED] = helpers.BoolToStr(*c.hashed)
	}
	return result, local != nil ||
		(c.enabled != (sch != BLOCK_ALL) && c.isPatchAllowed()) // if column is not aligned with sch it should not be included if patch not allowed
}
