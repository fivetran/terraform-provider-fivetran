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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// --- Schema ---

func connectionSchemaTableConfigSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages column-level settings for a specific table within a Fivetran connection schema: which columns are enabled or disabled, hashed, and primary key columns.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique resource identifier ({connection_id}:{schema_name}:{table_name}).",
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
			"table_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The name of the table within the schema.",
			},
			"disabled_columns": schema.ListAttribute{
				CustomType:  fastStringSetType,
				Optional:    true,
				ElementType: types.StringType,
				Description: "Set of column names to disable. Use when the connection's schema_change_handling is ALLOW_ALL or ALLOW_COLUMNS.",
				Validators: []validator.List{
					listvalidator.ConflictsWith(path.MatchRoot("enabled_columns")),
				},
			},
			"enabled_columns": schema.ListAttribute{
				CustomType:  fastStringSetType,
				Optional:    true,
				ElementType: types.StringType,
				Description: "Set of column names to enable. Use when the connection's schema_change_handling is BLOCK_ALL.",
				Validators: []validator.List{
					listvalidator.ConflictsWith(path.MatchRoot("disabled_columns")),
				},
			},
			"hashed_columns": schema.ListAttribute{
				CustomType:  fastStringSetType,
				Optional:    true,
				ElementType: types.StringType,
				Description: "Set of column names to hash.",
			},
			"primary_key_columns": schema.ListAttribute{
				CustomType:  fastStringSetType,
				Optional:    true,
				ElementType: types.StringType,
				Description: "Set of column names to designate as primary keys.",
			},
		},
	}
}

// --- Model ---

type connectionSchemaTableConfigModel struct {
	Id                types.String                     `tfsdk:"id"`
	ConnectionId      types.String                     `tfsdk:"connection_id"`
	SchemaName        types.String                     `tfsdk:"schema_name"`
	TableName         types.String                     `tfsdk:"table_name"`
	DisabledColumns   fivetrantypes.FastStringSetValue `tfsdk:"disabled_columns"`
	EnabledColumns    fivetrantypes.FastStringSetValue `tfsdk:"enabled_columns"`
	HashedColumns     fivetrantypes.FastStringSetValue `tfsdk:"hashed_columns"`
	PrimaryKeyColumns fivetrantypes.FastStringSetValue `tfsdk:"primary_key_columns"`
}

// readFromColumns populates the model from the column list API response.
// Column enable/disable lists are populated based on the schema_change_handling policy.
// hashed_columns and primary_key_columns report all columns matching the predicate
// so drift from external changes is detected.
// readFromColumns populates the model from the column list API response.
// Populates whichever list the user has in their config (prior state):
// disabled_columns gets columns where enabled==false, enabled_columns gets
// columns where enabled==true. The other list stays null.
func (d *connectionSchemaTableConfigModel) readFromColumns(
	columns map[string]*connections.ConnectionSchemaConfigColumnResponse,
	priorHashed fivetrantypes.FastStringSetValue,
	priorPK fivetrantypes.FastStringSetValue,
	refresh bool,
) {
	if !d.DisabledColumns.IsNull() {
		names := collectColumnNames(columns, false)
		if !refresh {
			names = filterByPrior(names, d.DisabledColumns)
		}
		d.DisabledColumns = buildOrderedSet(names, d.DisabledColumns)
	} else if !d.EnabledColumns.IsNull() {
		names := collectColumnNames(columns, true)
		if !refresh {
			names = filterByPrior(names, d.EnabledColumns)
		}
		d.EnabledColumns = buildOrderedSet(names, d.EnabledColumns)
	} else {
		// Import case: both lists are null — populate disabled columns by default
		names := collectColumnNames(columns, false)
		if len(names) > 0 {
			d.DisabledColumns = buildOrderedSet(names, d.DisabledColumns)
		}
	}

	d.HashedColumns = readBoolColumnSet(columns, priorHashed, func(c *connections.ConnectionSchemaConfigColumnResponse) bool {
		return c.Hashed != nil && *c.Hashed
	})
	d.PrimaryKeyColumns = readBoolColumnSet(columns, priorPK, func(c *connections.ConnectionSchemaConfigColumnResponse) bool {
		return c.IsPrimaryKey != nil && *c.IsPrimaryKey
	})
}

// fetchColumns calls the column list API to get all columns for a specific table.
// This is necessary because the schema details response only includes columns
// that were previously configured, not the full set of available columns.
func fetchColumns(
	ctx context.Context,
	client *fivetran.Client,
	connectionId string,
	schemaName string,
	tableName string,
) (map[string]*connections.ConnectionSchemaConfigColumnResponse, error) {
	resp, err := client.NewConnectionColumnConfigListService().
		ConnectionId(connectionId).
		Schema(schemaName).
		Table(tableName).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to list columns for table %q.%q: %v", schemaName, tableName, err)
	}
	return resp.Data.Columns, nil
}

// collectColumnNames extracts column names that match the desired enabled state.
func collectColumnNames(
	columns map[string]*connections.ConnectionSchemaConfigColumnResponse,
	wantEnabled bool,
) map[string]bool {
	result := make(map[string]bool)
	for name, c := range columns {
		if c.Enabled != nil && *c.Enabled == wantEnabled {
			result[name] = true
		}
	}
	return result
}


// fastSetAsMap extracts a FastStringSetValue as a Go map for O(1) lookups.
func fastSetAsMap(set fivetrantypes.FastStringSetValue) map[string]bool {
	result := map[string]bool{}
	if set.IsNull() || set.IsUnknown() {
		return result
	}
	for _, elem := range set.Elements() {
		if strVal, ok := elem.(types.String); ok {
			result[strVal.ValueString()] = true
		}
	}
	return result
}

// readBoolColumnSet builds a FastStringSetValue of column names matching a boolean
// predicate (e.g. hashed==true, is_primary_key==true). Only tracked when the user
// explicitly configured the attribute (prior is non-null). If the user never set
// the attribute (prior is null), returns null — external changes are ignored.
// When tracked, ALL columns matching the predicate are included so drift from
// external changes is detected.
func readBoolColumnSet(
	columns map[string]*connections.ConnectionSchemaConfigColumnResponse,
	prior fivetrantypes.FastStringSetValue,
	predicate func(*connections.ConnectionSchemaConfigColumnResponse) bool,
) fivetrantypes.FastStringSetValue {
	if prior.IsNull() {
		return fivetrantypes.NewFastStringSetNull()
	}

	names := make(map[string]bool)
	for colName, c := range columns {
		if predicate(c) {
			names[colName] = true
		}
	}

	return buildOrderedSet(names, prior)
}

// applyColumnSettings sends a PATCH request to update column-level settings
// within a specific table. Validates EnabledPatchSettings before applying.
// Returns the PATCH response and any error.
// applyColumnSettings sends a PATCH request to update column-level settings.
// Only columns explicitly listed in disabled_columns or enabled_columns are
// touched; all other columns are left in their current state.
// Validates EnabledPatchSettings and hashed/PK consistency before applying.
func applyColumnSettings(
	ctx context.Context,
	client *fivetran.Client,
	connectionId string,
	schemaName string,
	tableName string,
	data connectionSchemaTableConfigModel,
	prior connectionSchemaTableConfigModel,
	columns map[string]*connections.ConnectionSchemaConfigColumnResponse,
) (connections.ConnectionSchemaDetailsResponse, error) {
	disabledSet := fastSetAsMap(data.DisabledColumns)
	enabledSet := fastSetAsMap(data.EnabledColumns)
	priorDisabledSet := fastSetAsMap(prior.DisabledColumns)
	priorEnabledSet := fastSetAsMap(prior.EnabledColumns)
	manageHashed := !data.HashedColumns.IsNull()
	hashedSet := fastSetAsMap(data.HashedColumns)
	managePK := !data.PrimaryKeyColumns.IsNull()
	pkSet := fastSetAsMap(data.PrimaryKeyColumns)

	// Validate enabled_patch_settings before attempting any changes
	var blocked []string
	for colName, c := range columns {
		wantDisable := disabledSet[colName]
		wantEnable := enabledSet[colName]
		if wantDisable && c.Enabled != nil && *c.Enabled {
			if c.EnabledPatchSettings.Allowed != nil && !*c.EnabledPatchSettings.Allowed {
				reason := "unknown reason"
				if c.EnabledPatchSettings.Reason != nil {
					reason = *c.EnabledPatchSettings.Reason
				}
				blocked = append(blocked, fmt.Sprintf("  - %s: %s", colName, reason))
			}
		}
		if wantEnable && c.Enabled != nil && !*c.Enabled {
			if c.EnabledPatchSettings.Allowed != nil && !*c.EnabledPatchSettings.Allowed {
				reason := "unknown reason"
				if c.EnabledPatchSettings.Reason != nil {
					reason = *c.EnabledPatchSettings.Reason
				}
				blocked = append(blocked, fmt.Sprintf("  - %s: %s", colName, reason))
			}
		}
	}
	if len(blocked) > 0 {
		sort.Strings(blocked)
		return connections.ConnectionSchemaDetailsResponse{}, fmt.Errorf(
			"the following columns in table %q.%q cannot be disabled/enabled:\n%s",
			schemaName, tableName, strings.Join(blocked, "\n"))
	}

	// Validate that hashed/primary_key columns are not in disabled_columns
	var disabledConflicts []string
	for colName := range hashedSet {
		if disabledSet[colName] {
			disabledConflicts = append(disabledConflicts, colName)
		}
	}
	for colName := range pkSet {
		if disabledSet[colName] {
			disabledConflicts = append(disabledConflicts, colName)
		}
	}
	if len(disabledConflicts) > 0 {
		sort.Strings(disabledConflicts)
		return connections.ConnectionSchemaDetailsResponse{}, fmt.Errorf(
			"the following columns in table %q.%q are configured in hashed_columns or primary_key_columns "+
				"but are also in disabled_columns: %s",
			schemaName, tableName, strings.Join(disabledConflicts, ", "))
	}

	svc := client.NewConnectionTableConfigUpdateService().
		ConnectionId(connectionId).
		Schema(schemaName).
		Table(tableName)

	for colName, c := range columns {
		colConfig := fivetran.NewConnectionSchemaConfigColumn()
		needsUpdate := false

		if disabledSet[colName] {
			if c.Enabled == nil || *c.Enabled {
				colConfig.Enabled(false)
				needsUpdate = true
			}
		} else if enabledSet[colName] {
			if c.Enabled == nil || !*c.Enabled {
				colConfig.Enabled(true)
				needsUpdate = true
			}
		} else if priorDisabledSet[colName] {
			if c.Enabled != nil && !*c.Enabled {
				colConfig.Enabled(true)
				needsUpdate = true
			}
		} else if priorEnabledSet[colName] {
			if c.Enabled != nil && *c.Enabled {
				colConfig.Enabled(false)
				needsUpdate = true
			}
		}

		if manageHashed {
			if hashedSet[colName] {
				if c.Hashed == nil || !*c.Hashed {
					colConfig.Hashed(true)
					needsUpdate = true
				}
			} else {
				if c.Hashed != nil && *c.Hashed {
					colConfig.Hashed(false)
					needsUpdate = true
				}
			}
		}

		if managePK {
			if pkSet[colName] {
				if c.IsPrimaryKey == nil || !*c.IsPrimaryKey {
					colConfig.IsPrimaryKey(true)
					needsUpdate = true
				}
			}
		}

		if needsUpdate {
			svc.Columns(colName, colConfig)
		}
	}

	return svc.Do(ctx)
}

// --- Resource ---

func ConnectionSchemaTableConfig() resource.Resource {
	return &connectionSchemaTableConfig{}
}

type connectionSchemaTableConfig struct {
	core.ProviderResource
}

var _ resource.ResourceWithConfigure = &connectionSchemaTableConfig{}
var _ resource.ResourceWithImportState = &connectionSchemaTableConfig{}

func (r *connectionSchemaTableConfig) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection_schema_table_config"
}

func (r *connectionSchemaTableConfig) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = connectionSchemaTableConfigSchema()
}

func (r *connectionSchemaTableConfig) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		resp.Diagnostics.AddError("Invalid Import ID",
			fmt.Sprintf("Expected format: connection_id:schema_name:table_name, got: %s", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("connection_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("schema_name"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("table_name"), parts[2])...)
}

func (r *connectionSchemaTableConfig) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError("Unconfigured Fivetran Client", "Please report this issue to the provider developers.")
		return
	}

	var data connectionSchemaTableConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	connectionId := data.ConnectionId.ValueString()
	schemaName := data.SchemaName.ValueString()
	tableName := data.TableName.ValueString()
	client := r.GetClient()

	core.SchemaLocks.Lock(connectionId)
	defer core.SchemaLocks.Unlock(connectionId)

	core.RetryOnSchemaConflict(ctx, &resp.Diagnostics, "Unable to Update Connection Schema Column Settings.", func() error {
		schemaResp, err := client.NewConnectionSchemaDetails().ConnectionID(connectionId).Do(ctx)
		if err != nil {
			return fmt.Errorf("schema details not available for connection %s. "+
				"Ensure the schema has been loaded (e.g. via fivetran_connection_schema_reload action). "+
				"Error: %v; code: %v; message: %v", connectionId, err, schemaResp.Code, schemaResp.Message)
		}

		columns, colErr := fetchColumns(ctx, client, connectionId, schemaName, tableName)
		if colErr != nil {
			return colErr
		}

		_, err = applyColumnSettings(ctx, client, connectionId, schemaName, tableName, data, connectionSchemaTableConfigModel{}, columns)
		return err
	})
	if resp.Diagnostics.HasError() {
		return
	}

	// Re-read columns from the column list API to get the final state
	schemaResp, err := client.NewConnectionSchemaDetails().ConnectionID(connectionId).Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Connection Schema After Update.",
			fmt.Sprintf("%v; code: %v; message: %v", err, schemaResp.Code, schemaResp.Message))
		return
	}
	columns, colErr := fetchColumns(ctx, client, connectionId, schemaName, tableName)
	if colErr != nil {
		resp.Diagnostics.AddError("Unable to Read Column Settings.", colErr.Error())
		return
	}

	data.readFromColumns(columns, data.HashedColumns, data.PrimaryKeyColumns, false)
	data.Id = types.StringValue(connectionId + ":" + schemaName + ":" + tableName)
	data.ConnectionId = types.StringValue(connectionId)
	data.SchemaName = types.StringValue(schemaName)
	data.TableName = types.StringValue(tableName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionSchemaTableConfig) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError("Unconfigured Fivetran Client", "Please report this issue to the provider developers.")
		return
	}

	var data connectionSchemaTableConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	connectionId := data.ConnectionId.ValueString()
	schemaName := data.SchemaName.ValueString()
	tableName := data.TableName.ValueString()

	if connectionId == "" {
		parts := strings.SplitN(data.Id.ValueString(), ":", 3)
		if len(parts) == 3 {
			connectionId = parts[0]
			schemaName = parts[1]
			tableName = parts[2]
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
		resp.Diagnostics.AddError("Unable to Read Connection Schema Column Settings.",
			fmt.Sprintf("%v; code: %v; message: %v", err, schemaResp.Code, schemaResp.Message))
		return
	}

	columns, colErr := fetchColumns(ctx, r.GetClient(), connectionId, schemaName, tableName)
	if colErr != nil {
		resp.Diagnostics.AddError("Unable to Read Column Settings.",
			fmt.Sprintf("Table %q in schema %q not found for connection %s: %v", tableName, schemaName, connectionId, colErr))
		return
	}

	priorHashed := data.HashedColumns
	priorPK := data.PrimaryKeyColumns

	data.readFromColumns(columns, priorHashed, priorPK, true)

	data.Id = types.StringValue(connectionId + ":" + schemaName + ":" + tableName)
	data.ConnectionId = types.StringValue(connectionId)
	data.SchemaName = types.StringValue(schemaName)
	data.TableName = types.StringValue(tableName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionSchemaTableConfig) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError("Unconfigured Fivetran Client", "Please report this issue to the provider developers.")
		return
	}

	var plan, state connectionSchemaTableConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	connectionId := state.ConnectionId.ValueString()
	schemaName := state.SchemaName.ValueString()
	tableName := state.TableName.ValueString()
	client := r.GetClient()

	core.SchemaLocks.Lock(connectionId)
	defer core.SchemaLocks.Unlock(connectionId)

	core.RetryOnSchemaConflict(ctx, &resp.Diagnostics, "Unable to Update Connection Schema Column Settings.", func() error {
		schemaResp, err := client.NewConnectionSchemaDetails().ConnectionID(connectionId).Do(ctx)
		if err != nil {
			return fmt.Errorf("%v; code: %v; message: %v", err, schemaResp.Code, schemaResp.Message)
		}

		columns, colErr := fetchColumns(ctx, client, connectionId, schemaName, tableName)
		if colErr != nil {
			return colErr
		}

		_, err = applyColumnSettings(ctx, client, connectionId, schemaName, tableName, plan, state, columns)
		return err
	})
	if resp.Diagnostics.HasError() {
		return
	}

	schemaResp, err := client.NewConnectionSchemaDetails().ConnectionID(connectionId).Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Connection Schema After Update.",
			fmt.Sprintf("%v; code: %v; message: %v", err, schemaResp.Code, schemaResp.Message))
		return
	}
	columns, colErr := fetchColumns(ctx, client, connectionId, schemaName, tableName)
	if colErr != nil {
		resp.Diagnostics.AddError("Unable to Read Column Settings.", colErr.Error())
		return
	}

	plan.readFromColumns(columns, plan.HashedColumns, plan.PrimaryKeyColumns, false)
	plan.Id = types.StringValue(connectionId + ":" + schemaName + ":" + tableName)
	plan.ConnectionId = types.StringValue(connectionId)
	plan.SchemaName = types.StringValue(schemaName)
	plan.TableName = types.StringValue(tableName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *connectionSchemaTableConfig) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No-op: column settings always exist on a table.
}
