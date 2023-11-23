package schema

import (
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/connectors"
	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type _table struct {
	_element
	syncMode *string
	columns  map[string]*_column
}

func (t *_table) setSyncMode(value *string) {
	if value != nil && *value != *t.syncMode {
		t.syncMode = value
		t.updated = true
	} else {
		t.syncMode = nil
	}
}
func (t _table) prepareRequest() *connectors.ConnectorSchemaConfigTable {
	result := fivetran.NewConnectorSchemaConfigTable()

	if t.enabledPatched && t.isPatchAllowed() {
		result.Enabled(t.enabled)
	}

	if t.syncMode != nil {
		result.SyncMode(*t.syncMode)
	}
	for k, v := range t.columns {
		if v.updated {
			result.Column(k, v.prepareRequest())
		}
	}
	return result
}
func (t *_table) override(local *_table, sch string) error {
	if local != nil {
		if local.enabled != t.enabled {
			if t.isPatchAllowed() {
				t.setEnabled(local.enabled)
			} else {
				return fmt.Errorf("attempt to patch locked table %s", t.name)
			}
		}

		t.setSyncMode(local.syncMode)
		// Handle columns that are managed in upstream and saved into standard config
		for cName, c := range t.columns {
			if lColumn, ok := local.columns[cName]; ok {
				err := c.override(lColumn, sch)
				if err != nil {
					return fmt.Errorf("error while patching table %s: \n\t%s", t.name, err.Error())
				}
				t.updated = t.updated || c.updated
			} else {
				err := c.override(nil, sch)
				if err != nil {
					return fmt.Errorf("error while patching table %s: \n\t%s", t.name, err.Error())
				}
				t.updated = t.updated || c.updated
			}
		}
		// Api returns only columns that were previosly managed by user
		// or columns that aren't aligned with the current schema chenge handling policy
		// So when we are applying current schema config we should keep it in mind

		// Handle columns that are not repesented in unsptream config
		for lcName, lc := range local.columns {
			if _, ok := t.columns[lcName]; !ok {
				t.columns[lcName] = lc
				t.columns[lcName].updated = true
				t.columns[lcName].enabledPatched = true
				t.updated = true
			}
		}
	} else {
		t.setEnabled(sch == ALLOW_ALL)
		t.setSyncMode(nil)
		// Handle columns that are managed in upstream and saved into standard config
		for _, c := range t.columns {
			err := c.override(nil, sch)
			if err != nil {
				return fmt.Errorf("error while patching table %s: \n\t%s", t.name, err.Error())
			}
			t.updated = t.updated || c.updated
		}
	}
	return nil
}
func (t *_table) readFromResourceData(source map[string]interface{}) {
	t.enabled = helpers.StrToBool(source[ENABLED].(string))
	t.name = source[NAME].(string)
	t.columns = make(map[string]*_column)

	// Set sync_mode only in case if it is configured locally
	if sm, ok := source[SYNC_MODE].(string); ok && sm != "" {
		t.syncMode = &sm
	}

	if columns, ok := source[COLUMN].(*schema.Set); ok && columns.Len() > 0 {
		for _, column := range columns.List() {
			cMap := column.(map[string]interface{})
			c := &_column{}
			c.readFromResourceData(cMap)
			t.columns[cMap[NAME].(string)] = c
		}
	}
}

func (t *_table) readFromResponse(name string, response *connectors.ConnectorSchemaConfigTableResponse) {
	t.name = name
	t.enabled = *response.Enabled
	t.patchAllowed = response.EnabledPatchSettings.Allowed

	t.syncMode = response.SyncMode
	t.columns = make(map[string]*_column)

	for k, v := range response.Columns {
		c := &_column{}
		c.readFromResponse(k, v)
		t.columns[k] = c
	}
}
func (t _table) toStateObject(sch string, local *_table) (map[string]interface{}, bool) {
	result := make(map[string]interface{})
	result[ENABLED] = helpers.BoolToStr(t.enabled)
	result[NAME] = t.name
	if t.syncMode != nil && (local != nil && local.syncMode != nil) { // save sync_mode in state only if it is configured!
		result[SYNC_MODE] = *t.syncMode
	}

	columns := make([]interface{}, 0)
	for k, v := range t.columns {
		var columnState map[string]interface{}
		var include bool
		if local != nil {
			if lc, ok := local.columns[k]; ok {
				columnState, include = v.toStateObject(sch, lc)
			} else {
				columnState, include = v.toStateObject(sch, nil)
			}
		} else {
			columnState, include = v.toStateObject(sch, nil)
		}
		if include {
			columns = append(columns, columnState)
		}
	}

	result[COLUMN] = columns

	// table has been configured locally OR has columns to include OR table inconsistent by policy (patch allowed)
	include := local != nil || len(columns) > 0 || (t.enabled != (sch == ALLOW_ALL) && t.isPatchAllowed())

	return result, include
}
