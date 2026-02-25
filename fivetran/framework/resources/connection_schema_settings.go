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

func connectionSchemaSettingsSchema() schema.Schema {
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

type connectionSchemaSettingsModel struct {
	Id                   types.String                      `tfsdk:"id"`
	ConnectionId         types.String                      `tfsdk:"connection_id"`
	SchemaChangeHandling types.String                      `tfsdk:"schema_change_handling"`
	DisabledSchemas      fivetrantypes.FastStringSetValue `tfsdk:"disabled_schemas"`
	EnabledSchemas       fivetrantypes.FastStringSetValue `tfsdk:"enabled_schemas"`
}

// readFromResponse populates the model from the API response.
// When the policy is ALLOW_ALL, disabled_schemas is populated with schemas where enabled==false.
// When the policy is BLOCK_ALL or ALLOW_COLUMNS, enabled_schemas is populated with schemas where enabled==true.
// The irrelevant set is set to null in each case.
func (d *connectionSchemaSettingsModel) readFromResponse(resp connections.ConnectionSchemaDetailsResponse) {
	d.SchemaChangeHandling = types.StringValue(resp.Data.SchemaChangeHandling)

	if resp.Data.SchemaChangeHandling == "ALLOW_ALL" {
		names := collectSchemaNames(resp, false)
		d.DisabledSchemas = buildOrderedSet(names, d.DisabledSchemas)
		d.EnabledSchemas = fivetrantypes.NewFastStringSetNull()
	} else {
		names := collectSchemaNames(resp, true)
		d.EnabledSchemas = buildOrderedSet(names, d.EnabledSchemas)
		d.DisabledSchemas = fivetrantypes.NewFastStringSetNull()
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

func ConnectionSchemaSettings() resource.Resource {
	return &connectionSchemaSettings{}
}

type connectionSchemaSettings struct {
	core.ProviderResource
}

var _ resource.ResourceWithConfigure = &connectionSchemaSettings{}
var _ resource.ResourceWithImportState = &connectionSchemaSettings{}

func (r *connectionSchemaSettings) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection_schema_settings"
}

func (r *connectionSchemaSettings) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = connectionSchemaSettingsSchema()
}

func (r *connectionSchemaSettings) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *connectionSchemaSettings) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError("Unconfigured Fivetran Client", "Please report this issue to the provider developers.")
		return
	}

	var data connectionSchemaSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diagMsg := validatePlanConsistency(data); diagMsg != "" {
		resp.Diagnostics.AddError("Invalid Configuration", diagMsg)
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

func (r *connectionSchemaSettings) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError("Unconfigured Fivetran Client", "Please report this issue to the provider developers.")
		return
	}

	var data connectionSchemaSettingsModel
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

func (r *connectionSchemaSettings) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError("Unconfigured Fivetran Client", "Please report this issue to the provider developers.")
		return
	}

	var plan, state connectionSchemaSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diagMsg := validatePlanConsistency(plan); diagMsg != "" {
		resp.Diagnostics.AddError("Invalid Configuration", diagMsg)
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

func (r *connectionSchemaSettings) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No-op: schema settings always exist on a connection.
}

// --- Helpers ---

// validatePlanConsistency checks that the user specified the correct schema
// list attribute for the chosen policy. Returns an error message if
// enabled_schemas is used with ALLOW_ALL (should use disabled_schemas) or
// disabled_schemas is used with BLOCK_ALL/ALLOW_COLUMNS (should use enabled_schemas).
// Returns an empty string when the configuration is valid.
func validatePlanConsistency(data connectionSchemaSettingsModel) string {
	policy := data.SchemaChangeHandling.ValueString()
	if policy == "ALLOW_ALL" && !data.EnabledSchemas.IsNull() {
		return "enabled_schemas cannot be used with schema_change_handling = ALLOW_ALL. Use disabled_schemas instead."
	}
	if policy != "ALLOW_ALL" && !data.DisabledSchemas.IsNull() {
		return fmt.Sprintf("disabled_schemas cannot be used with schema_change_handling = %s. Use enabled_schemas instead.", policy)
	}
	return ""
}

// getUserSchemaNames extracts schema names from whichever list attribute is
// set (disabled_schemas or enabled_schemas) and returns them as a Go map for
// O(1) membership checks. Returns an empty map if neither attribute is set.
func getUserSchemaNames(data connectionSchemaSettingsModel) map[string]bool {
	result := map[string]bool{}
	var setToRead fivetrantypes.FastStringSetValue
	if !data.DisabledSchemas.IsNull() && !data.DisabledSchemas.IsUnknown() {
		setToRead = data.DisabledSchemas
	} else if !data.EnabledSchemas.IsNull() && !data.EnabledSchemas.IsUnknown() {
		setToRead = data.EnabledSchemas
	} else {
		return result
	}
	for _, elem := range setToRead.Elements() {
		if strVal, ok := elem.(types.String); ok {
			result[strVal.ValueString()] = true
		}
	}
	return result
}

// computeDesiredEnabled determines whether a schema should be enabled based on
// the schema_change_handling policy and the user's schema list.
// ALLOW_ALL: schemas are enabled by default; those in userSet are disabled.
// BLOCK_ALL/ALLOW_COLUMNS: schemas are disabled by default; those in userSet are enabled.
func computeDesiredEnabled(policy string, schemaName string, userSet map[string]bool) bool {
	if policy == "ALLOW_ALL" {
		_, inSet := userSet[schemaName]
		return !inSet
	}
	_, inSet := userSet[schemaName]
	return inSet
}

// applySchemaSettings sends a PATCH request to update the connection's
// schema_change_handling policy and per-schema enabled state. It iterates
// over all schemas present in the API response and sets each to its desired
// enabled state based on the policy and user's list. Schemas not present
// in the API response (e.g. stale references in config) are silently skipped.
// Returns the raw error from the API call (nil on success).
func applySchemaSettings(
	ctx context.Context,
	client *fivetran.Client,
	connectionId string,
	data connectionSchemaSettingsModel,
	schemaResp connections.ConnectionSchemaDetailsResponse,
) (connections.ConnectionSchemaDetailsResponse, error) {
	policy := data.SchemaChangeHandling.ValueString()
	userSet := getUserSchemaNames(data)

	svc := client.NewConnectionSchemaUpdateService().ConnectionID(connectionId)
	svc.SchemaChangeHandling(policy)

	for schemaName, s := range schemaResp.Data.Schemas {
		desired := computeDesiredEnabled(policy, schemaName, userSet)
		if s.Enabled == nil || *s.Enabled != desired {
			svc.Schema(schemaName, fivetran.NewConnectionSchemaConfigSchema().Enabled(desired))
		}
	}

	return svc.Do(ctx)
}
