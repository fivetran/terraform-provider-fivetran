package fivetran

import (
	"context"
	"fmt"
	"hash/fnv"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	ALLOW_ALL     = "ALLOW_ALL"
	ALLOW_COLUMNS = "ALLOW_COLUMNS"
	BLOCK_ALL     = "BLOCK_ALL"
	SOFT_DELETE   = "SOFT_DELETE"
	HISTORY       = "HISTORY"
	LIVE          = "LIVE"
)

func resourceSchemaConfig() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSchemaConfigCreate,
		ReadContext:   resourceSchemaConfigRead,
		UpdateContext: resourceSchemaConfigUpdate,
		DeleteContext: resourceSchemaConfigDelete,
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Schema: map[string]*schema.Schema{
			"id":                     {Type: schema.TypeString, Computed: true},
			"connector_id":           {Type: schema.TypeString, Required: true, ForceNew: true},
			"schema_change_handling": resourceSchemaConfigSchemaShangeHandling(),
			"schema":                 resourceSchemaConfigSchema(),
		},
	}
}

func resourceSchemaConfigSchemaShangeHandling() *schema.Schema {
	return &schema.Schema{Type: schema.TypeString, Required: true,
		ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
			v := val.(string)
			if !(v == ALLOW_ALL || v == ALLOW_COLUMNS || v == BLOCK_ALL) {
				errs = append(errs, fmt.Errorf("%q allowed values are: %v, %v or %v, got: %v", key, ALLOW_ALL, ALLOW_COLUMNS, BLOCK_ALL, v))
			}
			return
		},
	}
}

func resourceSchemaConfigBooleanValidateFunc(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if !(v == "true" || v == "false") {
		errs = append(errs, fmt.Errorf("%q allowed values are: %v or %v, got: %v", key, "true", "false", v))
	}
	return
}

func resourceSchemaConfigSchema() *schema.Schema {
	return &schema.Schema{Type: schema.TypeSet, Optional: true, Set: resourceSchemaConfigHash,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name":    {Type: schema.TypeString, Required: true},
				"enabled": {Type: schema.TypeString, Optional: true, Default: "true", ValidateFunc: resourceSchemaConfigBooleanValidateFunc},
				"table":   resourceSchemaConfigTable(),
			},
		},
	}
}

func resourceSchemaConfigTable() *schema.Schema {
	return &schema.Schema{Type: schema.TypeSet, Optional: true, Set: resourceTableConfigHash,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name":      {Type: schema.TypeString, Required: true},
				"enabled":   {Type: schema.TypeString, Optional: true, Default: "true", ValidateFunc: resourceSchemaConfigBooleanValidateFunc},
				"sync_mode": resourceSchemaConfigSyncMode(),
				"column":    resourceSchemaConfigColumn(),
			},
		},
	}
}

func resourceSchemaConfigSyncMode() *schema.Schema {
	return &schema.Schema{Type: schema.TypeString, Optional: true,
		ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
			v := val.(string)
			if !(v == HISTORY || v == SOFT_DELETE || v == LIVE) {
				errs = append(errs, fmt.Errorf("%q allowed values are: %v, %v or %v, got: %v", key, SOFT_DELETE, HISTORY, LIVE, v))
			}
			return
		},
	}
}

func resourceSchemaConfigColumn() *schema.Schema {
	return &schema.Schema{Type: schema.TypeSet, Optional: true, Set: resourceColumnConfigHash,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name":    {Type: schema.TypeString, Required: true},
				"enabled": {Type: schema.TypeString, Optional: true, Default: "true", ValidateFunc: resourceSchemaConfigBooleanValidateFunc},
				"hashed":  {Type: schema.TypeString, Optional: true, Default: "false", ValidateFunc: resourceSchemaConfigBooleanValidateFunc},
			},
		},
	}
}

func resourceSchemaConfigCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	connectorID := d.Get("connector_id").(string)
	client := m.(*fivetran.Client)
	var schemaChangeHandling = d.Get("schema_change_handling").(string)

	// apply SCH policy from config
	svc := client.NewConnectorSchemaUpdateService()
	updateHandlingResp, err := svc.SchemaChangeHandling(schemaChangeHandling).ConnectorID(connectorID).Do(ctx)

	if err != nil && updateHandlingResp.Code != "IllegalState" {
		return newDiagAppend(
			diags,
			diag.Error,
			"create error",
			fmt.Sprintf("%v; code: %v; message: %v", err, updateHandlingResp.Code, updateHandlingResp.Message))
	}

	// apply schema config
	applyDiags, ok := applyLocalSchemaConfig(
		d.Get("schema").(*schema.Set).List(),
		connectorID, schemaChangeHandling,
		"create error",
		ctx, client)

	if !ok {
		return applyDiags
	}

	d.SetId(connectorID)
	return resourceSchemaConfigRead(ctx, d, m)
}

func resourceSchemaConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	connectorID := d.Get("id").(string)

	schemaResponse, getDiags := getUpstreamConfigResponse(client, ctx, connectorID, "read error")
	if schemaResponse == nil {
		return getDiags
	}

	// exclude all items that are consistent with SCH policy
	alignedConfig := excludeConfigBySCH(
		readUpstreamConfig(schemaResponse),
		schemaResponse.Data.SchemaChangeHandling)

	// if local schema config aligned to SCH policy we need to include it to state to avoid drifts
	if ls, ok := d.GetOk("schema"); ok {
		s, _ := includeLocalConfiguredSchemas(alignedConfig["schema"].(map[string]interface{}), mapSchemas(ls.(*schema.Set).List()))
		alignedConfig["schema"] = s
	}

	// transform config to flat sets
	flatConfig := flattenConfig(removeExcludedSchemas(alignedConfig))

	// set state
	for k, v := range flatConfig {
		if err := d.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	return diags
}

func resourceSchemaConfigUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	connectorID := d.Get("id").(string)
	client := m.(*fivetran.Client)
	var schemaChangeHandling = d.Get("schema_change_handling").(string)

	// update SCH policy if needed
	if d.HasChange("schema_change_handling") {
		svc := client.NewConnectorSchemaUpdateService()
		updateHandlingResp, err := svc.SchemaChangeHandling(schemaChangeHandling).ConnectorID(connectorID).Do(ctx)
		// check for IllegalState error will be removed further when Fivetran API will allow to  set the same policy as it already is
		if err != nil && updateHandlingResp.Code != "IllegalState" {
			return newDiagAppend(
				diags,
				diag.Error,
				"update error",
				fmt.Sprintf("%v; code: %v; message: %v", err, updateHandlingResp.Code, updateHandlingResp.Message))
		}
	}

	// apply schema config
	applyDiags, ok := applyLocalSchemaConfig(
		d.Get("schema").(*schema.Set).List(),
		connectorID, schemaChangeHandling,
		"update error",
		ctx, client)
	if !ok {
		return applyDiags
	}

	return resourceSchemaConfigRead(ctx, d, m)
}

func resourceSchemaConfigDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// do nothing - we can't delete schema settings
	return diags
}

func applyLocalSchemaConfig(
	localSchemas []interface{},
	connectorID, sch, errorMessage string,
	ctx context.Context,
	client *fivetran.Client) (diag.Diagnostics, bool) {
	// read upstream schema config
	var diags diag.Diagnostics
	schemaResponse, getDiags := getUpstreamConfigResponse(client, ctx, connectorID, errorMessage)
	if schemaResponse == nil {
		return getDiags, false
	}

	// prepare config patch
	var alignedConfig = excludeConfigBySCH(readUpstreamConfig(schemaResponse), sch)
	config := make(map[string]interface{})
	config["schema"] = applyConfigOnAlignedUpstreamConfig(
		alignedConfig["schema"].(map[string]interface{}),
		mapSchemas(localSchemas),
		sch)
	configPatch := removeExcludedSchemas(config)

	// convert patch into request
	if schemas, ok := configPatch["schema"].(map[string]interface{}); ok {
		svc := client.NewConnectorSchemaUpdateService().ConnectorID(connectorID)
		for sname, s := range schemas {
			srequest, _ := createUpdateSchemaConfigRequest(s.(map[string]interface{}))
			svc.Schema(sname, srequest)
		}
		response, err := svc.Do(ctx)
		if err != nil {
			return newDiagAppend(
				diags,
				diag.Error,
				errorMessage,
				fmt.Sprintf("%v; code: %v, message %v", err, response.Code, response.Message)), false
		}
	}

	return diags, true
}

func includeLocalConfiguredSchemas(upstream, local map[string]interface{}) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := copyMapDeep(upstream)
	diags = newDiagAppend(diags, diag.Warning, "Handling schemas", "")
	for k, ls := range local {
		if us, ok := upstream[k]; ok {
			lsmap := ls.(map[string]interface{})
			usmap := us.(map[string]interface{})
			if ltables, ok := lsmap["table"].(map[string]interface{}); ok {
				if utables, ok := usmap["table"].(map[string]interface{}); ok {
					t, d := includeLocalConfiguredTables(utables, ltables)
					diags = append(diags, d...)
					usmap["table"] = t
				}
			}
			result[k] = include(usmap)
			diags = newDiagAppend(diags, diag.Warning, fmt.Sprintf("Handling schema %v: %+v", k, result[k]), "")
		}
	}
	diags = newDiagAppend(diags, diag.Warning, fmt.Sprintf("Updated schemas %+v", result), "")
	return result, diags
}

func includeLocalConfiguredTables(upstream, local map[string]interface{}) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := copyMapDeep(upstream)
	diags = newDiagAppend(diags, diag.Warning, "Handling tables", "")
	for k, ls := range local {
		if us, ok := upstream[k]; ok {
			lsmap := ls.(map[string]interface{})
			usmap := us.(map[string]interface{})
			if lcolumns, ok := lsmap["column"].(map[string]interface{}); ok {
				if ucolumns, ok := usmap["column"].(map[string]interface{}); ok {
					c, d := includeLocalConfiguredColumns(ucolumns, lcolumns)
					diags = append(diags, d...)
					usmap["column"] = c
				}
			}
			result[k] = include(usmap)
			diags = newDiagAppend(diags, diag.Warning, fmt.Sprintf("Handling table %v: %+v", k, result[k]), "")
		}
	}
	diags = newDiagAppend(diags, diag.Warning, fmt.Sprintf("Updated tables %+v", result), "")
	return result, diags
}

func includeLocalConfiguredColumns(upstream, local map[string]interface{}) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := copyMapDeep(upstream)
	diags = newDiagAppend(diags, diag.Warning, "Handling columns", "")
	for k := range local {
		if us, ok := upstream[k]; ok {
			usmap := us.(map[string]interface{})
			result[k] = include(usmap)
			diags = newDiagAppend(diags, diag.Warning, fmt.Sprintf("Handling column %v: %+v", k, result[k]), "")
		}
	}
	diags = newDiagAppend(diags, diag.Warning, fmt.Sprintf("Updated columns %+v", result), "")
	return result, diags
}

func createUpdateSchemaConfigRequest(schemaConfig map[string]interface{}) (*fivetran.ConnectorSchemaConfigSchema, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := fivetran.NewConnectorSchemaConfigSchema()
	if enabled, ok := schemaConfig["enabled"].(string); ok && enabled != "" {
		result.Enabled(strToBool(enabled))
	}
	if tables, ok := schemaConfig["table"]; ok && len(tables.(map[string]interface{})) > 0 {
		for tname, table := range tables.(map[string]interface{}) {
			diags = newDiagAppend(diags, diag.Warning, fmt.Sprintf("Table config for %v:\n %+v", tname, table), "")
			treq, rd := createUpdateTableConfigRequest(table.(map[string]interface{}))
			diags = append(diags, rd...)
			diags = newDiagAppend(diags, diag.Warning, fmt.Sprintf("Table request for %v:\n %+v", tname, treq), "")
			result.Table(tname, treq)
		}
	}
	return result, diags
}

func createUpdateTableConfigRequest(tableConfig map[string]interface{}) (*fivetran.ConnectorSchemaConfigTable, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := fivetran.NewConnectorSchemaConfigTable()
	if enabled, ok := tableConfig["enabled"].(string); ok && enabled != "" && !isLocked(tableConfig) {
		result.Enabled(strToBool(enabled))
	}
	if sync_mode, ok := tableConfig["sync_mode"].(string); ok && sync_mode != "" {
		result.SyncMode(sync_mode)
	}
	if columns, ok := tableConfig["column"]; ok && len(columns.(map[string]interface{})) > 0 {
		for cname, column := range columns.(map[string]interface{}) {
			diags = newDiagAppend(diags, diag.Warning, fmt.Sprintf("Column config for %v:\n %+v", cname, column), "")
			creq, rd := createUpdateColumnConfigRequest(column.(map[string]interface{}))
			diags = append(diags, rd...)
			diags = newDiagAppend(diags, diag.Warning, fmt.Sprintf("column request for %v:\n %+v", cname, creq), "")
			result.Column(cname, creq)

		}
	}
	return result, diags
}

func createUpdateColumnConfigRequest(columnConfig map[string]interface{}) (*fivetran.ConnectorSchemaConfigColumn, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := fivetran.NewConnectorSchemaConfigColumn()
	if enabled, ok := columnConfig["enabled"].(string); ok && enabled != "" && !isLocked(columnConfig) {
		result.Enabled(strToBool(enabled))
	}
	if hashed, ok := columnConfig["hashed"].(string); ok && hashed != "" && !isLocked(columnConfig) {
		result.Hashed(strToBool(hashed))
	}
	return result, diags
}

func applyConfigOnAlignedUpstreamConfig(alignedConfigSchemas map[string]interface{}, localConfigSchemas map[string]interface{}, sch string) map[string]interface{} {
	result := copyMapDeep(alignedConfigSchemas)
	for sname, s := range localConfigSchemas {
		if rs, ok := result[sname]; ok {
			result[sname] = applySchemaConfig(rs.(map[string]interface{}), s.(map[string]interface{}))
		} else {
			result[sname] = include(s.(map[string]interface{}))
		}
	}
	for rname := range result {
		result[rname] = invertUnhandledSchema(result[rname].(map[string]interface{}), sch)
	}
	return result
}

func shouldInvert(item map[string]interface{}) bool {
	return !isHandled(item) && !isLocked(item) && !isExcluded(item)
}

func invertUnhandledSchema(schema map[string]interface{}, sch string) map[string]interface{} {
	if shouldInvert(schema) {
		schema["enabled"] = boolToStr(sch == ALLOW_ALL)
	}
	if stable, ok := schema["table"].(map[string]interface{}); ok {
		invertedTables := make(map[string]interface{})
		for tname, t := range stable {
			invertedTables[tname] = invertUnhandledTable(t.(map[string]interface{}), sch)
		}
		schema["table"] = invertedTables
	}
	return schema
}

func invertUnhandledTable(table map[string]interface{}, sch string) map[string]interface{} {
	if shouldInvert(table) {
		table["enabled"] = boolToStr(sch == ALLOW_ALL)
	}
	if scolumn, ok := table["column"].(map[string]interface{}); ok {
		invertedColumns := make(map[string]interface{})
		for cname, c := range scolumn {
			invertedColumns[cname] = invertUnhandledColumn(c.(map[string]interface{}), sch)
		}
		table["column"] = invertedColumns
	}
	return table
}

func invertUnhandledColumn(column map[string]interface{}, sch string) map[string]interface{} {
	if shouldInvert(column) {
		column["enabled"] = boolToStr(sch == ALLOW_ALL || sch == ALLOW_COLUMNS)
		column["hashed"] = "false"
	}
	return column
}

func applySchemaConfig(alignedSchema map[string]interface{}, localSchema map[string]interface{}) map[string]interface{} {
	result := copyMapDeep(alignedSchema)
	if lenabled, ok := localSchema["enabled"]; ok && lenabled.(string) != "" {
		result["enabled"] = lenabled
	}
	rtables := make(map[string]interface{})
	if rts, ok := result["table"].(map[string]interface{}); ok {
		rtables = rts
	}
	if ltables, ok := localSchema["table"]; ok && len(ltables.(map[string]interface{})) > 0 {
		for ltname, lt := range ltables.(map[string]interface{}) {
			if rt, ok := rtables[ltname]; ok {
				rtables[ltname] = applyTableConfig(rt.(map[string]interface{}), lt.(map[string]interface{}))
			} else {
				rtables[ltname] = include(lt.(map[string]interface{}))
			}
		}
	}
	result["table"] = rtables
	return include(result)
}

func applyTableConfig(alignedTable map[string]interface{}, localTable map[string]interface{}) map[string]interface{} {
	result := copyMapDeep(alignedTable)
	if lenabled, ok := localTable["enabled"]; ok && lenabled.(string) != "" && !isLocked(alignedTable) {
		result["enabled"] = localTable["enabled"]
	}
	if lsync_mode, ok := localTable["sync_mode"]; ok && lsync_mode.(string) != "" {
		result["sync_mode"] = localTable["sync_mode"]
	}
	rcolumns := make(map[string]interface{})
	if rcs, ok := result["column"].(map[string]interface{}); ok {
		rcolumns = rcs
	}
	if lcolumns, ok := localTable["column"]; ok && len(lcolumns.(map[string]interface{})) > 0 {
		for lcname, lc := range lcolumns.(map[string]interface{}) {
			if rc, ok := rcolumns[lcname]; ok {
				rcolumns[lcname] = applyColumnConfig(rc.(map[string]interface{}), lc.(map[string]interface{}))
			} else {
				rcolumns[lcname] = include(lc.(map[string]interface{}))
			}
		}
	}
	result["column"] = rcolumns
	return include(result)
}

func applyColumnConfig(alignedColumn map[string]interface{}, localColumn map[string]interface{}) map[string]interface{} {
	result := copyMapDeep(alignedColumn)
	if lenabled, ok := localColumn["enabled"]; ok && lenabled.(string) != "" && !isLocked(localColumn) {
		result["enabled"] = localColumn["enabled"]
	}
	if lhashed, ok := localColumn["hashed"]; ok && lhashed.(string) != "" && !isLocked(localColumn) {
		result["hashed"] = localColumn["hashed"]
	}
	result["excluded"] = false
	return include(result)
}

func include(item map[string]interface{}) map[string]interface{} {
	item["excluded"] = false
	item["handled"] = true
	return item
}

func getUpstreamConfigResponse(
	client *fivetran.Client,
	ctx context.Context,
	connectorID, errorMessage string) (*fivetran.ConnectorSchemaDetailsResponse, diag.Diagnostics) {
	var diags diag.Diagnostics
	schemaResponse, err := client.NewConnectorSchemaDetails().ConnectorID(connectorID).Do(ctx)
	if err != nil {
		if schemaResponse.Code == "NotFound_SchemaConfig" {
			schemaResponse, err := client.NewConnectorSchemaReload().ConnectorID(connectorID).Do(ctx)
			if err != nil {
				err := fmt.Sprintf("%v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message)
				return nil, newDiagAppend(diags, diag.Error, errorMessage, err)
			}
		} else {
			err := fmt.Sprintf("%v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message)
			return nil, newDiagAppend(diags, diag.Error, errorMessage, err)
		}
	}
	return &schemaResponse, diags
}

func mapSchemas(schemas []interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, s := range schemas {
		smap := s.(map[string]interface{})
		sname := smap["name"].(string)
		rschema := make(map[string]interface{})
		rschema["enabled"] = smap["enabled"]
		if tables, ok := smap["table"].(*schema.Set); ok && len(tables.List()) > 0 {
			rschema["table"] = mapTables(tables.List())
		}
		result[sname] = rschema
	}
	return result
}

func mapTables(tables []interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, t := range tables {
		tmap := t.(map[string]interface{})
		tname := tmap["name"].(string)
		rtable := make(map[string]interface{})
		if enabled, ok := tmap["enabled"].(string); ok && enabled != "" {
			rtable["enabled"] = enabled
		}
		if sync_mode, ok := tmap["sync_mode"].(string); ok && sync_mode != "" {
			rtable["sync_mode"] = sync_mode
		}
		if columns, ok := tmap["column"].(*schema.Set); ok && len(columns.List()) > 0 {
			rtable["column"] = mapColumns(columns.List())
		}
		result[tname] = rtable
	}
	return result
}

func mapColumns(columns []interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, c := range columns {
		cmap := c.(map[string]interface{})
		cname := cmap["name"].(string)
		rcolumn := make(map[string]interface{})

		if enabled, ok := cmap["enabled"].(string); ok && enabled != "" {
			rcolumn["enabled"] = enabled
		}
		if hashed, ok := cmap["hashed"].(string); ok && hashed != "" {
			rcolumn["hashed"] = hashed
		}

		result[cname] = rcolumn
	}
	return result
}

func flattenConfig(config map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	result["schema_change_handling"] = config["schema_change_handling"]
	if schemas, ok := config["schema"]; ok {
		result["schema"] = flattenSchemas(schemas.(map[string]interface{}))
	}
	return result
}

func flattenSchemas(schemas map[string]interface{}) []interface{} {
	result := make([]interface{}, 0)
	for k, v := range schemas {
		vmap := v.(map[string]interface{})
		s := make(map[string]interface{})
		s["name"] = k
		if enabled, ok := vmap["enabled"].(string); ok && enabled != "" {
			s["enabled"] = enabled
		}
		if tables, ok := vmap["table"].(map[string]interface{}); ok {
			s["table"] = flattenTables(tables)
		}
		result = append(result, s)
	}
	return result
}

func flattenTables(tables map[string]interface{}) []interface{} {
	result := make([]interface{}, 0)
	for k, v := range tables {
		vmap := v.(map[string]interface{})
		t := make(map[string]interface{})
		t["name"] = k
		if enabled, ok := vmap["enabled"].(string); ok && enabled != "" {
			t["enabled"] = enabled
		}
		if sync_mode, ok := vmap["sync_mokde"].(string); ok && sync_mode != "" {
			t["sync_mode"] = sync_mode
		}
		if tables, ok := vmap["column"].(map[string]interface{}); ok {
			t["column"] = flattenColumns(tables)
		}
		result = append(result, t)
	}
	return result
}

func flattenColumns(columns map[string]interface{}) []interface{} {
	result := make([]interface{}, 0)
	for k, v := range columns {
		vmap := v.(map[string]interface{})
		c := make(map[string]interface{})
		c["name"] = k
		if enabled, ok := vmap["enabled"].(string); ok && enabled != "" {
			c["enabled"] = enabled
		}
		if hashed, ok := vmap["hashed"].(string); ok && hashed != "" {
			c["hashed"] = hashed
		}
		result = append(result, c)
	}
	return result
}

func excludeConfigBySCH(config map[string]interface{}, sch string) map[string]interface{} {
	result := copyMap(config)
	allSchemas := make(map[string]interface{})
	if schemas, ok := config["schema"].(map[string]interface{}); ok {
		for sname, s := range schemas {
			as := excluedSchemaBySCH(sname, s.(map[string]interface{}), sch)
			allSchemas[sname] = as
		}
		result["schema"] = allSchemas
	}
	return result
}

func excluedSchemaBySCH(sname string, schema map[string]interface{}, sch string) map[string]interface{} {
	result := copyMap(schema)
	includedTablesCount := 0
	result["table"] = make(map[string]interface{})
	if tables, ok := schema["table"].(map[string]interface{}); ok {
		for tname, t := range tables {
			at, excluded := excludeTableBySCH(tname, t.(map[string]interface{}), sch)
			if !excluded {
				includedTablesCount++
			}
			result["table"].(map[string]interface{})[tname] = at
		}
	}
	result["excluded"] = includedTablesCount == 0 && schemaEnabledAlignToSCH(schema["enabled"].(string), sch)
	return result
}

func excludeTableBySCH(tname string, table map[string]interface{}, sch string) (map[string]interface{}, bool) {
	includedColumnsCount := 0
	result := copyMap(table)
	result["column"] = make(map[string]interface{})
	if columns, ok := table["column"].(map[string]interface{}); ok {
		for cname, c := range columns {
			ac, excluded := excludeColumnBySCH(cname, c.(map[string]interface{}), sch)
			if !excluded {
				includedColumnsCount++
			}
			result["column"].(map[string]interface{})[cname] = ac
		}
	}
	excluded := includedColumnsCount == 0 && !hasSyncMode(table) && (tableEnabledAlignToSCH(table["enabled"].(string), sch) || isLocked(table))
	result["excluded"] = excluded
	return result, excluded
}

func excludeColumnBySCH(cname string, column map[string]interface{}, sch string) (map[string]interface{}, bool) {
	result := copyMap(column)
	excluded := isLocked(column) || columnEnabledAlignToSCH(column["enabled"].(string), sch)
	if !isLocked(column) && isHashed(column) {
		excluded = false
	}
	result["excluded"] = excluded
	return result, excluded
}

func columnEnabledAlignToSCH(enabled string, sch string) bool {
	if enabled == "" {
		return true
	}
	e := strToBool(enabled)
	return (sch == ALLOW_ALL || sch == ALLOW_COLUMNS) && e || sch == BLOCK_ALL && !e
}

func tableEnabledAlignToSCH(enabled string, sch string) bool {
	if enabled == "" {
		return true
	}
	e := strToBool(enabled)
	return sch == ALLOW_ALL && e || (sch == BLOCK_ALL || sch == ALLOW_COLUMNS) && !e
}

func schemaEnabledAlignToSCH(enabled string, sch string) bool {
	return tableEnabledAlignToSCH(enabled, sch)
}

func isHashed(column map[string]interface{}) bool {
	v, ok := column["hashed"]
	return ok && strToBool(v.(string))
}

func isLocked(item map[string]interface{}) bool {
	v, ok := item["patch_allowed"].(string)
	return ok && !strToBool(v)
}

func hasSyncMode(table map[string]interface{}) bool {
	v, ok := table["sync_mode"].(string)
	return ok && v != ""
}

func isExcluded(item map[string]interface{}) bool {
	v, ok := item["excluded"].(bool)
	return ok && v
}

func isHandled(item map[string]interface{}) bool {
	v, ok := item["handled"].(bool)
	return ok && v
}

func removeExcludedSchemas(config map[string]interface{}) map[string]interface{} {
	result := copyMap(config)
	result["schema"] = make(map[string]interface{})
	if schemas, ok := config["schema"]; ok {
		for sname, s := range schemas.(map[string]interface{}) {
			schema := s.(map[string]interface{})
			if excluded, ok := schema["excluded"].(bool); ok && !excluded {
				result["schema"].(map[string]interface{})[sname] = removeExcludedTables(schema)
			}
		}
	}
	return result
}

func notExcludedFilter(value interface{}) bool {
	valueMap := value.(map[string]interface{})
	excluded, ok := valueMap["excluded"].(bool)
	return !(ok && excluded)
}

func removeExcludedTables(schema map[string]interface{}) map[string]interface{} {
	result := copyMap(schema)
	delete(result, "table")
	if tables, ok := schema["table"]; ok {
		result["table"] = filterMap(tables.(map[string]interface{}), notExcludedFilter, removeExcludedColumns)
	}
	return result
}

func removeExcludedColumns(t interface{}) interface{} {
	table := t.(map[string]interface{})
	result := copyMap(table)
	delete(result, "column")
	if columns, ok := table["column"]; ok {
		result["column"] = filterMap(columns.(map[string]interface{}), notExcludedFilter, nil)
	}
	return result
}

// Function maps response without filtering by SCH (Schema Change Handling) policy
func readUpstreamConfig(response *fivetran.ConnectorSchemaDetailsResponse) map[string]interface{} {
	result := make(map[string]interface{})
	result["schema_change_handling"] = response.Data.SchemaChangeHandling
	schemas := make(map[string]interface{})
	for sname, schema := range response.Data.Schemas {
		schemaMap := readUpstreamSchema(schema)
		schemas[sname] = schemaMap
	}
	result["schema"] = schemas
	return result
}

func readUpstreamSchema(schemaResponse *fivetran.ConnectorSchemaConfigSchemaResponse) map[string]interface{} {
	result := make(map[string]interface{})
	result["enabled"] = boolPointerToStr(schemaResponse.Enabled)
	tables := make(map[string]interface{})
	for tname, table := range schemaResponse.Tables {
		tableMap := readUpstreamTable(table)
		tables[tname] = tableMap
	}
	result["table"] = tables
	return result
}

func readUpstreamTable(tableResponse *fivetran.ConnectorSchemaConfigTableResponse) map[string]interface{} {
	result := make(map[string]interface{})
	columns := make(map[string]interface{})
	for cname, column := range tableResponse.Columns {
		columnMap := readUpstreamColumn(column)
		columns[cname] = columnMap
	}
	result["column"] = columns
	result["enabled"] = boolPointerToStr(tableResponse.Enabled)
	result["sync_mode"] = tableResponse.SyncMode
	result["patch_allowed"] = boolPointerToStr(tableResponse.EnabledPatchSettings.Allowed)
	return result
}

func readUpstreamColumn(columnResponse *fivetran.ConnectorSchemaConfigColumnResponse) map[string]interface{} {
	result := make(map[string]interface{})
	result["enabled"] = boolPointerToStr(columnResponse.Enabled)
	if columnResponse.Hashed != nil {
		result["hashed"] = boolPointerToStr(columnResponse.Hashed)
	}
	result["patch_allowed"] = boolPointerToStr(columnResponse.EnabledPatchSettings.Allowed)
	return result
}

func resourceSchemaConfigHash(v interface{}) int {
	h := fnv.New32a()
	vmap := v.(map[string]interface{})
	var hashKey = vmap["name"].(string) + vmap["enabled"].(string)

	if tables, ok := vmap["table"]; ok {
		tablesHash := ""
		for _, c := range tables.(*schema.Set).List() {
			tablesHash = tablesHash + intToStr(resourceTableConfigHash(c))
		}
		hashKey = hashKey + tablesHash
	}

	h.Write([]byte(hashKey))
	return int(h.Sum32())
}

func resourceTableConfigHash(v interface{}) int {
	h := fnv.New32a()
	vmap := v.(map[string]interface{})
	var hashKey = vmap["name"].(string) + vmap["enabled"].(string) + vmap["sync_mode"].(string)

	if columns, ok := vmap["column"]; ok {
		columnsHash := ""
		for _, c := range columns.(*schema.Set).List() {
			columnsHash = columnsHash + intToStr(resourceColumnConfigHash(c))
		}
		hashKey = hashKey + columnsHash
	}

	h.Write([]byte(hashKey))
	return int(h.Sum32())
}

func resourceColumnConfigHash(v interface{}) int {
	h := fnv.New32a()
	vmap := v.(map[string]interface{})

	hashed := "false"
	if h, ok := vmap["hashed"].(string); ok {
		hashed = h
	}

	var hashKey = vmap["name"].(string) + vmap["enabled"].(string) + hashed

	h.Write([]byte(hashKey))
	return int(h.Sum32())
}
