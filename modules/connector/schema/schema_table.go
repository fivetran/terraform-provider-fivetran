package schema

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/connectors"
	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type _table struct {
	_element
	syncMode *string
	columns  map[string]*_column
}

func (t _table) validateColumns(
	connectorId, sName, tName string,
	responseTable *connectors.ConnectorSchemaConfigTableResponse,
	client fivetran.Client,
	ctx context.Context) error {
	if len(t.columns) > 0 {
		if responseTable.SupportsColumnsConfig != nil && !*responseTable.SupportsColumnsConfig {
			return fmt.Errorf("Table `%v` of schema `%s` doesn't support columns configuration.", tName, sName)
		}
		columnsWereFetched := false
		if len(responseTable.Columns) == 0 {
			response, err := client.NewConnectorColumnConfigListService().ConnectorId(connectorId).Schema(sName).Table(tName).Do(ctx)
			if err != nil {
				return fmt.Errorf("Error while retrieving columns config for table `%s` of schema `%s. Error: %v; Code: `%v`.",
					tName, sName, err, response.Code)
			}
			responseTable.Columns = response.Data.Columns
			columnsWereFetched = true
		} else {
			for cName, _ := range t.columns {
				if _, ok := responseTable.Columns[cName]; !ok {
					if !columnsWereFetched {
						response, err := client.NewConnectorColumnConfigListService().ConnectorId(connectorId).Schema(sName).Table(tName).Do(ctx)
						if err != nil {
							return fmt.Errorf("Error while retrieving columns config for table `%s` of schema `%s. Error: %v; Code: `%v`.",
								tName, sName, err, response.Code)
						}
						responseTable.Columns = response.Data.Columns
						columnsWereFetched = true
						if _, ok := responseTable.Columns[cName]; !ok {
							return fmt.Errorf("Column `%v` of table with name `%s` not found in source schema `%s`.", cName, tName, sName)
						}
					} else {
						return fmt.Errorf("Column `%v` of table with name `%s` not found in source schema `%s`.", cName, tName, sName)
					}

				}
			}
		}
	}
	return nil
}

func (t *_table) setSyncMode(value *string) {
	if value != nil && (t.syncMode == nil || *value != *t.syncMode) {
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
func (t _table) prepareCreateRequest() *connectors.ConnectorSchemaConfigTable {
	result := fivetran.NewConnectorSchemaConfigTable()
	result.Enabled(t.enabled)
	if t.syncMode != nil {
		result.SyncMode(*t.syncMode)
	}
	for k, v := range t.columns {
		result.Column(k, v.prepareCreateRequest())
	}
	return result
}
func (t *_table) override(local *_table, sch string) error {
	if local != nil {
		if local.enabled != t.enabled {
			if t.isPatchAllowed() {
				t.setEnabled(local.enabled)
			} else {
				return fmt.Errorf("Attempt to patch locked table %s. The table is not allowed to change `enabled` value, reason: %v.", t.name, t.getLockReason())
			}
		}
		t.setSyncMode(local.syncMode)
		if len(local.columns) > 0 {
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
		}
	} else {
		t.setEnabled(sch == ALLOW_ALL)
		t.setSyncMode(nil)
		if t.enabled {
			// Handle columns that are managed in upstream and saved into standard config
			for _, c := range t.columns {
				err := c.override(nil, sch)
				if err != nil {
					return fmt.Errorf("error while patching table %s: \n\t%s", t.name, err.Error())
				}
				t.updated = t.updated || c.updated
			}
		}
	}
	return nil
}
func (t *_table) readFromResourceData(source map[string]interface{}, sch string) {

	t.name = source[NAME].(string)
	t.columns = make(map[string]*_column)
	// Set sync_mode only in case if it is configured locally
	if sm, ok := source[SYNC_MODE].(string); ok && sm != "" {
		t.syncMode = &sm
	}
	columns := getColumns(source)
	if len(columns) > 0 {
		t.readColumns(columns, sch)
	}

	if enabled, ok := source[ENABLED]; ok {
		t.enabled = getBoolValue(enabled)
	} else {
		t.enabled = len(columns) > 0 || sch == ALLOW_ALL || t.syncMode != nil
	}
}

func getColumns(source map[string]interface{}) []interface{} {
	if columns, ok := source[COLUMN].([]interface{}); ok {
		return columns
	}
	return []interface{}{}
}

func (t *_table) readColumns(columns []interface{}, sch string) {
	for _, column := range columns {
		cMap := column.(map[string]interface{})
		c := &_column{}
		c.readFromResourceData(cMap, sch)
		t.columns[cMap[NAME].(string)] = c
	}
}

func (t *_table) readFromResponse(name string, response *connectors.ConnectorSchemaConfigTableResponse) {
	t.name = name
	t.enabled = *response.Enabled
	t.patchAllowed = response.EnabledPatchSettings.Allowed
	if !t.isPatchAllowed() {
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
		t.lockReason = &lockReason
	}

	t.syncMode = response.SyncMode
	t.columns = make(map[string]*_column)

	for k, v := range response.Columns {
		c := &_column{}
		c.readFromResponse(k, v)
		t.columns[k] = c
	}
}
func (t _table) toStateObject(sch string, local *_table, diag *diag.Diagnostics, schema string) (map[string]interface{}, bool) {
	result := make(map[string]interface{})
	result[ENABLED] = helpers.BoolToStr(t.enabled)

	// In case if table patch is not allowed we have to preserve local value in state to avoid conflict
	if local != nil {
		if t.patchAllowed != nil && !*t.patchAllowed && t.enabled != local.enabled {
			lockReason := "Unknown"
			if t.lockReason != nil {
				lockReason = *t.lockReason
			}
			diag.AddWarning(
				"Schema might be missconfigured.",
				fmt.Sprintf(
					"Table `%v` of schema `%v`, defined in your config, doesn't allowed to be enabled or disabled:\n"+
						"Reason: %v;\n"+
						"Configured `enabled = %v` value ignored and not applied. Effective value: %v", t.name, schema, lockReason, local.enabled, t.enabled),
			)
			result[ENABLED] = helpers.BoolToStr(local.enabled)
		}
	}

	result[NAME] = t.name
	if t.syncMode != nil && (local != nil && local.syncMode != nil) { // save sync_mode in state only if it is configured!
		result[SYNC_MODE] = *t.syncMode
	}
	columns := make([]interface{}, 0)
	if local != nil && len(local.columns) > 0 {
		for k, v := range t.columns {
			var columnState map[string]interface{}
			var include bool
			if local != nil {
				if lc, ok := local.columns[k]; ok {
					columnState, include = v.toStateObject(sch, lc, diag, schema, k)
				} else {
					columnState, include = v.toStateObject(sch, nil, diag, schema, k)
				}
			} else {
				columnState, include = v.toStateObject(sch, nil, diag, schema, k)
			}
			if include {
				columns = append(columns, columnState)
			}
		}
		// Include columns that are defined in config, but not returned in response
		for k, v := range local.columns {
			if _, ok := t.columns[k]; !ok {
				diag.AddWarning(
					"Schema might be missconfigured.",
					fmt.Sprintf(
						"Column with name `%v` in table `%v` of schema `%v`, defined in your config, not found in upstream source config.\n"+
							"Table might be deleted from source or renamed.\n "+
							"Please remove it from your configuration, or align its name with source schema.", k, t.name, schema),
				)
				columnState, include := v.toStateObject(sch, nil, diag, schema, k)
				if include {
					columns = append(columns, columnState)
				}
			}
		}
		result[COLUMN] = columns
	}

	// table has been configured locally OR has columns to include OR table inconsistent by policy (patch allowed)
	include := local != nil || len(columns) > 0 || (t.enabled != (sch == ALLOW_ALL) && t.isPatchAllowed())

	return result, include
}
