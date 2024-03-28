package schema

import (
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/connectors"
	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
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
	return result
}

func (c *_column) override(local *_column, sch string) error {
	if local != nil {
		if local.enabled != c.enabled {
			if c.isPatchAllowed() {
				c.setEnabled(local.enabled)
			} else {
				return fmt.Errorf("attempt to patch locked column %s", c.name)
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
	c.name = source[NAME].(string)
}

func (c *_column) readFromResponse(name string, response *connectors.ConnectorSchemaConfigColumnResponse) {
	c.name = name
	c.enabled = *response.Enabled
	c.hashed = response.Hashed
	c.patchAllowed = response.EnabledPatchSettings.Allowed
}

func (c _column) toStateObject(sch string, local *_column) (map[string]interface{}, bool) {
	result := make(map[string]interface{})

	result[ENABLED] = helpers.BoolToStr(c.enabled)
	result[NAME] = c.name

	if c.hashed != nil {
		result[HASHED] = helpers.BoolToStr(*c.hashed)
	}
	return result, local != nil ||
		(c.enabled != (sch != BLOCK_ALL) && c.isPatchAllowed()) // if column is not aligned with sch it should not be included if patch not allowed
}
