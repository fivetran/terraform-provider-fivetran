package resources

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/connections"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/fivetrantypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// --- Schema ---

func connectionSchemaTablesConfigSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages table-level settings for a specific schema within a Fivetran connection: which tables are enabled or disabled, and per-table sync modes.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique resource identifier ({connection_id}:{schema_name}).",
			},
			"connection_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The unique identifier for the connection within the Fivetran system.",
			},
			"schema_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The name of the schema within the connection.",
			},
			"disabled_tables": schema.ListAttribute{
				CustomType:  fastStringSetType,
				Optional:    true,
				ElementType: types.StringType,
				Description: "Set of table names to disable. Use when the connection's schema_change_handling is ALLOW_ALL.",
				Validators: []validator.List{
					listvalidator.ConflictsWith(path.MatchRoot("enabled_tables")),
				},
			},
			"enabled_tables": schema.ListAttribute{
				CustomType:  fastStringSetType,
				Optional:    true,
				ElementType: types.StringType,
				Description: "Set of table names to enable. Use when the connection's schema_change_handling is BLOCK_ALL or ALLOW_COLUMNS.",
				Validators: []validator.List{
					listvalidator.ConflictsWith(path.MatchRoot("disabled_tables")),
				},
			},
			"sync_mode": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Map of table name to sync mode. Allowed values: LIVE, SOFT_DELETE, HISTORY.",
				Validators: []validator.Map{
					mapvalidator.ValueStringsAre(
						stringvalidator.OneOf("LIVE", "SOFT_DELETE", "HISTORY"),
					),
				},
			},
		},
	}
}

// --- Model ---

type connectionSchemaTablesConfigModel struct {
	Id             types.String                     `tfsdk:"id"`
	ConnectionId   types.String                     `tfsdk:"connection_id"`
	SchemaName     types.String                     `tfsdk:"schema_name"`
	DisabledTables fivetrantypes.FastStringSetValue `tfsdk:"disabled_tables"`
	EnabledTables  fivetrantypes.FastStringSetValue `tfsdk:"enabled_tables"`
	SyncMode       types.Map                        `tfsdk:"sync_mode"`
}

// readFromResponse populates the model from the API response for a specific schema.
// Table enable/disable lists are populated based on the connection's schema_change_handling policy.
// sync_mode is populated only for tables present in priorSyncMode (to avoid drift from
// unmanaged tables); on import (priorSyncMode is null), all tables with non-nil sync_mode
// are included.
// readFromResponse populates the model from the API response.
// Populates whichever list the user has in their config (prior state):
// disabled_tables gets tables where enabled==false, enabled_tables gets
// tables where enabled==true. The other list stays null.
func (d *connectionSchemaTablesConfigModel) readFromResponse(
	schemaResp connections.ConnectionSchemaDetailsResponse,
	schemaName string,
	priorSyncMode types.Map,
) error {
	tables, err := getTablesFromResponse(schemaResp, schemaName)
	if err != nil {
		return err
	}

	if !d.DisabledTables.IsNull() {
		d.DisabledTables = buildOrderedSet(collectTableNames(tables, false), d.DisabledTables)
	} else if !d.EnabledTables.IsNull() {
		d.EnabledTables = buildOrderedSet(collectTableNames(tables, true), d.EnabledTables)
	} else {
		// Import: populate based on API policy
		if schemaResp.Data.SchemaChangeHandling == "ALLOW_ALL" || schemaResp.Data.SchemaChangeHandling == "ALLOW_COLUMNS" {
			names := collectTableNames(tables, false)
			if len(names) > 0 {
				d.DisabledTables = buildOrderedSet(names, d.DisabledTables)
			}
		} else {
			names := collectTableNames(tables, true)
			if len(names) > 0 {
				d.EnabledTables = buildOrderedSet(names, d.EnabledTables)
			}
		}
	}

	d.SyncMode = readSyncModes(tables, priorSyncMode)
	return nil
}

// getTablesFromResponse extracts the tables map from the API response for a
// specific schema. Returns an error if the schema is not found.
func getTablesFromResponse(
	schemaResp connections.ConnectionSchemaDetailsResponse,
	schemaName string,
) (map[string]*connections.ConnectionSchemaConfigTableResponse, error) {
	schemaData, ok := schemaResp.Data.Schemas[schemaName]
	if !ok {
		return nil, fmt.Errorf("schema %q not found in API response", schemaName)
	}
	return schemaData.Tables, nil
}

// collectTableNames extracts table names that match the desired enabled state.
// For example, wantEnabled=false returns tables where enabled==false
// (used for disabled_tables under ALLOW_ALL).
func collectTableNames(
	tables map[string]*connections.ConnectionSchemaConfigTableResponse,
	wantEnabled bool,
) map[string]bool {
	result := make(map[string]bool)
	for name, t := range tables {
		if t.Enabled != nil && *t.Enabled == wantEnabled {
			result[name] = true
		}
	}
	return result
}

// getSyncModeMap extracts the sync_mode map from the model as a Go map[string]string.
func getSyncModeMap(data connectionSchemaTablesConfigModel) map[string]string {
	result := map[string]string{}
	if data.SyncMode.IsNull() || data.SyncMode.IsUnknown() {
		return result
	}
	for k, v := range data.SyncMode.Elements() {
		if strVal, ok := v.(types.String); ok {
			result[k] = strVal.ValueString()
		}
	}
	return result
}

// readSyncModes builds a types.Map of table sync modes from the API response.
// Only tables present in priorSyncMode are included (to avoid drift from
// unmanaged tables). On import (priorSyncMode is null), all tables with a
// non-nil sync_mode are included.
func readSyncModes(
	tables map[string]*connections.ConnectionSchemaConfigTableResponse,
	priorSyncMode types.Map,
) types.Map {
	// If sync_mode was not set (null), keep it null to avoid drift
	// from unmanaged tables.
	if priorSyncMode.IsNull() {
		return types.MapNull(types.StringType)
	}

	tablesToInclude := make(map[string]bool)
	for k := range priorSyncMode.Elements() {
		tablesToInclude[k] = true
	}

	result := map[string]attr.Value{}
	for tableName, t := range tables {
		if t.SyncMode == nil {
			continue
		}
		if tablesToInclude[tableName] {
			result[tableName] = types.StringValue(*t.SyncMode)
		}
	}

	if len(result) == 0 {
		return types.MapNull(types.StringType)
	}

	mapVal, _ := types.MapValue(types.StringType, result)
	return mapVal
}

// applyTableSettings sends a PATCH request to update table-level settings
// within a specific schema. Only tables explicitly listed in disabled_tables
// or enabled_tables are touched; all other tables are left in their current
// state. Validates EnabledPatchSettings before applying.
// applyTableSettings sends a PATCH request to update table-level settings.
// disabled_tables: listed tables are disabled, all others are enabled.
// enabled_tables: listed tables are enabled, all others are disabled.
// Validates EnabledPatchSettings before applying.
func applyTableSettings(
	ctx context.Context,
	client *fivetran.Client,
	connectionId string,
	schemaName string,
	data connectionSchemaTablesConfigModel,
	tables map[string]*connections.ConnectionSchemaConfigTableResponse,
) (connections.ConnectionSchemaDetailsResponse, error) {
	disabledSet := fastSetAsMap(data.DisabledTables)
	enabledSet := fastSetAsMap(data.EnabledTables)
	syncModes := getSyncModeMap(data)

	// Validate enabled_patch_settings before attempting any changes
	var blocked []string
	for tableName, t := range tables {
		var desired bool
		if len(disabledSet) > 0 {
			desired = !disabledSet[tableName]
		} else if len(enabledSet) > 0 {
			desired = enabledSet[tableName]
		} else {
			continue
		}
		if t.Enabled != nil && *t.Enabled != desired {
			if t.EnabledPatchSettings.Allowed != nil && !*t.EnabledPatchSettings.Allowed {
				reason := "unknown reason"
				if t.EnabledPatchSettings.Reason != nil {
					reason = *t.EnabledPatchSettings.Reason
				}
				blocked = append(blocked, fmt.Sprintf("  - %s: %s", tableName, reason))
			}
		}
	}
	if len(blocked) > 0 {
		sort.Strings(blocked)
		return connections.ConnectionSchemaDetailsResponse{}, fmt.Errorf("the following tables in schema %q cannot be disabled/enabled:\n%s",
			schemaName, strings.Join(blocked, "\n"))
	}

	svc := client.NewConnectionDatabaseSchemaConfigUpdateService().
		ConnectionId(connectionId).
		Schema(schemaName)

	for tableName, t := range tables {
		tableConfig := fivetran.NewConnectionSchemaConfigTable()
		needsUpdate := false

		if len(disabledSet) > 0 {
			desired := !disabledSet[tableName]
			if t.Enabled == nil || *t.Enabled != desired {
				tableConfig.Enabled(desired)
				needsUpdate = true
			}
		} else if len(enabledSet) > 0 {
			desired := enabledSet[tableName]
			if t.Enabled == nil || *t.Enabled != desired {
				tableConfig.Enabled(desired)
				needsUpdate = true
			}
		}

		if sm, ok := syncModes[tableName]; ok {
			if t.SyncMode == nil || *t.SyncMode != sm {
				tableConfig.SyncMode(sm)
				needsUpdate = true
			}
		}

		if needsUpdate {
			svc.Tables(tableName, tableConfig)
		}
	}

	return svc.Do(ctx)
}

// --- Resource ---

func ConnectionSchemaTablesConfig() resource.Resource {
	return &connectionSchemaTablesConfig{}
}

type connectionSchemaTablesConfig struct {
	core.ProviderResource
}

var _ resource.ResourceWithConfigure = &connectionSchemaTablesConfig{}
var _ resource.ResourceWithImportState = &connectionSchemaTablesConfig{}
var _ resource.ResourceWithModifyPlan = &connectionSchemaTablesConfig{}

func (r *connectionSchemaTablesConfig) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection_schema_tables_config"
}

func (r *connectionSchemaTablesConfig) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = connectionSchemaTablesConfigSchema()
}

func (r *connectionSchemaTablesConfig) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid Import ID",
			fmt.Sprintf("Expected format: connection_id:schema_name, got: %s", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("connection_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("schema_name"), parts[1])...)
}

func (r *connectionSchemaTablesConfig) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() || r.GetClient() == nil {
		return
	}

	var plan connectionSchemaTablesConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if core.ConnectionSyncStatus.HasSynced(ctx, r.GetClient(), plan.ConnectionId.ValueString()) {
		resp.Diagnostics.AddWarning(
			"Schema Changes on a Synced Connection",
			"This connection has already synced data. Modifying table settings "+
				"(enabling/disabling tables, changing sync modes) may trigger "+
				"a full or partial resync, which can cause data delays and increased load "+
				"on the destination. Review the planned changes carefully before applying.",
		)
	}
}

func (r *connectionSchemaTablesConfig) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError("Unconfigured Fivetran Client", "Please report this issue to the provider developers.")
		return
	}

	var data connectionSchemaTablesConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	connectionId := data.ConnectionId.ValueString()
	schemaName := data.SchemaName.ValueString()
	client := r.GetClient()

	core.SchemaLocks.Lock(connectionId)
	defer core.SchemaLocks.Unlock(connectionId)

	var updatedResp connections.ConnectionSchemaDetailsResponse
	core.RetryOnSchemaConflict(ctx, &resp.Diagnostics, "Unable to Update Connection Schema Table Settings.", func() error {
		schemaResp, err := client.NewConnectionSchemaDetails().ConnectionID(connectionId).Do(ctx)
		if err != nil {
			return fmt.Errorf("schema details not available for connection %s. "+
				"Ensure the schema has been loaded (e.g. via fivetran_connection_schema_reload action). "+
				"Error: %v; code: %v; message: %v", connectionId, err, schemaResp.Code, schemaResp.Message)
		}

		tables, tableErr := getTablesFromResponse(schemaResp, schemaName)
		if tableErr != nil {
			return fmt.Errorf("schema %q not found for connection %s", schemaName, connectionId)
		}

		updatedResp, err = applyTableSettings(ctx, client, connectionId, schemaName, data, tables)
		return err
	})
	if resp.Diagnostics.HasError() {
		return
	}

	if readErr := data.readFromResponse(updatedResp, schemaName, data.SyncMode); readErr != nil {
		resp.Diagnostics.AddError("Unable to Read Table Settings.", readErr.Error())
		return
	}

	data.Id = types.StringValue(connectionId + ":" + schemaName)
	data.ConnectionId = types.StringValue(connectionId)
	data.SchemaName = types.StringValue(schemaName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionSchemaTablesConfig) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError("Unconfigured Fivetran Client", "Please report this issue to the provider developers.")
		return
	}

	var data connectionSchemaTablesConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	connectionId := data.ConnectionId.ValueString()
	schemaName := data.SchemaName.ValueString()

	if connectionId == "" {
		parts := strings.SplitN(data.Id.ValueString(), ":", 2)
		if len(parts) == 2 {
			connectionId = parts[0]
			schemaName = parts[1]
		}
	}

	schemaResp, err := r.GetClient().NewConnectionSchemaDetails().ConnectionID(connectionId).Do(ctx)
	if err != nil {
		if schemaResp.Code == "NotFound_SchemaConfig" {
			resp.Diagnostics.AddError("Connection Schema Not Loaded.",
				fmt.Sprintf("Schema config not found for connection %s. "+
					"Ensure the schema has been loaded (e.g. via fivetran_connection_schema_reload action).", connectionId))
			return
		}
		if schemaResp.Code == "NotFound_Connector" || schemaResp.Code == "NotFound_Connection" {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Read Connection Schema Table Settings.",
			fmt.Sprintf("%v; code: %v; message: %v", err, schemaResp.Code, schemaResp.Message))
		return
	}

	priorSyncMode := data.SyncMode

	if readErr := data.readFromResponse(schemaResp, schemaName, priorSyncMode); readErr != nil {
		resp.Diagnostics.AddError("Schema Not Found.",
			fmt.Sprintf("Schema %q not found in API response for connection %s.", schemaName, connectionId))
		return
	}

	data.Id = types.StringValue(connectionId + ":" + schemaName)
	data.ConnectionId = types.StringValue(connectionId)
	data.SchemaName = types.StringValue(schemaName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionSchemaTablesConfig) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError("Unconfigured Fivetran Client", "Please report this issue to the provider developers.")
		return
	}

	var plan, state connectionSchemaTablesConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	connectionId := state.ConnectionId.ValueString()
	schemaName := state.SchemaName.ValueString()
	client := r.GetClient()

	core.SchemaLocks.Lock(connectionId)
	defer core.SchemaLocks.Unlock(connectionId)

	var updatedResp connections.ConnectionSchemaDetailsResponse
	core.RetryOnSchemaConflict(ctx, &resp.Diagnostics, "Unable to Update Connection Schema Table Settings.", func() error {
		schemaResp, err := client.NewConnectionSchemaDetails().ConnectionID(connectionId).Do(ctx)
		if err != nil {
			return fmt.Errorf("%v; code: %v; message: %v", err, schemaResp.Code, schemaResp.Message)
		}

		tables, tableErr := getTablesFromResponse(schemaResp, schemaName)
		if tableErr != nil {
			return tableErr
		}

		updatedResp, err = applyTableSettings(ctx, client, connectionId, schemaName, plan, tables)
		return err
	})
	if resp.Diagnostics.HasError() {
		return
	}

	if readErr := plan.readFromResponse(updatedResp, schemaName, plan.SyncMode); readErr != nil {
		resp.Diagnostics.AddError("Unable to Read Table Settings.", readErr.Error())
		return
	}

	plan.Id = types.StringValue(connectionId + ":" + schemaName)
	plan.ConnectionId = types.StringValue(connectionId)
	plan.SchemaName = types.StringValue(schemaName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *connectionSchemaTablesConfig) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No-op: table settings always exist on a schema.
}
