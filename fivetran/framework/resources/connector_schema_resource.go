package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/connections"
	"github.com/fivetran/go-fivetran/metadata"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	configSchema "github.com/fivetran/terraform-provider-fivetran/modules/connector/schema"
	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// defaultSchemaOperationTimeout bounds the reload+validate+apply sequence in Create/Update
// when the user hasn't set an explicit `timeouts` block. Reload has no server-side polling —
// it's a single HTTP call — but on a large/slow source schema it can still take a while, and
// the provider's HTTP client has no request timeout of its own (see go-fivetran's http.Client),
// so without this, a stalled reload call could hang the whole `terraform apply` indefinitely.
const defaultSchemaOperationTimeout = 30 * time.Minute

func ConnectorSchema() resource.Resource {
	return &connectorSchema{}
}

type connectorSchema struct {
	core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &connectorSchema{}
var _ resource.ResourceWithImportState = &connectorSchema{}

func (r *connectorSchema) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connector_schema_config"
}

func (r *connectorSchema) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.GetConnectorSchemaResourceSchema(ctx)
}

func (r *connectorSchema) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *connectorSchema) reloadSchema(ctx context.Context, connectorID string, currentSchemaResponse connections.ConnectionSchemaDetailsResponse, diag *diag.Diagnostics) connections.ConnectionSchemaDetailsResponse {
	client := r.GetClient()
	if client == nil {
		diag.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return connections.ConnectionSchemaDetailsResponse{}
	}

	// Fetch connector service to check if reload is supported
	connDetails, err := client.NewConnectionDetails().ConnectionID(connectorID).Do(ctx)
	if err == nil {
		service := connDetails.Data.Service
		if service != "" {
			// Check if this connector type supports schema reload
			cache := r.GetMetadataCache()
			connMeta, _ := core.GetCachedConnectorMetadata(ctx, client, cache, service)
			if connMeta != nil && connMeta.SupportsSchemaReload != nil && !*connMeta.SupportsSchemaReload {
				diag.AddWarning(
					"Schema reload not supported for this connection type.",
					"This connection type does not support schema reload. Proceeding with current schema state.",
				)
				// Return empty response — caller will skip validation
				return connections.ConnectionSchemaDetailsResponse{}
			}
		}
	}

	// Reload schema: we can't update schema if connector doesn't have it yet.
	// Some connection types (e.g. UCM) may not support the /reload endpoint.
	excludeMode := "PRESERVE"

	reloadResponse, err := client.NewConnectionSchemaReload().ExcludeMode(excludeMode).ConnectionID(connectorID).Do(ctx)
	if err != nil {
		// Check if reload is not supported for this connection type (not a temporary error)
		if reloadResponse.Code == "NotSupported_SchemaReload" ||
		   reloadResponse.Code == "NotAllowed_SchemaReload" ||
		   reloadResponse.Code == "NotImplemented_SchemaReload" {
			diag.AddWarning(
				"Schema reload not supported for this connection type.",
				fmt.Sprintf("This connection type does not support schema reload. Proceeding with current schema state. Details: %v", reloadResponse.Message),
			)
			// Return empty response — caller must handle gracefully
			return connections.ConnectionSchemaDetailsResponse{}
		}
		diag.AddError(
			"Unable to manage connector schema settings.",
			fmt.Sprintf("Error during schema reloading. %v; code: %v; message: %v", err, reloadResponse.Code, reloadResponse.Message),
		)
		return connections.ConnectionSchemaDetailsResponse{}
	}
	return reloadResponse
}

// primaryKeyFileServices lists file connector service IDs that support configuring is_primary_key before the first sync
var primaryKeyFileServices = map[string]bool{
	"azure_blob_storage":    true,
	"box":                   true,
	"dropbox":               true,
	"email":                 true,
	"ftp":                   true,
	"gcs":                   true,
	"google_drive":          true,
	"google_sheets":         true,
	"s3":                    true,
	"s3_compatible_storage": true,
	"sftp":                  true,
	"share_point":           true,
	"wasabi_cloud_storage":  true,
}

func canChangePrimaryKey(service string) bool {
	return primaryKeyFileServices[service]
}

func findConnectorIdByGroupAndSchemaName(ctx context.Context, client *fivetran.Client, model *model.ConnectorSchemaResourceModel) (string, error) {

	if model.GroupId.IsNull() || model.ConnectorName.IsNull() {
		return "", fmt.Errorf("Either 'connector_id' or 'group_id'+'connector_name' are required to identify a connector (connection).")
	}

	list, err := client.NewConnectionsList().
		GroupID(model.GroupId.ValueString()).
		Schema(model.ConnectorName.ValueString()).
		Do(ctx)

	if err != nil {
		return "", fmt.Errorf("%v; code: %v; message: %v", err, list.Code, list.Message)
	}

	if len(list.Data.Items) == 0 {
		return "", fmt.Errorf("Connector with '%v' group_id and '%v' connector_name doesn't exist.", model.GroupId.ValueString(), model.ConnectorName.ValueString())
	}

	if len(list.Data.Items) > 1 {
		return "", fmt.Errorf("Ambiguous connectors found with '%v' group_id and '%v' connector_name.", model.GroupId.ValueString(), model.ConnectorName.ValueString())
	}

	return list.Data.Items[0].ID, nil
}

func (r *connectorSchema) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.ConnectorSchemaResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	// Error while reading plan
	if resp.Diagnostics.HasError() {
		return
	}

	// Plan is inconsistent
	// TODO(schema-config-plan-validation): now caught earlier by ValidateConfig (see
	// connector_schema_validate.go). Remove this apply-time duplicate once ValidateConfig
	// has been out for a release or two and we're confident it always runs first.
	if !data.IsValid() {
		resp.Diagnostics.AddError(
			"Unable to Create Connector Schema Resource.",
			"You can use solely one field to define schema settings.",
		)
		return
	}

	createTimeout, diags := data.Timeouts.Create(ctx, defaultSchemaOperationTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := helpers.SetContextTimeout(ctx, createTimeout)
	defer cancel()

	var connectorID = data.ConnectorId.ValueString()
	schemaChangeHandling := data.SchemaChangeHandling.ValueString()

	client := r.GetClient()

	if connectorID == "" {
		foundConnectorID, err := findConnectorIdByGroupAndSchemaName(ctx, client, &data)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Connector Schema Resource.",
				fmt.Sprintf("Error while finding connector ID. %v", err),
			)
			return
		}

		connectorID = foundConnectorID
		data.ConnectorId = types.StringValue(foundConnectorID)
	}

	schemaResponse, err := client.NewConnectionSchemaDetails().ConnectionID(connectorID).Do(ctx)
	// We might have to reload schema in case if there's no schema settings at all, or schema is out of sync with source
	needReload := false
	if err != nil {
		if schemaResponse.Code != "NotFound_SchemaConfig" {
			resp.Diagnostics.AddError(
				"Unable to Create Connector Schema Resource.",
				fmt.Sprintf("Error while retrieving existing schema. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message),
			)
			return
		} else {
			// Reload because connector doesn't have any schema settings yet — reload is
			// required even when validation_level is NONE, since PATCH can't update a
			// schema config that doesn't exist yet; reload is what materializes it.
			needReload = true
		}
	} else {
		// We might have to refresh schema, not all tables might be saved in current configuration
		err, needReloadSchema := data.ValidateSchemaElements(schemaResponse, false, *client, ctx)
		if err != nil {
			// Reload as schema might be out of sync with the real source schema
			needReload = needReloadSchema
			if !needReloadSchema {
				resp.Diagnostics.AddError(
					"Unable to create Connector Schema Resource",
					fmt.Sprintf("Column config validation failed. %v", err),
				)
				return
			}
		}
	}

	if needReload {
		schemaResponse = r.reloadSchema(ctx, connectorID, schemaResponse, &resp.Diagnostics)
		// If reload failed with a warning (not supported), skip validation and continue
		// If reload succeeded, validate the reloaded schema
		if resp.Diagnostics.HasError() {
			return
		}
		// Only validate if we actually got a schema back (reload succeeded)
		if len(schemaResponse.Data.Schemas) > 0 {
			forceValidateColumns := schemaChangeHandling == configSchema.BLOCK_ALL
			err, _ = data.ValidateSchemaElements(schemaResponse, forceValidateColumns, *client, ctx)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to create Connector Schema Resource.",
					fmt.Sprintf("Schema configuration is not aligned with source schema. Details:\n %v;", err),
				)
				return
			}
		}
		// If schemaResponse is empty (reload not supported), we proceed with current schema
		// The warning from reloadSchema() informs the user
	}

	// is_primary_key can't be changed after the first sync, fail before sending the change instead of
	// getting an API error or an inconsistent result
	var configData model.ConnectorSchemaResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &configData)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, service, changed := primaryKeyChangesAfterSync(ctx, client, &data, configData.ConfiguredPrimaryKeys(), &schemaResponse); len(changed) > 0 {
		resp.Diagnostics.AddError(
			"Unable to create Connector Schema Resource.",
			primaryKeyChangeMessage(connectorID, service, changed),
		)
		return
	}

	// read upstream config
	config := configSchema.SchemaConfig{}
	config.ReadFromResponse(schemaResponse)

	// read local config
	localConfig := data.GetSchemaConfig()

	// apply local config, managing upstream config according to schema change handling policy
	err = config.Override(&localConfig, schemaChangeHandling)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Connector Schema Resource.",
			fmt.Sprintf("Error while applying schema config patch. %v;", err),
		)
		return
	}

	if config.HasUpdates() {
		// applying patch
		svc := config.PrepareRequest(client.NewConnectionSchemaUpdateService())
		svc.ConnectionID(connectorID)
		// update schema_change_handling if needed
		if schemaChangeHandling != "" && schemaChangeHandling != schemaResponse.Data.SchemaChangeHandling {
			svc.SchemaChangeHandling(schemaChangeHandling)
		}
		// we should not parse response here because it will contain only applied diffs, not the whole configuration
		applyResponse, err := svc.Do(ctx)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Connector Schema Resource.",
				fmt.Sprintf("Error while applying schema config patch. %v; code: %v; message: %v", err, applyResponse.Code, applyResponse.Message),
			)
			return
		}
	} else {
		// we update only schema_change_handling if needed
		if schemaChangeHandling != "" && schemaChangeHandling != schemaResponse.Data.SchemaChangeHandling {
			svc := client.NewConnectionSchemaUpdateService().ConnectionID(connectorID)
			svc.SchemaChangeHandling(schemaChangeHandling)
			schResponse, err := svc.Do(ctx)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to Create Connector Schema Resource.",
					fmt.Sprintf("Error while applying schema change handling policy. %v; code: %v; message: %v", err, schResponse.Code, schResponse.Message),
				)
				return
			}
		}
	}

	// We need to re-read schema
	schemaResponse, err = client.NewConnectionSchemaDetails().ConnectionID(connectorID).Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Connector Schema Resource.",
			fmt.Sprintf("Error while reading schema after schema change handling apply. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message),
		)
		return
	}

	if needReload && schemaChangeHandling == configSchema.BLOCK_ALL && len(schemaResponse.Data.Schemas) > 0 {
		// response doesn't contain columns, need to go through tables and get columns
		// Only validate if reload actually returned schema data (not unsupported)
		err, _ = data.ValidateSchemaElements(schemaResponse, true, *client, ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to create Connector Schema Resource.",
				fmt.Sprintf("Schema configuration is not aligned with source schema after update. Details:\n %v;", err),
			)
			return
		}
	}

	// after applying changes it may come that columns weren't saved in table configs, but after switching schema_change_handling - new columns apper in enabled tables.
	// we have to additionally disable them if table has non empty columns configuration

	configAfterApply := configSchema.SchemaConfig{}
	configAfterApply.ReadFromResponse(schemaResponse)

	// apply local config, managing upstream config according to schema change handling policy
	err = configAfterApply.Override(&localConfig, schemaChangeHandling)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Connector Schema Resource.",
			fmt.Sprintf("Error while applying schema config patch. %v.", err),
		)
		return
	}
	if configAfterApply.HasUpdates() {
		svc := configAfterApply.PrepareRequest(client.NewConnectionSchemaUpdateService())
		svc.ConnectionID(connectorID)
		// we should not parse response here because it will contain only applied diffs, not the whole configuration
		applyResponse, err := svc.Do(ctx)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Connector Schema Resource.",
				fmt.Sprintf("Error while applying schema config patch. %v; code: %v; message: %v", err, applyResponse.Code, applyResponse.Message),
			)
			return
		}

		// We need to re-read schema
		schemaResponse, err = client.NewConnectionSchemaDetails().ConnectionID(connectorID).Do(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Connector Schema Resource.",
				fmt.Sprintf("Error while reading schema after schema change handling apply. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message),
			)
			return
		}
	}

	// read data from response and merge with existing config
	data.ReadFromResponse(schemaResponse, false, &resp.Diagnostics)
	data.Id = types.StringValue(connectorID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectorSchema) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	client := r.GetClient()

	var data model.ConnectorSchemaResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readTimeout, diags := data.Timeouts.Read(ctx, defaultSchemaOperationTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := helpers.SetContextTimeout(ctx, readTimeout)
	defer cancel()

	isImportOperation := data.ConnectorId.IsNull()
	if isImportOperation {
		data.ConnectorId = types.StringValue(data.Id.ValueString())
	}

	connectorID := data.ConnectorId.ValueString()

	schemaResponse, err := client.NewConnectionSchemaDetails().ConnectionID(connectorID).Do(ctx)
	if err != nil {
		if schemaResponse.Code == "NotFound_Connector" || schemaResponse.Code == "NotFound_Connection" || schemaResponse.Code == "NotFound_SchemaConfig" {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Connector Schema Resource.",
			fmt.Sprintf("Error while retrieving existing schema. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message),
		)
		return
	}
	data.ReadFromResponse(schemaResponse, isImportOperation, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectorSchema) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	client := r.GetClient()
	if client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var plan, state model.ConnectorSchemaResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO(schema-config-plan-validation): the plan.IsValid() half is now caught earlier by
	// ValidateConfig (see connector_schema_validate.go) and can be dropped once that's been out
	// for a release or two. state.IsValid() still needs to stay — ValidateConfig only sees the
	// new config (req.Config), not prior applied state.
	if !plan.IsValid() || !state.IsValid() {
		resp.Diagnostics.AddError(
			"Unable to Update Connector Schema Resource.",
			"You can use solely one field to define schema settings.",
		)
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, defaultSchemaOperationTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := helpers.SetContextTimeout(ctx, updateTimeout)
	defer cancel()

	var configData model.ConnectorSchemaResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &configData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// the connector may differ from state when it was recreated, the addressing mode comes from the configuration:
	// connector_id is computed, so the plan alone doesn't tell whether it was configured
	connectorID := state.ConnectorId.ValueString()
	if !configData.ConnectorId.IsNull() && !plan.ConnectorId.IsUnknown() && plan.ConnectorId.ValueString() != "" {
		connectorID = plan.ConnectorId.ValueString()
	} else if !configData.GroupId.IsNull() && !configData.ConnectorName.IsNull() {
		foundConnectorID, err := findConnectorIdByGroupAndSchemaName(ctx, client, &plan)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update Connector Schema Resource.",
				fmt.Sprintf("Error while finding connector ID. %v", err),
			)
			return
		}
		connectorID = foundConnectorID
	}
	plan.ConnectorId = types.StringValue(connectorID)

	schemaResponse, err := client.NewConnectionSchemaDetails().ConnectionID(connectorID).Do(ctx)
	forceColumnsPopulationAfterSchemaReloaded := false
	schemaReloaded := false
	if err != nil {
		if schemaResponse.Code != "NotFound_SchemaConfig" {
			resp.Diagnostics.AddError(
				"Unable to Update Connector Schema Resource.",
				fmt.Sprintf("Error while retrieving existing schema settings. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message),
			)
			return
		}
		// Match Create: a recreated connector may have no schema settings yet, reload materializes them
		// (required even when validation_level is NONE, PATCH can't update a schema config that doesn't exist)
		schemaResponse = r.reloadSchema(ctx, connectorID, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		schemaReloaded = true
		forceColumnsPopulationAfterSchemaReloaded = (plan.SchemaChangeHandling.ValueString() == configSchema.BLOCK_ALL)
	}

	if schemaReloaded && plan.ValidationLevel.ValueString() != "NONE" {
		err, _ = plan.ValidateSchemaElements(schemaResponse, forceColumnsPopulationAfterSchemaReloaded, *client, ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to update Connector Schema Resource.",
				fmt.Sprintf("Schema configuration is not aligned with source schema. Details:\n %v;", err),
			)
			return
		}
	} else if plan.ValidationLevel.ValueString() != "NONE" {
		// Before applying changes we should validate existing state and planned changes and decide if we need to reload schema
		err, needReloadSchema := plan.ValidateSchemaElements(schemaResponse, false, *client, ctx)
		if err != nil {
			// Match Create: only reload if the validation module says reloading could
			// help. A validation error that isn't reload-fixable (e.g. a genuinely
			// misnamed column) should fail immediately instead of wasting a reload call.
			if !needReloadSchema {
				resp.Diagnostics.AddError(
					"Unable to Update Connector Schema Resource",
					fmt.Sprintf("Column config validation failed. %v", err),
				)
				return
			}
			schemaResponse = r.reloadSchema(ctx, connectorID, schemaResponse, &resp.Diagnostics)
			if resp.Diagnostics.HasError() {
				return
			}
			// Only validate if reload succeeded (returned schema data)
			if len(schemaResponse.Data.Schemas) > 0 {
				forceColumnsPopulationAfterSchemaReloaded = (plan.SchemaChangeHandling.ValueString() == configSchema.BLOCK_ALL)

				err, _ = plan.ValidateSchemaElements(schemaResponse, forceColumnsPopulationAfterSchemaReloaded, *client, ctx)
				if err != nil {
					resp.Diagnostics.AddError(
						"Unable to update Connector Schema Resource.",
						fmt.Sprintf("Schema configuration is not aligned with source schema. Details:\n %v;", err),
					)
					return
				}
			}
		}
	}

	// is_primary_key can't be changed after the first sync, fail before sending the change instead of
	// getting an API error or an inconsistent result
	if _, service, changed := primaryKeyChangesAfterSync(ctx, client, &plan, configData.ConfiguredPrimaryKeys(), &schemaResponse); len(changed) > 0 {
		resp.Diagnostics.AddError(
			"Unable to Update Connector Schema Resource.",
			primaryKeyChangeMessage(connectorID, service, changed),
		)
		return
	}

	// read upstream config
	config := configSchema.SchemaConfig{}
	config.ReadFromResponse(schemaResponse)

	// read local config
	localConfig := plan.GetSchemaConfig()

	// apply local config, managing upstream config according to schema change handling policy
	err = config.Override(&localConfig, plan.SchemaChangeHandling.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Connector Schema Resource.",
			fmt.Sprintf("Error while applying schema config patch. %v.", err),
		)
		return
	}

	if config.HasUpdates() {
		// applying patch
		svc := config.PrepareRequest(client.NewConnectionSchemaUpdateService())
		svc.ConnectionID(connectorID)
		// update schema_change_handling as well if needed
		if plan.SchemaChangeHandling.ValueString() != "" && plan.SchemaChangeHandling.ValueString() != schemaResponse.Data.SchemaChangeHandling {
			svc.SchemaChangeHandling(plan.SchemaChangeHandling.ValueString())
		}
		// we should not parse response here because it will contain only applied diffs, not the whole configuration
		applyResponse, err := svc.Do(ctx)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Connector Schema Resource.",
				fmt.Sprintf("Error while applying schema config patch. %v; code: %v; message: %v", err, applyResponse.Code, applyResponse.Message),
			)
			return
		}

	} else {
		// update schema_change_handling if needed
		if plan.SchemaChangeHandling.ValueString() != "" && plan.SchemaChangeHandling.ValueString() != schemaResponse.Data.SchemaChangeHandling {
			svc := client.NewConnectionSchemaUpdateService().ConnectionID(connectorID)
			svc.SchemaChangeHandling(plan.SchemaChangeHandling.ValueString())
			schResponse, err := svc.Do(ctx)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to Update Connector Schema Resource.",
					fmt.Sprintf("Error while updating schema change handling policy. %v; code: %v; message: %v", err, schResponse.Code, schResponse.Message),
				)
				return
			}
		}
	}

	// re-read schema after apply changes
	schemaResponse, err = client.NewConnectionSchemaDetails().ConnectionID(connectorID).Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Connector Schema Resource.",
			fmt.Sprintf("Error while reading upstream schema. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message),
		)
		return
	}
	// read data from response and merge with existing config

	if forceColumnsPopulationAfterSchemaReloaded {
		// response doesn't contain columns, need to go through tables and get columns
		err, _ = plan.ValidateSchemaElements(schemaResponse, forceColumnsPopulationAfterSchemaReloaded, *client, ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to update Connector Schema Resource.",
				fmt.Sprintf("Schema configuration is not aligned with source schema after update. Details:\n %v;", err),
			)
			return
		}
	}

	configAfterApply := configSchema.SchemaConfig{}
	configAfterApply.ReadFromResponse(schemaResponse)

	// apply local config, managing upstream config according to schema change handling policy
	err = configAfterApply.Override(&localConfig, plan.SchemaChangeHandling.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Connector Schema Resource.",
			fmt.Sprintf("Error while applying schema config patch. %v.", err),
		)
		return
	}

	if configAfterApply.HasUpdates() {
		svc := configAfterApply.PrepareRequest(client.NewConnectionSchemaUpdateService())
		svc.ConnectionID(connectorID)
		applyResponse, err := svc.Do(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Connector Schema Resource.",
				fmt.Sprintf("Error while applying schema config patch. %v; code: %v; message: %v", err, applyResponse.Code, applyResponse.Message),
			)
			return
		}
	}

	// re-read schema after apply changes
	schemaResponse, err = client.NewConnectionSchemaDetails().ConnectionID(connectorID).Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Connector Schema Resource.",
			fmt.Sprintf("Error while reading upstream schema. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message),
		)
		return
	}

	plan.ReadFromResponse(schemaResponse, false, &resp.Diagnostics)
	plan.Id = types.StringValue(connectorID)
	plan.ConnectorId = types.StringValue(connectorID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *connectorSchema) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Nothing to do
}
