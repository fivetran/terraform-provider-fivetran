package resources

import (
	"context"
	"fmt"
	"sort"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/connections"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/fivetrantypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// --- Schema ---

var fastStringSetType = fivetrantypes.FastStringSetType{
	ListType: basetypes.ListType{ElemType: types.StringType},
}

func connectionSchemasConfigSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages schema-level settings for a Fivetran connection: the schema_change_handling policy and which schemas are enabled or disabled.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique resource identifier (equals to connection_id).",
			},
			"connection_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The unique identifier for the connection within the Fivetran system.",
			},
			"schema_change_handling": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("ALLOW_ALL", "ALLOW_COLUMNS", "BLOCK_ALL"),
				},
				Description: "The value specifying how new source schema changes are handled. One of: ALLOW_ALL, ALLOW_COLUMNS, BLOCK_ALL.",
			},
			"disabled_schemas": schema.ListAttribute{
				CustomType:  fastStringSetType,
				Optional:    true,
				ElementType: types.StringType,
				Description: "Set of schema names to disable. Use when schema_change_handling is ALLOW_ALL.",
				Validators: []validator.List{
					listvalidator.ConflictsWith(path.MatchRoot("enabled_schemas")),
				},
			},
			"enabled_schemas": schema.ListAttribute{
				CustomType:  fastStringSetType,
				Optional:    true,
				ElementType: types.StringType,
				Description: "Set of schema names to enable. Use when schema_change_handling is BLOCK_ALL or ALLOW_COLUMNS.",
				Validators: []validator.List{
					listvalidator.ConflictsWith(path.MatchRoot("disabled_schemas")),
				},
			},
		},
	}
}

// --- Model ---

type connectionSchemasConfigModel struct {
	Id                   types.String                      `tfsdk:"id"`
	ConnectionId         types.String                      `tfsdk:"connection_id"`
	SchemaChangeHandling types.String                      `tfsdk:"schema_change_handling"`
	DisabledSchemas      fivetrantypes.FastStringSetValue `tfsdk:"disabled_schemas"`
	EnabledSchemas       fivetrantypes.FastStringSetValue `tfsdk:"enabled_schemas"`
}

// readFromResponse populates the model from the API response.
// If the user has disabled_schemas: reports all schemas where enabled==false.
// If the user has enabled_schemas: reports all schemas where enabled==true.
// On import (both null): falls back to policy-based population.
func (d *connectionSchemasConfigModel) readFromResponse(resp connections.ConnectionSchemaDetailsResponse) {
	d.SchemaChangeHandling = types.StringValue(resp.Data.SchemaChangeHandling)

	if !d.DisabledSchemas.IsNull() {
		d.DisabledSchemas = buildOrderedSet(collectSchemaNames(resp, false), d.DisabledSchemas)
	} else if !d.EnabledSchemas.IsNull() {
		d.EnabledSchemas = buildOrderedSet(collectSchemaNames(resp, true), d.EnabledSchemas)
	} else {
		// Import: populate based on API policy
		if resp.Data.SchemaChangeHandling == "ALLOW_ALL" {
			names := collectSchemaNames(resp, false)
			if len(names) > 0 {
				d.DisabledSchemas = buildOrderedSet(names, d.DisabledSchemas)
			}
		} else {
			names := collectSchemaNames(resp, true)
			if len(names) > 0 {
				d.EnabledSchemas = buildOrderedSet(names, d.EnabledSchemas)
			}
		}
	}
}

// collectSchemaNames extracts schema names from the API response that match
// the desired enabled state. For example, wantEnabled=false returns schemas
// where enabled==false (used for disabled_schemas under ALLOW_ALL).
func collectSchemaNames(resp connections.ConnectionSchemaDetailsResponse, wantEnabled bool) map[string]bool {
	result := make(map[string]bool)
	for name, s := range resp.Data.Schemas {
		if s.Enabled != nil && *s.Enabled == wantEnabled {
			result[name] = true
		}
	}
	return result
}

// buildOrderedSet builds a FastStringSetValue from a set of schema names,
// preserving the element order from the prior state to avoid false diffs.
// Items present in both prior and names keep their prior order; new items
// (not in prior, e.g. after import or upstream drift) are appended in sorted
// order for deterministic state.
func buildOrderedSet(names map[string]bool, prior fivetrantypes.FastStringSetValue) fivetrantypes.FastStringSetValue {
	var ordered []string

	// Preserve prior order for items that still exist
	if !prior.IsNull() && !prior.IsUnknown() {
		for _, elem := range prior.Elements() {
			if strVal, ok := elem.(types.String); ok {
				name := strVal.ValueString()
				if names[name] {
					ordered = append(ordered, name)
					delete(names, name)
				}
			}
		}
	}

	// Append any new items (not in prior state) in sorted order for determinism
	remaining := make([]string, 0, len(names))
	for name := range names {
		remaining = append(remaining, name)
	}
	sort.Strings(remaining)
	ordered = append(ordered, remaining...)

	return fivetrantypes.NewFastStringSetFromStrings(ordered)
}

// --- Resource ---

func ConnectionSchemasConfig() resource.Resource {
	return &connectionSchemasConfig{}
}

type connectionSchemasConfig struct {
	core.ProviderResource
}

var _ resource.ResourceWithConfigure = &connectionSchemasConfig{}
var _ resource.ResourceWithImportState = &connectionSchemasConfig{}
var _ resource.ResourceWithModifyPlan = &connectionSchemasConfig{}

func (r *connectionSchemasConfig) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection_schemas_config"
}

func (r *connectionSchemasConfig) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = connectionSchemasConfigSchema()
}

func (r *connectionSchemasConfig) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *connectionSchemasConfig) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() || r.GetClient() == nil {
		return
	}

	var plan connectionSchemasConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if core.ConnectionSyncStatus.HasSynced(ctx, r.GetClient(), plan.ConnectionId.ValueString()) {
		resp.Diagnostics.AddWarning(
			"Schema Changes on a Synced Connection",
			"This connection has already synced data. Modifying schema settings "+
				"(enabling/disabling schemas, changing schema_change_handling) may trigger "+
				"a full or partial resync, which can cause data delays and increased load "+
				"on the destination. Review the planned changes carefully before applying.",
		)
	}
}

func (r *connectionSchemasConfig) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError("Unconfigured Fivetran Client", "Please report this issue to the provider developers.")
		return
	}

	var data connectionSchemasConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	connectionId := data.ConnectionId.ValueString()
	client := r.GetClient()

	core.SchemaLocks.Lock(connectionId)
	defer core.SchemaLocks.Unlock(connectionId)

	var updatedResp connections.ConnectionSchemaDetailsResponse
	core.RetryOnSchemaConflict(ctx, &resp.Diagnostics, "Unable to Update Connection Schema Settings.", func() error {
		schemaResp, err := client.NewConnectionSchemaDetails().ConnectionID(connectionId).Do(ctx)
		if err != nil {
			return fmt.Errorf("schema details not available for connection %s. "+
				"Ensure the schema has been loaded (e.g. via fivetran_connection_schema_reload action). "+
				"Error: %v; code: %v; message: %v", connectionId, err, schemaResp.Code, schemaResp.Message)
		}
		updatedResp, err = applySchemaSettings(ctx, client, connectionId, data, schemaResp)
		return err
	})
	if resp.Diagnostics.HasError() {
		return
	}

	data.readFromResponse(updatedResp)
	data.Id = types.StringValue(connectionId)
	data.ConnectionId = types.StringValue(connectionId)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionSchemasConfig) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError("Unconfigured Fivetran Client", "Please report this issue to the provider developers.")
		return
	}

	var data connectionSchemasConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	connectionId := data.ConnectionId.ValueString()
	if connectionId == "" {
		connectionId = data.Id.ValueString()
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
		resp.Diagnostics.AddError("Unable to Read Connection Schema Settings.",
			fmt.Sprintf("%v; code: %v; message: %v", err, schemaResp.Code, schemaResp.Message))
		return
	}

	data.readFromResponse(schemaResp)
	data.Id = types.StringValue(connectionId)
	data.ConnectionId = types.StringValue(connectionId)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionSchemasConfig) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError("Unconfigured Fivetran Client", "Please report this issue to the provider developers.")
		return
	}

	var plan, state connectionSchemasConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	connectionId := state.ConnectionId.ValueString()
	client := r.GetClient()

	core.SchemaLocks.Lock(connectionId)
	defer core.SchemaLocks.Unlock(connectionId)

	var updatedResp connections.ConnectionSchemaDetailsResponse
	core.RetryOnSchemaConflict(ctx, &resp.Diagnostics, "Unable to Update Connection Schema Settings.", func() error {
		schemaResp, err := client.NewConnectionSchemaDetails().ConnectionID(connectionId).Do(ctx)
		if err != nil {
			return fmt.Errorf("%v; code: %v; message: %v", err, schemaResp.Code, schemaResp.Message)
		}
		updatedResp, err = applySchemaSettings(ctx, client, connectionId, plan, schemaResp)
		return err
	})
	if resp.Diagnostics.HasError() {
		return
	}

	plan.readFromResponse(updatedResp)
	plan.Id = types.StringValue(connectionId)
	plan.ConnectionId = types.StringValue(connectionId)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *connectionSchemasConfig) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No-op: schema settings always exist on a connection.
}

// --- Helpers ---

// applySchemaSettings sends a PATCH request to update the connection's
// schema_change_handling policy and per-schema enabled state.
// disabled_schemas: listed schemas are disabled, all others are enabled.
// enabled_schemas: listed schemas are enabled, all others are disabled.
func applySchemaSettings(
	ctx context.Context,
	client *fivetran.Client,
	connectionId string,
	data connectionSchemasConfigModel,
	schemaResp connections.ConnectionSchemaDetailsResponse,
) (connections.ConnectionSchemaDetailsResponse, error) {
	policy := data.SchemaChangeHandling.ValueString()

	svc := client.NewConnectionSchemaUpdateService().ConnectionID(connectionId)
	svc.SchemaChangeHandling(policy)

	disabledSet := fastSetAsMap(data.DisabledSchemas)
	enabledSet := fastSetAsMap(data.EnabledSchemas)

	for schemaName, s := range schemaResp.Data.Schemas {
		var desired bool
		if len(disabledSet) > 0 {
			desired = !disabledSet[schemaName] // not in disabled list → enabled
		} else if len(enabledSet) > 0 {
			desired = enabledSet[schemaName] // in enabled list → enabled
		} else {
			continue // neither list set, skip
		}
		if s.Enabled == nil || *s.Enabled != desired {
			svc.Schema(schemaName, fivetran.NewConnectionSchemaConfigSchema().Enabled(desired))
		}
	}

	return svc.Do(ctx)
}
