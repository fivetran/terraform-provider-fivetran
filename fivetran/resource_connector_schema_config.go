package fivetran

import (
	"context"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/connectors"
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

	SCHEMA_CHANGE_HANDLING = "schema_change_handling"
	SCHEMA                 = "schema"
	TABLE                  = "table"
	COLUMN                 = "column"
	NAME                   = "name"
	ENABLED                = "enabled"
	HASHED                 = "hashed"
	SYNC_MODE              = "sync_mode"

	HANDLED       = "handled"
	EXCLUDED      = "excluded"
	PATCH_ALLOWED = "patch_allowed"
)

func resourceSchemaConfig() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSchemaConfigCreate,
		ReadWithoutTimeout:   resourceSchemaConfigRead,
		UpdateWithoutTimeout: resourceSchemaConfigUpdate,
		DeleteContext:        resourceSchemaConfigDelete,
		Importer:             &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Schema: map[string]*schema.Schema{
			ID: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique resource identifier (equals to `connector_id`).",
			},
			CONNECTOR_ID: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The unique identifier for the connector within the Fivetran system.",
			},
			SCHEMA_CHANGE_HANDLING: resourceSchemaConfigSchemaShangeHandling(),
			SCHEMA:                 resourceSchemaConfigSchema(),
		},
		Timeouts: &schema.ResourceTimeout{
			Read:   schema.DefaultTimeout(2 * time.Hour), // Import operation can trigger schema reload
			Create: schema.DefaultTimeout(2 * time.Hour),
			Update: schema.DefaultTimeout(2 * time.Hour),
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
				NAME: {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The schema name within your destination in accordance with Fivetran conventional rules.",
				},
				ENABLED: {
					Type:         schema.TypeString,
					Optional:     true,
					Default:      "true",
					ValidateFunc: resourceSchemaConfigBooleanValidateFunc,
					Description:  "The boolean value specifying whether the sync for the schema into the destination is enabled.",
				},
				TABLE: resourceSchemaConfigTable(),
			},
		},
	}
}

func resourceSchemaConfigTable() *schema.Schema {
	return &schema.Schema{Type: schema.TypeSet, Optional: true, Set: resourceTableConfigHash,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				NAME: {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The table name within your destination in accordance with Fivetran conventional rules.",
				},
				ENABLED: {
					Type:         schema.TypeString,
					Optional:     true,
					Default:      "true",
					ValidateFunc: resourceSchemaConfigBooleanValidateFunc,
					Description:  "The boolean value specifying whether the sync of table into the destination is enabled.",
				},
				SYNC_MODE: resourceSchemaConfigSyncMode(),
				COLUMN:    resourceSchemaConfigColumn(),
			},
		},
	}
}

func resourceSchemaConfigSyncMode() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "This field appears in the response if the connector supports switching sync modes for tables.",
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
				NAME: {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The column name within your destination in accordance with Fivetran conventional rules.",
				},
				ENABLED: {
					Type:         schema.TypeString,
					Optional:     true,
					Default:      "true",
					ValidateFunc: resourceSchemaConfigBooleanValidateFunc,
					Description:  "The boolean value specifying whether the sync of the column into the destination is enabled.",
				},
				HASHED: {
					Type:         schema.TypeString,
					Optional:     true,
					Default:      "false",
					ValidateFunc: resourceSchemaConfigBooleanValidateFunc,
					Description:  "The boolean value specifying whether a column should be hashed",
				},
			},
		},
	}
}

func resourceSchemaConfigCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	connectorID := d.Get(CONNECTOR_ID).(string)
	client := m.(*fivetran.Client)
	var schemaChangeHandling = d.Get(SCHEMA_CHANGE_HANDLING).(string)

	ctx, cancel := setContextTimeout(ctx, d.Timeout(schema.TimeoutCreate))
	defer cancel()

	// ensure connector has standard config with schema reloaded
	upstreamSchema, schemaDiags := getUpstreamConfigResponse(client, ctx, connectorID, "create")

	if upstreamSchema == nil {
		return schemaDiags
	}

	if upstreamSchema.Data.SchemaChangeHandling != schemaChangeHandling {
		// apply SCH policy from config
		svc := client.NewConnectorSchemaUpdateService()
		updateHandlingResp, err := svc.SchemaChangeHandling(schemaChangeHandling).ConnectorID(connectorID).Do(ctx)

		if err != nil {
			return newDiagAppend(
				diags,
				diag.Error,
				"create error",
				fmt.Sprintf("%v; code: %v; message: %v", err, updateHandlingResp.Code, updateHandlingResp.Message))
		}
	}

	// apply schema config
	applyDiags, ok := applyLocalSchemaConfig(
		d.Get(SCHEMA).(*schema.Set).List(),
		connectorID, schemaChangeHandling,
		"create error",
		ctx, client, upstreamSchema)

	if !ok {
		return applyDiags
	}

	d.SetId(connectorID)

	return resourceSchemaConfigReadImpl(ctx, d, m, false)
}

func resourceSchemaConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceSchemaConfigReadImpl(ctx, d, m, true)
}

func resourceSchemaConfigReadImpl(ctx context.Context, d *schema.ResourceData, m interface{}, setTimeout bool) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	connectorID := d.Get(ID).(string)

	// we don't need to set timeout additionally if it was already set in caller func (Create/Update)
	if setTimeout {
		var cancel context.CancelFunc
		ctx, cancel = setContextTimeout(ctx, d.Timeout(schema.TimeoutRead))
		defer cancel()
	}

	schemaResponse, getDiags := getUpstreamConfigResponse(client, ctx, connectorID, "read error")
	if schemaResponse == nil {
		return getDiags
	}

	// exclude all items that are consistent with SCH policy
	alignedConfig := excludeConfigBySCH(
		readUpstreamConfig(schemaResponse),
		schemaResponse.Data.SchemaChangeHandling)

	// if local schema config aligned to SCH policy we need to include it to state to avoid drifts
	if ls, ok := d.GetOk(SCHEMA); ok {
		s, _ := includeLocalConfiguredSchemas(alignedConfig[SCHEMA].(map[string]interface{}), mapSchemas(ls.(*schema.Set).List()))
		alignedConfig[SCHEMA] = s
	}

	// transform config to flat sets
	flatConfig := flattenConfig(removeExcludedSchemas(alignedConfig))
	flatConfig[CONNECTOR_ID] = connectorID

	// set state
	for k, v := range flatConfig {
		if err := d.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	d.SetId(connectorID)

	return diags
}

func resourceSchemaConfigUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	connectorID := d.Get(ID).(string)
	client := m.(*fivetran.Client)
	var schemaChangeHandling = d.Get(SCHEMA_CHANGE_HANDLING).(string)
	var upstreamSchema *connectors.ConnectorSchemaDetailsResponse

	ctx, cancel := setContextTimeout(ctx, d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	// update SCH policy if needed
	if d.HasChange(SCHEMA_CHANGE_HANDLING) {
		svc := client.NewConnectorSchemaUpdateService()
		updateHandlingResp, err := svc.SchemaChangeHandling(schemaChangeHandling).ConnectorID(connectorID).Do(ctx)
		// check for IllegalState error will be removed further when Fivetran API will allow to  set the same policy as it already is
		if err != nil && updateHandlingResp.Code != "IllegalState" {
			return newDiagAppend(
				diags,
				diag.Error,
				"update error",
				fmt.Sprintf("%v; code: %v; message: %v", err, updateHandlingResp.Code, updateHandlingResp.Message))
		} else {
			upstreamSchema = &updateHandlingResp
		}
	}

	// apply schema config
	applyDiags, ok := applyLocalSchemaConfig(
		d.Get(SCHEMA).(*schema.Set).List(),
		connectorID, schemaChangeHandling,
		"update error",
		ctx, client, upstreamSchema)
	if !ok {
		return applyDiags
	}

	return resourceSchemaConfigReadImpl(ctx, d, m, false)
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
	client *fivetran.Client,
	upstreamSchemaResponse *connectors.ConnectorSchemaDetailsResponse) (diag.Diagnostics, bool) {
	var diags diag.Diagnostics
	schemaResponse := upstreamSchemaResponse
	if schemaResponse == nil {
		// read upstream schema config
		upstreamResponse, getDiags := getUpstreamConfigResponse(client, ctx, connectorID, errorMessage)
		if upstreamResponse == nil {
			return getDiags, false
		}
		schemaResponse = upstreamResponse
	}

	// prepare config patch
	var upstreamConfig = readUpstreamConfig(schemaResponse)
	var alignedConfig = excludeConfigBySCH(upstreamConfig, sch)
	config := make(map[string]interface{})
	config[SCHEMA] = applyConfigOnAlignedUpstreamConfig(
		alignedConfig[SCHEMA].(map[string]interface{}),
		mapSchemas(localSchemas),
		sch)
	configPatch := removeExcludedSchemas(config)

	// convert patch into request
	if schemas, ok := configPatch[SCHEMA].(map[string]interface{}); ok && len(schemas) > 0 {
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
			if ltables, ok := lsmap[TABLE].(map[string]interface{}); ok {
				if utables, ok := usmap[TABLE].(map[string]interface{}); ok {
					t, d := includeLocalConfiguredTables(utables, ltables)
					diags = append(diags, d...)
					usmap[TABLE] = t
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
			if lcolumns, ok := lsmap[COLUMN].(map[string]interface{}); ok {
				if ucolumns, ok := usmap[COLUMN].(map[string]interface{}); ok {
					c, d := includeLocalConfiguredColumns(ucolumns, lcolumns)
					diags = append(diags, d...)
					usmap[COLUMN] = c
				}
			}

			// do not save sync_mode from upstream to state if it's not managed
			if !hasSyncMode(lsmap) {
				delete(usmap, "sync_mode")
			}

			result[k] = copyMapDeep(include(usmap))
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

func createUpdateSchemaConfigRequest(schemaConfig map[string]interface{}) (*connectors.ConnectorSchemaConfigSchema, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := fivetran.NewConnectorSchemaConfigSchema()
	if enabled, ok := schemaConfig[ENABLED].(string); ok && enabled != "" {
		result.Enabled(strToBool(enabled))
	}
	if tables, ok := schemaConfig[TABLE]; ok && len(tables.(map[string]interface{})) > 0 {
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

func createUpdateTableConfigRequest(tableConfig map[string]interface{}) (*connectors.ConnectorSchemaConfigTable, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := fivetran.NewConnectorSchemaConfigTable()
	if enabled, ok := tableConfig[ENABLED].(string); ok && enabled != "" && !isLocked(tableConfig) {
		result.Enabled(strToBool(enabled))
	}
	if sync_mode, ok := tableConfig[SYNC_MODE].(string); ok && sync_mode != "" {
		result.SyncMode(sync_mode)
	}
	if columns, ok := tableConfig[COLUMN]; ok && len(columns.(map[string]interface{})) > 0 {
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

func createUpdateColumnConfigRequest(columnConfig map[string]interface{}) (*connectors.ConnectorSchemaConfigColumn, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := fivetran.NewConnectorSchemaConfigColumn()
	if enabled, ok := columnConfig[ENABLED].(string); ok && enabled != "" && !isLocked(columnConfig) {
		result.Enabled(strToBool(enabled))
	}
	if hashed, ok := columnConfig[HASHED].(string); ok && hashed != "" && !isLocked(columnConfig) {
		result.Hashed(strToBool(hashed))
	}
	return result, diags
}

func applyConfigOnAlignedUpstreamConfig(alignedUpstreamConfigSchemas map[string]interface{}, localConfigSchemas map[string]interface{}, sch string) map[string]interface{} {
	result := copyMapDeep(alignedUpstreamConfigSchemas)
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
		schema[ENABLED] = boolToStr(sch == ALLOW_ALL)
	}
	if stable, ok := schema[TABLE].(map[string]interface{}); ok {
		invertedTables := make(map[string]interface{})
		for tname, t := range stable {
			invertedTables[tname] = invertUnhandledTable(t.(map[string]interface{}), sch)
		}
		schema[TABLE] = invertedTables
	}
	return schema
}

func invertUnhandledTable(table map[string]interface{}, sch string) map[string]interface{} {
	if shouldInvert(table) {
		table[ENABLED] = boolToStr(sch == ALLOW_ALL)
	}
	if scolumn, ok := table[COLUMN].(map[string]interface{}); ok {
		invertedColumns := make(map[string]interface{})
		for cname, c := range scolumn {
			invertedColumns[cname] = invertUnhandledColumn(c.(map[string]interface{}), sch)
		}
		table[COLUMN] = invertedColumns
	}
	// for table unhandled in config we should not touch sync_mode
	delete(table, "sync_mode")
	return copyMapDeep(table)
}

func invertUnhandledColumn(column map[string]interface{}, sch string) map[string]interface{} {
	if shouldInvert(column) {
		column[ENABLED] = boolToStr(sch == ALLOW_ALL || sch == ALLOW_COLUMNS)
		column[HASHED] = "false"
	}
	return column
}

func applySchemaConfig(alignedSchema map[string]interface{}, localSchema map[string]interface{}) map[string]interface{} {
	result := copyMapDeep(alignedSchema)
	needInclude := false
	if lenabled, ok := localSchema[ENABLED]; ok && lenabled.(string) != "" {
		if renabled, ok := result[ENABLED].(string); !ok || renabled != lenabled {
			needInclude = true
		}
		result[ENABLED] = lenabled
	}
	rtables := make(map[string]interface{})
	if rts, ok := result[TABLE].(map[string]interface{}); ok {
		rtables = rts
	}
	if ltables, ok := localSchema[TABLE]; ok && len(ltables.(map[string]interface{})) > 0 {
		for ltname, lt := range ltables.(map[string]interface{}) {
			if rt, ok := rtables[ltname]; ok {
				ut := applyTableConfig(rt.(map[string]interface{}), lt.(map[string]interface{}))
				rtables[ltname] = ut
				if !isExcluded(ut) {
					needInclude = true
				}
			} else {
				rtables[ltname] = include(lt.(map[string]interface{}))
				needInclude = true
			}
		}
	}
	result[TABLE] = rtables
	if needInclude {
		return include(result)
	}
	return handle(result)
}

func applyTableConfig(alignedTable map[string]interface{}, localTable map[string]interface{}) map[string]interface{} {
	result := copyMapDeep(alignedTable)
	needInclude := false
	if lenabled, ok := localTable[ENABLED]; ok && lenabled.(string) != "" && !isLocked(alignedTable) {
		if renabled, ok := result[ENABLED].(string); !ok || renabled != lenabled {
			needInclude = true
		}
		result[ENABLED] = localTable[ENABLED]
	}
	if lsync_mode, ok := localTable[SYNC_MODE]; ok && lsync_mode.(string) != "" {
		if rsync_mode, ok := result[SYNC_MODE].(string); !ok || rsync_mode != lsync_mode {
			needInclude = true
		}
		result[SYNC_MODE] = localTable[SYNC_MODE]
	}
	rcolumns := make(map[string]interface{})
	if rcs, ok := result[COLUMN].(map[string]interface{}); ok {
		rcolumns = rcs
	}
	if lcolumns, ok := localTable[COLUMN]; ok && len(lcolumns.(map[string]interface{})) > 0 {
		for lcname, lc := range lcolumns.(map[string]interface{}) {
			if rc, ok := rcolumns[lcname]; ok {
				uc := applyColumnConfig(rc.(map[string]interface{}), lc.(map[string]interface{}))
				rcolumns[lcname] = uc
				if !isExcluded(uc) {
					needInclude = true
				}
			} else {
				rcolumns[lcname] = include(lc.(map[string]interface{}))
				needInclude = true
			}
		}
	}
	result[COLUMN] = rcolumns
	if needInclude {
		return include(result)
	}
	return handle(result)
}

func applyColumnConfig(alignedColumn map[string]interface{}, localColumn map[string]interface{}) map[string]interface{} {
	result := copyMapDeep(alignedColumn)
	needInclude := false
	if lenabled, ok := localColumn[ENABLED]; ok && lenabled.(string) != "" && !isLocked(localColumn) {
		if renabled, ok := result[ENABLED].(string); !ok || renabled != lenabled {
			needInclude = true
		}
		result[ENABLED] = localColumn[ENABLED]
	}
	if lhashed, ok := localColumn[HASHED]; ok && lhashed.(string) != "" && !isLocked(localColumn) {
		if rhashed, ok := result[HASHED].(string); !ok || rhashed != lhashed {
			needInclude = true
		}
		result[HASHED] = localColumn[HASHED]
	}
	if needInclude {
		return include(result)
	}
	return handle(result)
}

func include(item map[string]interface{}) map[string]interface{} {
	item[EXCLUDED] = false
	return handle(item)
}
func handle(item map[string]interface{}) map[string]interface{} {
	item[HANDLED] = true
	return item
}

func getUpstreamConfigResponse(
	client *fivetran.Client,
	ctx context.Context,
	connectorID,
	errorMessage string) (*connectors.ConnectorSchemaDetailsResponse, diag.Diagnostics) {
	var diags diag.Diagnostics
	schemaResponse, err := client.NewConnectorSchemaDetails().ConnectorID(connectorID).Do(ctx)
	if err != nil {
		if schemaResponse.Code == "NotFound_SchemaConfig" {
			schemaReloadResponse, err := client.NewConnectorSchemaReload().ConnectorID(connectorID).Do(ctx)
			if err != nil {
				err := fmt.Sprintf("%v; code: %v; message: %v", err, schemaReloadResponse.Code, schemaReloadResponse.Message)
				return nil, newDiagAppend(diags, diag.Error, errorMessage, err)
			}
			return &schemaReloadResponse, diags
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
		sname := smap[NAME].(string)
		rschema := make(map[string]interface{})
		rschema[ENABLED] = smap[ENABLED]
		if tables, ok := smap[TABLE].(*schema.Set); ok && len(tables.List()) > 0 {
			rschema[TABLE] = mapTables(tables.List())
		}
		result[sname] = rschema
	}
	return result
}

func mapTables(tables []interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, t := range tables {
		tmap := t.(map[string]interface{})
		tname := tmap[NAME].(string)
		rtable := make(map[string]interface{})
		if enabled, ok := tmap[ENABLED].(string); ok && enabled != "" {
			rtable[ENABLED] = enabled
		}
		if sync_mode, ok := tmap[SYNC_MODE].(string); ok && sync_mode != "" {
			rtable[SYNC_MODE] = sync_mode
		}
		if columns, ok := tmap[COLUMN].(*schema.Set); ok && len(columns.List()) > 0 {
			rtable[COLUMN] = mapColumns(columns.List())
		}
		result[tname] = rtable
	}
	return result
}

func mapColumns(columns []interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, c := range columns {
		cmap := c.(map[string]interface{})
		cname := cmap[NAME].(string)
		rcolumn := make(map[string]interface{})

		if enabled, ok := cmap[ENABLED].(string); ok && enabled != "" {
			rcolumn[ENABLED] = enabled
		}
		if hashed, ok := cmap[HASHED].(string); ok && hashed != "" {
			rcolumn[HASHED] = hashed
		}

		result[cname] = rcolumn
	}
	return result
}

func flattenConfig(config map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	result[SCHEMA_CHANGE_HANDLING] = config[SCHEMA_CHANGE_HANDLING]
	if schemas, ok := config[SCHEMA]; ok {
		result[SCHEMA] = flattenSchemas(schemas.(map[string]interface{}))
	}
	return result
}

func flattenSchemas(schemas map[string]interface{}) []interface{} {
	result := make([]interface{}, 0)
	for k, v := range schemas {
		vmap := v.(map[string]interface{})
		s := make(map[string]interface{})
		s[NAME] = k
		if enabled, ok := vmap[ENABLED].(string); ok && enabled != "" {
			s[ENABLED] = enabled
		}
		if tables, ok := vmap[TABLE].(map[string]interface{}); ok {
			s[TABLE] = flattenTables(tables)
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
		t[NAME] = k
		if enabled, ok := vmap[ENABLED].(string); ok && enabled != "" {
			t[ENABLED] = enabled
		}
		if sync_mode, ok := vmap[SYNC_MODE].(string); ok && sync_mode != "" {
			t[SYNC_MODE] = sync_mode
		}
		if tables, ok := vmap[COLUMN].(map[string]interface{}); ok {
			t[COLUMN] = flattenColumns(tables)
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
		c[NAME] = k
		if enabled, ok := vmap[ENABLED].(string); ok && enabled != "" {
			c[ENABLED] = enabled
		}
		if hashed, ok := vmap[HASHED].(string); ok && hashed != "" {
			c[HASHED] = hashed
		}
		result = append(result, c)
	}
	return result
}

func excludeConfigBySCH(config map[string]interface{}, sch string) map[string]interface{} {
	result := copyMap(config)
	allSchemas := make(map[string]interface{})
	if schemas, ok := config[SCHEMA].(map[string]interface{}); ok {
		for sname, s := range schemas {
			as := excluedSchemaBySCH(sname, s.(map[string]interface{}), sch)
			allSchemas[sname] = as
		}
		result[SCHEMA] = allSchemas
	}
	return result
}

func excluedSchemaBySCH(sname string, schema map[string]interface{}, sch string) map[string]interface{} {
	result := copyMap(schema)
	includedTablesCount := 0
	result[TABLE] = make(map[string]interface{})
	if tables, ok := schema[TABLE].(map[string]interface{}); ok {
		for tname, t := range tables {
			at, excluded := excludeTableBySCH(tname, t.(map[string]interface{}), sch)
			if !excluded {
				includedTablesCount++
			}
			result[TABLE].(map[string]interface{})[tname] = at
		}
	}
	result[EXCLUDED] = includedTablesCount == 0 && schemaEnabledAlignToSCH(schema[ENABLED].(string), sch)
	return result
}

func excludeTableBySCH(tname string, table map[string]interface{}, sch string) (map[string]interface{}, bool) {
	includedColumnsCount := 0
	result := copyMap(table)
	result[COLUMN] = make(map[string]interface{})
	if columns, ok := table[COLUMN].(map[string]interface{}); ok {
		for cname, c := range columns {
			ac, excluded := excludeColumnBySCH(cname, c.(map[string]interface{}), sch)
			if !excluded {
				includedColumnsCount++
			}
			result[COLUMN].(map[string]interface{})[cname] = ac
		}
	}

	hasSyncMode := false //strToBool(table[ENABLED].(string)) && hasSyncMode(table)

	excluded := includedColumnsCount == 0 && !hasSyncMode && (tableEnabledAlignToSCH(table[ENABLED].(string), sch) || isLocked(table))
	result[EXCLUDED] = excluded
	return result, excluded
}

func excludeColumnBySCH(cname string, column map[string]interface{}, sch string) (map[string]interface{}, bool) {
	result := copyMap(column)
	excluded := isLocked(column) || columnEnabledAlignToSCH(column[ENABLED].(string), sch)
	if !isLocked(column) && isHashed(column) {
		excluded = false
	}
	result[EXCLUDED] = excluded
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
	v, ok := column[HASHED]
	return ok && strToBool(v.(string))
}

func isLocked(item map[string]interface{}) bool {
	v, ok := item[PATCH_ALLOWED].(string)
	return ok && !strToBool(v)
}

func hasSyncMode(table map[string]interface{}) bool {
	v, ok := table[SYNC_MODE].(string)
	return ok && v != ""
}

func isExcluded(item map[string]interface{}) bool {
	v, ok := item[EXCLUDED].(bool)
	return ok && v
}

func isHandled(item map[string]interface{}) bool {
	v, ok := item[HANDLED].(bool)
	return ok && v
}

func removeExcludedSchemas(config map[string]interface{}) map[string]interface{} {
	result := copyMap(config)
	result[SCHEMA] = make(map[string]interface{})
	if schemas, ok := config[SCHEMA]; ok {
		for sname, s := range schemas.(map[string]interface{}) {
			schema := s.(map[string]interface{})
			if excluded, ok := schema[EXCLUDED].(bool); ok && !excluded {
				result[SCHEMA].(map[string]interface{})[sname] = removeExcludedTables(schema)
			}
		}
	}
	return result
}

func notExcludedFilter(value interface{}) bool {
	valueMap := value.(map[string]interface{})
	excluded, ok := valueMap[EXCLUDED].(bool)
	return !(ok && excluded)
}

func removeExcludedTables(schema map[string]interface{}) map[string]interface{} {
	result := copyMap(schema)
	delete(result, TABLE)
	if tables, ok := schema[TABLE]; ok {
		result[TABLE] = filterMap(tables.(map[string]interface{}), notExcludedFilter, removeExcludedColumns)
	}
	return result
}

func removeExcludedColumns(t interface{}) interface{} {
	table := t.(map[string]interface{})
	result := copyMap(table)
	delete(result, COLUMN)
	if columns, ok := table[COLUMN]; ok {
		result[COLUMN] = filterMap(columns.(map[string]interface{}), notExcludedFilter, nil)
	}
	return result
}

// Function maps response without filtering by SCH (Schema Change Handling) policy
func readUpstreamConfig(response *connectors.ConnectorSchemaDetailsResponse) map[string]interface{} {
	result := make(map[string]interface{})
	result[SCHEMA_CHANGE_HANDLING] = response.Data.SchemaChangeHandling
	schemas := make(map[string]interface{})
	for sname, schema := range response.Data.Schemas {
		schemaMap := readUpstreamSchema(schema)
		schemas[sname] = schemaMap
	}
	result[SCHEMA] = schemas
	return result
}

func readUpstreamSchema(schemaResponse *connectors.ConnectorSchemaConfigSchemaResponse) map[string]interface{} {
	result := make(map[string]interface{})
	result[ENABLED] = boolPointerToStr(schemaResponse.Enabled)
	tables := make(map[string]interface{})
	for tname, table := range schemaResponse.Tables {
		tableMap := readUpstreamTable(table)
		tables[tname] = tableMap
	}
	result[TABLE] = tables
	return result
}

func readUpstreamTable(tableResponse *connectors.ConnectorSchemaConfigTableResponse) map[string]interface{} {
	result := make(map[string]interface{})
	columns := make(map[string]interface{})
	for cname, column := range tableResponse.Columns {
		columnMap := readUpstreamColumn(column)
		columns[cname] = columnMap
	}
	result[COLUMN] = columns
	result[ENABLED] = boolPointerToStr(tableResponse.Enabled)
	if tableResponse.SyncMode != nil {
		result[SYNC_MODE] = *tableResponse.SyncMode
	}
	result[PATCH_ALLOWED] = boolPointerToStr(tableResponse.EnabledPatchSettings.Allowed)
	return result
}

func readUpstreamColumn(columnResponse *connectors.ConnectorSchemaConfigColumnResponse) map[string]interface{} {
	result := make(map[string]interface{})
	result[ENABLED] = boolPointerToStr(columnResponse.Enabled)
	if columnResponse.Hashed != nil {
		result[HASHED] = boolPointerToStr(columnResponse.Hashed)
	}
	result[PATCH_ALLOWED] = boolPointerToStr(columnResponse.EnabledPatchSettings.Allowed)
	return result
}

func resourceSchemaConfigHash(v interface{}) int {
	h := fnv.New32a()
	vmap := v.(map[string]interface{})
	var hashKey = vmap[NAME].(string) + vmap[ENABLED].(string)

	if tables, ok := vmap[TABLE]; ok {
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
	var hashKey = vmap[NAME].(string) + vmap[ENABLED].(string) + vmap[SYNC_MODE].(string)

	if columns, ok := vmap[COLUMN]; ok {
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
	if h, ok := vmap[HASHED].(string); ok {
		hashed = h
	}

	var hashKey = vmap[NAME].(string) + vmap[ENABLED].(string) + hashed

	h.Write([]byte(hashKey))
	return int(h.Sum32())
}
