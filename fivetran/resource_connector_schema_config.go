package fivetran

import (
	"context"
	"fmt"
	"hash/fnv"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// (s)tring to (i)nterface Map
type siMap map[string]interface{}

const ALLOW_ALL = "ALLOW_ALL"
const ALLOW_COLUMNS = "ALLOW_COLUMNS"
const BLOCK_ALL = "BLOCK_ALL"

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
			"schema_change_handling": {Type: schema.TypeString, Required: true},
			"schema":                 resourceSchemaConfigSchema(),
		},
	}
}

func resourceSchemaConfigSchema() *schema.Schema {
	return &schema.Schema{Type: schema.TypeSet, Optional: true, Set: resourceSchemaConfigHash,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name":    {Type: schema.TypeString, Required: true},
				"enabled": {Type: schema.TypeString, Optional: true, Default: "true"},
				"table":   resourceSchemaConfigTable(),
			},
		},
	}
}

func resourceSchemaConfigTable() *schema.Schema {
	return &schema.Schema{Type: schema.TypeSet, Optional: true, Set: resourceTableConfigHash,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name":    {Type: schema.TypeString, Required: true},
				"enabled": {Type: schema.TypeString, Optional: true, Default: "true"},
				"column":  resourceSchemaConfigColumn(),
			},
		},
	}
}

func resourceSchemaConfigColumn() *schema.Schema {
	return &schema.Schema{Type: schema.TypeSet, Optional: true, Set: resourceColumnConfigHash,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name":    {Type: schema.TypeString, Required: true},
				"enabled": {Type: schema.TypeString, Optional: true, Default: "true"},
				"hashed":  {Type: schema.TypeString, Optional: true, Default: "false"},
			},
		},
	}
}

func resourceSchemaConfigCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	connectorID := d.Get("connector_id").(string)
	client := m.(*fivetran.Client)
	var schemaChangeHandling = d.Get("schema_change_handling").(string)

	svc := client.NewConnectorSchemaUpdateService()
	updateHandlingResp, err := svc.SchemaChangeHandling(schemaChangeHandling).ConnectorID(connectorID).Do(ctx)

	if err != nil && updateHandlingResp.Code != "IllegalState" {
		return newDiagAppend(
			diags,
			diag.Error,
			"create error",
			fmt.Sprintf("%v; code: %v; message: %v", err, updateHandlingResp.Code, updateHandlingResp.Message))
	}

	schemaResponse, getDiags := getUpstreamConfigResponse(client, ctx, connectorID)
	if schemaResponse == nil {
		return getDiags
	}
	var alignedConfig = excludeConfigBySCH(readUpstreamConfig(schemaResponse), schemaChangeHandling)

	diags = newDiagAppend(diags, diag.Warning, fmt.Sprintf("Upstream config:\n %+v", alignedConfig), "")

	config := make(siMap)
	config["schema_change_handling"] = schemaChangeHandling
	config["schema"] = applyConfigOnAlignedUpstreamConfig(
		alignedConfig["schema"].(siMap),
		mapSchemas(d.Get("schema").(*schema.Set).List()),
		schemaChangeHandling)
	diags = newDiagAppend(diags, diag.Warning, fmt.Sprintf("Full config:\n %+v", config), "")
	configPatch := removeExcludedSchemas(config)
	diags = newDiagAppend(diags, diag.Warning, fmt.Sprintf("Config patch:\n %+v", configPatch), "")

	if schemas, ok := configPatch["schema"].(siMap); ok {
		svc := client.NewConnectorSchemaUpdateService().ConnectorID(connectorID)
		for sname, s := range schemas {
			srequest, rd := createUpdateSchemaConfigRequest(s.(siMap))
			diags = append(diags, rd...)
			diags = newDiagAppend(diags, diag.Warning, fmt.Sprintf("Schema request for %v:\n %+v", sname, srequest), "")
			svc.Schema(sname, srequest)
		}
		response, err := svc.Do(ctx)
		if err != nil {
			return newDiagAppend(diags, diag.Warning, fmt.Sprintf("Error code: %v, message %v", response.Code, response.Message), "")
		}
	}

	d.SetId(connectorID)
	return resourceSchemaConfigRead(ctx, d, m)
}

func resourceSchemaConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	connectorID := d.Get("id").(string)

	schemaResponse, getDiags := getUpstreamConfigResponse(client, ctx, connectorID)

	if schemaResponse == nil {
		return getDiags
	}

	sch := schemaResponse.Data.SchemaChangeHandling

	fullConfig := readUpstreamConfig(schemaResponse)

	alignedConfig := excludeConfigBySCH(fullConfig, sch)

	if ls, ok := d.GetOk("schema"); ok {
		s, _ := includeLocalConfiguredSchemas(alignedConfig["schema"].(siMap), mapSchemas(ls.(*schema.Set).List()))
		alignedConfig["schema"] = s
	}

	cleanConfig := removeExcludedSchemas(alignedConfig)

	flatConfig := flattenConfig(cleanConfig)

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

	if d.HasChange("schema_change_handling") {
		svc := client.NewConnectorSchemaUpdateService()
		updateHandlingResp, err := svc.SchemaChangeHandling(schemaChangeHandling).ConnectorID(connectorID).Do(ctx)
		if err != nil && updateHandlingResp.Code != "IllegalState" {
			return newDiagAppend(
				diags,
				diag.Error,
				"update error",
				fmt.Sprintf("%v; code: %v; message: %v", err, updateHandlingResp.Code, updateHandlingResp.Message))
		}
	}

	schemaResponse, getDiags := getUpstreamConfigResponse(client, ctx, connectorID)
	if schemaResponse == nil {
		return getDiags
	}
	var alignedConfig = excludeConfigBySCH(readUpstreamConfig(schemaResponse), schemaChangeHandling)
	diags = newDiagAppend(diags, diag.Warning, fmt.Sprintf("Upstream config:\n %+v", alignedConfig), "")
	config := make(siMap)
	config["schema_change_handling"] = schemaChangeHandling
	config["schema"] = applyConfigOnAlignedUpstreamConfig(
		alignedConfig["schema"].(siMap),
		mapSchemas(d.Get("schema").(*schema.Set).List()),
		schemaChangeHandling)
	diags = newDiagAppend(diags, diag.Warning, fmt.Sprintf("Full config:\n %+v", config), "")
	configPatch := removeExcludedSchemas(config)
	diags = newDiagAppend(diags, diag.Warning, fmt.Sprintf("Config patch:\n %+v", configPatch), "")

	if schemas, ok := configPatch["schema"].(siMap); ok {
		svc := client.NewConnectorSchemaUpdateService().ConnectorID(connectorID)
		for sname, s := range schemas {
			diags = newDiagAppend(diags, diag.Warning, fmt.Sprintf("Schema config for %v:\n %+v", sname, s), "")
			srequest, rd := createUpdateSchemaConfigRequest(s.(siMap))
			diags = append(diags, rd...)
			diags = newDiagAppend(diags, diag.Warning, fmt.Sprintf("Schema request for %v:\n %+v", sname, srequest), "")
			svc.Schema(sname, srequest)
		}
		response, err := svc.Do(ctx)
		if err != nil {
			return newDiagAppend(
				diags,
				diag.Error,
				"update error",
				fmt.Sprintf("Error code: %v, message %v", response.Code, response.Message))
		}
	}

	return resourceSchemaConfigRead(ctx, d, m)
}

func resourceSchemaConfigDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// do nothing - we can't delete schema settings
	return diags
}

func includeLocalConfiguredSchemas(upstream, local siMap) (siMap, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := copyMapDeep(upstream)
	diags = newDiagAppend(diags, diag.Warning, "Handling schemas", "")
	for k, ls := range local {
		if us, ok := upstream[k]; ok {
			lsmap := ls.(siMap)
			usmap := us.(siMap)
			if ltables, ok := lsmap["table"].(siMap); ok {
				if utables, ok := usmap["table"].(siMap); ok {
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

func includeLocalConfiguredTables(upstream, local siMap) (siMap, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := copyMapDeep(upstream)
	diags = newDiagAppend(diags, diag.Warning, "Handling tables", "")
	for k, ls := range local {
		if us, ok := upstream[k]; ok {
			lsmap := ls.(siMap)
			usmap := us.(siMap)
			if ltables, ok := lsmap["column"].(siMap); ok {
				if utables, ok := usmap["column"].(siMap); ok {
					c, d := includeLocalConfiguredColumns(utables, ltables)
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

func includeLocalConfiguredColumns(upstream, local siMap) (siMap, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := copyMapDeep(upstream)
	diags = newDiagAppend(diags, diag.Warning, "Handling columns", "")
	for k := range local {
		if us, ok := upstream[k]; ok {
			usmap := us.(siMap)
			result[k] = include(usmap)
			diags = newDiagAppend(diags, diag.Warning, fmt.Sprintf("Handling column %v: %+v", k, result[k]), "")
		}
	}
	diags = newDiagAppend(diags, diag.Warning, fmt.Sprintf("Updated columns %+v", result), "")
	return result, diags
}

func createUpdateSchemaConfigRequest(schemaConfig siMap) (*fivetran.ConnectorSchemaConfigSchema, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := fivetran.NewConnectorSchemaConfigSchema()
	if enabled, ok := schemaConfig["enabled"].(string); ok && enabled != "" {
		result.Enabled(strToBool(enabled))
	}
	if tables, ok := schemaConfig["table"]; ok && len(tables.(siMap)) > 0 {
		for tname, table := range tables.(siMap) {
			diags = newDiagAppend(diags, diag.Warning, fmt.Sprintf("Table config for %v:\n %+v", tname, table), "")
			treq, rd := createUpdateTableConfigRequest(table.(siMap))
			diags = append(diags, rd...)
			diags = newDiagAppend(diags, diag.Warning, fmt.Sprintf("Table request for %v:\n %+v", tname, treq), "")
			result.Table(tname, treq)
		}
	}
	return result, diags
}

func createUpdateTableConfigRequest(tableConfig siMap) (*fivetran.ConnectorSchemaConfigTable, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := fivetran.NewConnectorSchemaConfigTable()
	if enabled, ok := tableConfig["enabled"].(string); ok && enabled != "" && !isLocked(tableConfig) {
		result.Enabled(strToBool(enabled))
	}
	if columns, ok := tableConfig["column"]; ok && len(columns.(siMap)) > 0 {
		for cname, column := range columns.(siMap) {
			diags = newDiagAppend(diags, diag.Warning, fmt.Sprintf("Column config for %v:\n %+v", cname, column), "")
			creq, rd := createUpdateColumnConfigRequest(column.(siMap))
			diags = append(diags, rd...)
			diags = newDiagAppend(diags, diag.Warning, fmt.Sprintf("column request for %v:\n %+v", cname, creq), "")
			result.Column(cname, creq)

		}
	}
	return result, diags
}

func createUpdateColumnConfigRequest(columnConfig siMap) (*fivetran.ConnectorSchemaConfigColumn, diag.Diagnostics) {
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

func applyConfigOnAlignedUpstreamConfig(alignedConfigSchemas siMap, localConfigSchemas siMap, sch string) siMap {
	result := copyMapDeep(alignedConfigSchemas)
	for sname, s := range localConfigSchemas {
		if rs, ok := result[sname]; ok {
			result[sname] = applySchemaConfig(rs.(siMap), s.(siMap))
		} else {
			result[sname] = include(s.(siMap))
		}
	}
	for rname := range result {
		result[rname] = invertUnhandledSchema(result[rname].(siMap), sch)
	}
	return result
}

func shouldInvert(item siMap) bool {
	return !isHandled(item) && !isLocked(item) && !isExcluded(item)
}

func invertUnhandledSchema(schema siMap, sch string) siMap {
	if shouldInvert(schema) {
		schema["enabled"] = boolToStr(sch == ALLOW_ALL)
	}
	if stable, ok := schema["table"].(siMap); ok {
		invertedTables := make(siMap)
		for tname, t := range stable {
			invertedTables[tname] = invertUnhandledTable(t.(siMap), sch)
		}
		schema["table"] = invertedTables
	}
	return schema
}

func invertUnhandledTable(table siMap, sch string) siMap {
	if shouldInvert(table) {
		table["enabled"] = boolToStr(sch == ALLOW_ALL)
	}
	if scolumn, ok := table["column"].(siMap); ok {
		invertedColumns := make(siMap)
		for cname, c := range scolumn {
			invertedColumns[cname] = invertUnhandledColumn(c.(siMap), sch)
		}
		table["column"] = invertedColumns
	}
	return table
}

func invertUnhandledColumn(column siMap, sch string) siMap {
	if shouldInvert(column) {
		column["enabled"] = boolToStr(sch == ALLOW_ALL || sch == ALLOW_COLUMNS)
		column["hashed"] = "false"
	}
	return column
}

func applySchemaConfig(alignedSchema siMap, localSchema siMap) siMap {
	result := copyMapDeep(alignedSchema)
	if lenabled, ok := localSchema["enabled"]; ok && lenabled.(string) != "" {
		result["enabled"] = lenabled
	}
	rtables := make(siMap)
	if rts, ok := result["table"].(siMap); ok {
		rtables = rts
	}
	if ltables, ok := localSchema["table"]; ok && len(ltables.(siMap)) > 0 {
		for ltname, lt := range ltables.(siMap) {
			if rt, ok := rtables[ltname]; ok {
				rtables[ltname] = applyTableConfig(rt.(siMap), lt.(siMap))
			} else {
				rtables[ltname] = include(lt.(siMap))
			}
		}
	}
	result["table"] = rtables
	return include(result)
}

func applyTableConfig(alignedTable siMap, localTable siMap) siMap {
	result := copyMapDeep(alignedTable)
	if lenabled, ok := localTable["enabled"]; ok && lenabled.(string) != "" && !isLocked(alignedTable) {
		result["enabled"] = localTable["enabled"]
	}
	rcolumns := make(siMap)
	if rcs, ok := result["column"].(siMap); ok {
		rcolumns = rcs
	}
	if lcolumns, ok := localTable["column"]; ok && len(lcolumns.(siMap)) > 0 {
		for lcname, lc := range lcolumns.(siMap) {
			if rc, ok := rcolumns[lcname]; ok {
				rcolumns[lcname] = applyColumnConfig(rc.(siMap), lc.(siMap))
			} else {
				rcolumns[lcname] = include(lc.(siMap))
			}
		}
	}
	result["column"] = rcolumns
	return include(result)
}

func applyColumnConfig(alignedColumn siMap, localColumn siMap) siMap {
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

func include(item siMap) siMap {
	item["excluded"] = false
	item["handled"] = true
	return item
}

func getUpstreamConfigResponse(client *fivetran.Client, ctx context.Context, connectorID string) (*fivetran.ConnectorSchemaDetailsResponse, diag.Diagnostics) {
	var diags diag.Diagnostics
	schemaResponse, err := client.NewConnectorSchemaDetails().ConnectorID(connectorID).Do(ctx)
	if err != nil {
		if schemaResponse.Code == "NotFound_SchemaConfig" {
			schemaResponse, err := client.NewConnectorSchemaReload().ConnectorID(connectorID).Do(ctx)
			if err != nil {
				err := fmt.Sprintf("%v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message)
				return nil, newDiagAppend(diags, diag.Error, err, "Creation error")
			}
		} else {
			err := fmt.Sprintf("%v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message)
			return nil, newDiagAppend(diags, diag.Error, err, "Creation error")
		}
	}
	return &schemaResponse, diags
}

func mapSchemas(schemas []interface{}) siMap {
	result := make(siMap)
	for _, s := range schemas {
		smap := s.(siMap)
		sname := smap["name"].(string)
		rschema := make(siMap)
		rschema["enabled"] = smap["enabled"]
		if tables, ok := smap["table"].(*schema.Set); ok && len(tables.List()) > 0 {
			rschema["table"] = mapTables(tables.List())
		}
		result[sname] = rschema
	}
	return result
}

func mapTables(tables []interface{}) siMap {
	result := make(siMap)
	for _, t := range tables {
		tmap := t.(siMap)
		tname := tmap["name"].(string)
		rtable := make(siMap)
		if enabled, ok := tmap["enabled"].(string); ok && enabled != "" {
			rtable["enabled"] = enabled
		}
		if columns, ok := tmap["column"].(*schema.Set); ok && len(columns.List()) > 0 {
			rtable["column"] = mapColumns(columns.List())
		}
		result[tname] = rtable
	}
	return result
}

func mapColumns(columns []interface{}) siMap {
	result := make(siMap)
	for _, c := range columns {
		cmap := c.(siMap)
		cname := cmap["name"].(string)
		rcolumn := make(siMap)

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

func flattenConfig(config siMap) siMap {
	result := make(siMap)
	result["schema_change_handling"] = config["schema_change_handling"]
	if schemas, ok := config["schema"]; ok {
		result["schema"] = flattenSchemas(schemas.(siMap))
	}
	return result
}

func flattenSchemas(schemas siMap) []interface{} {
	result := make([]interface{}, 0)
	for k, v := range schemas {
		vmap := v.(siMap)
		s := make(siMap)
		s["name"] = k
		if enabled, ok := vmap["enabled"].(string); ok && enabled != "" {
			s["enabled"] = enabled
		}
		if tables, ok := vmap["table"].(siMap); ok {
			s["table"] = flattenTables(tables)
		}
		result = append(result, s)
	}
	return result
}

func flattenTables(tables siMap) []interface{} {
	result := make([]interface{}, 0)
	for k, v := range tables {
		vmap := v.(siMap)
		t := make(siMap)
		t["name"] = k
		if enabled, ok := vmap["enabled"].(string); ok && enabled != "" {
			t["enabled"] = enabled
		}
		if tables, ok := vmap["column"].(siMap); ok {
			t["column"] = flattenColumns(tables)
		}
		result = append(result, t)
	}
	return result
}

func flattenColumns(columns siMap) []interface{} {
	result := make([]interface{}, 0)
	for k, v := range columns {
		vmap := v.(siMap)
		c := make(siMap)
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

func excludeConfigBySCH(config siMap, sch string) siMap {
	result := copyMap(config)
	allSchemas := make(siMap)
	if schemas, ok := config["schema"].(siMap); ok {
		for sname, s := range schemas {
			as := excluedSchemaBySCH(sname, s.(siMap), sch)
			allSchemas[sname] = as
		}
		result["schema"] = allSchemas
	}
	return result
}

func excluedSchemaBySCH(sname string, schema siMap, sch string) siMap {
	result := copyMap(schema)
	includedTablesCount := 0
	result["table"] = make(siMap)
	if tables, ok := schema["table"].(siMap); ok {
		for tname, t := range tables {
			at, excluded := excludeTableBySCH(tname, t.(siMap), sch)
			if !excluded {
				includedTablesCount++
			}
			result["table"].(siMap)[tname] = at
		}
	}
	result["excluded"] = includedTablesCount == 0 && schemaEnabledAlignToSCH(schema["enabled"].(string), sch)
	return result
}

func excludeTableBySCH(tname string, table siMap, sch string) (siMap, bool) {
	includedColumnsCount := 0
	result := copyMap(table)
	result["column"] = make(siMap)
	if columns, ok := table["column"].(siMap); ok {
		for cname, c := range columns {
			ac, excluded := excludeColumnBySCH(cname, c.(siMap), sch)
			if !excluded {
				includedColumnsCount++
			}
			result["column"].(siMap)[cname] = ac
		}
	}
	excluded := includedColumnsCount == 0 && (tableEnabledAlignToSCH(table["enabled"].(string), sch) || isLocked(table))
	result["excluded"] = excluded
	return result, excluded
}

func excludeColumnBySCH(cname string, column siMap, sch string) (siMap, bool) {
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

func isHashed(column siMap) bool {
	v, ok := column["hashed"]
	return ok && strToBool(v.(string))
}

func isLocked(item siMap) bool {
	v, ok := item["patch_allowed"].(string)
	return ok && !strToBool(v)
}

func isExcluded(item siMap) bool {
	v, ok := item["excluded"].(bool)
	return ok && v
}

func isHandled(item siMap) bool {
	v, ok := item["handled"].(bool)
	return ok && v
}

func removeExcludedSchemas(config siMap) siMap {
	result := copyMap(config)
	result["schema"] = make(siMap)
	if schemas, ok := config["schema"]; ok {
		for sname, s := range schemas.(siMap) {
			schema := s.(siMap)
			if excluded, ok := schema["excluded"].(bool); ok && !excluded {
				result["schema"].(siMap)[sname] = removeExcludedTables(schema)
			}
		}
	}
	return result
}

func notExcludedFilter(value interface{}) bool {
	valueMap := value.(siMap)
	excluded, ok := valueMap["excluded"].(bool)
	return !(ok && excluded)
}

func removeExcludedTables(schema siMap) siMap {
	result := copyMap(schema)
	delete(result, "table")
	if tables, ok := schema["table"]; ok {
		result["table"] = filterMap(tables.(siMap), notExcludedFilter, removeExcludedColumns)
	}
	return result
}

func removeExcludedColumns(t interface{}) interface{} {
	table := t.(siMap)
	result := copyMap(table)
	delete(result, "column")
	if columns, ok := table["column"]; ok {
		result["column"] = filterMap(columns.(siMap), notExcludedFilter, nil)
	}
	return result
}

// Function maps response without filtering by SCH (Schema Change Handling) policy
func readUpstreamConfig(response *fivetran.ConnectorSchemaDetailsResponse) siMap {
	result := make(siMap)
	result["schema_change_handling"] = response.Data.SchemaChangeHandling
	schemas := make(siMap)
	for sname, schema := range response.Data.Schemas {
		schemaMap := readUpstreamSchema(schema)
		schemas[sname] = schemaMap
	}
	result["schema"] = schemas
	return result
}

func readUpstreamSchema(schemaResponse *fivetran.ConnectorSchemaConfigSchemaResponse) siMap {
	result := make(siMap)
	result["enabled"] = boolPointerToStr(schemaResponse.Enabled)
	tables := make(siMap)
	for tname, table := range schemaResponse.Tables {
		tableMap := readUpstreamTable(table)
		tables[tname] = tableMap
	}
	result["table"] = tables
	return result
}

func readUpstreamTable(tableResponse *fivetran.ConnectorSchemaConfigTableResponse) siMap {
	result := make(siMap)
	columns := make(siMap)
	for cname, column := range tableResponse.Columns {
		columnMap := readUpstreamColumn(column)
		columns[cname] = columnMap
	}
	result["column"] = columns
	result["enabled"] = boolPointerToStr(tableResponse.Enabled)
	result["patch_allowed"] = boolPointerToStr(tableResponse.EnabledPatchSettings.Allowed)
	return result
}

func readUpstreamColumn(columnResponse *fivetran.ConnectorSchemaConfigColumnResponse) siMap {
	result := make(siMap)
	result["enabled"] = boolPointerToStr(columnResponse.Enabled)
	if columnResponse.Hashed != nil {
		result["hashed"] = boolPointerToStr(columnResponse.Hashed)
	}
	result["patch_allowed"] = boolPointerToStr(columnResponse.EnabledPatchSettings.Allowed)
	return result
}

func resourceSchemaConfigHash(v interface{}) int {
	h := fnv.New32a()
	vmap := v.(siMap)
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
	vmap := v.(siMap)
	var hashKey = vmap["name"].(string) + vmap["enabled"].(string)

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
	vmap := v.(siMap)

	hashed := "false"
	if h, ok := vmap["hashed"].(string); ok {
		hashed = h
	}

	var hashKey = vmap["name"].(string) + vmap["enabled"].(string) + hashed

	h.Write([]byte(hashKey))
	return int(h.Sum32())
}
