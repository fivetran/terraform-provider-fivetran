package resources

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran/connections"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	configSchema "github.com/fivetran/terraform-provider-fivetran/modules/connection/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ConnectionSchema() resource.Resource {
	return &connectionSchema{}
}

type connectionSchema struct {
	core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &connectionSchema{}
var _ resource.ResourceWithImportState = &connectionSchema{}

func (r *connectionSchema) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection_schema_config"
}

func (r *connectionSchema) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.GetConnectionSchemaResourceSchema(ctx)
}

func (r *connectionSchema) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *connectionSchema) reloadSchema(ctx context.Context, schemaChangeHandling, connectionID string, diag diag.Diagnostics) connections.ConnectionSchemaDetailsResponse {
	client := r.GetClient()
	if client == nil {
		diag.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return connections.ConnectionSchemaDetailsResponse{}
	}

	// Reload schema: we can't update schema if connection doesn't have it yet.
	excludeMode := "PRESERVE"

	// If we want to disable everything by default - we can do it in schema reload
	if schemaChangeHandling == configSchema.BLOCK_ALL {
		excludeMode = "EXCLUDE"
	}

	schemaResponse, err := client.NewConnectionSchemaReload().ExcludeMode(excludeMode).ConnectionID(connectionID).Do(ctx)
	if err != nil {
		diag.AddError(
			"Unable to manage connection schema settings.",
			fmt.Sprintf("Error during schema reloading. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message),
		)
		return connections.ConnectionSchemaDetailsResponse{}
	}
	return schemaResponse
}

func (r *connectionSchema) createNewSchema(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	client := r.GetClient()
	if client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.ConnectionSchemaResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	// Error while reading plan
	if resp.Diagnostics.HasError() {
		return
	}

	// Plan is inconsistent
	if !data.IsValid() {
		resp.Diagnostics.AddError(
			"Unable to Create Connection Schema Resource.",
			"You can use solely one field to define schema settings.",
		)
		return
	}

	connectionID := data.ConnectionId.ValueString()
	schemaChangeHandling := data.SchemaChangeHandling.ValueString()
	localConfig := data.GetSchemaConfig()
	svc := localConfig.PrepareCreateRequest(client.NewConnectionSchemaCreateService()).
		ConnectionID(connectionID).
		SchemaChangeHandling(schemaChangeHandling)

	// we should not parse response here because it will contain only applied diffs, not the whole configuration
	applyResponse, err := svc.Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Connection Schema Resource.",
			fmt.Sprintf("Error while applying schema config patch. %v; code: %v; message: %v", err, applyResponse.Code, applyResponse.Message),
		)
		return
	}
	data.ReadFromResponse(applyResponse, &resp.Diagnostics)
	data.Id = types.StringValue(connectionID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionSchema) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.ConnectionSchemaResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	// Error while reading plan
	if resp.Diagnostics.HasError() {
		return
	}

	// Plan is inconsistent
	if !data.IsValid() {
		resp.Diagnostics.AddError(
			"Unable to Create Connection Schema Resource.",
			"You can use solely one field to define schema settings.",
		)
		return
	}

	connectionID := data.ConnectionId.ValueString()
	schemaChangeHandling := data.SchemaChangeHandling.ValueString()

	client := r.GetClient()

	schemaResponse, err := client.NewConnectionSchemaDetails().ConnectionID(connectionID).Do(ctx)
	// We might have to reload schema in case if there's no schema settings at all, or schema is out of sync with source
	needReload := false
	if err != nil {
		if schemaResponse.Code != "NotFound_SchemaConfig" {
			resp.Diagnostics.AddError(
				"Unable to Create Connection Schema Resource.",
				fmt.Sprintf("Error while retrieving existing schema. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message),
			)
			return
		} else {
			if data.ValidationLevel.ValueString() == "NONE" {
				// create new desired schema
				r.createNewSchema(ctx, req, resp)
				return
			} else {
				// reload because connection doens't have any schema settings yet
				needReload = true
			}
		}
	} else {
		// We might have to refresh schema, not all tables might be saved in current configuration
		err, needReloadSchema := data.ValidateSchemaElements(schemaResponse, *client, ctx)
		if err != nil {
			// Reload as schema might be out of sync with the real source schema
			needReload = needReloadSchema
			if !needReloadSchema {
				resp.Diagnostics.AddError(
					"Unable to create Connection Schema Resource",
					fmt.Sprintf("Column config validation failed. %v", err),
				)
				return
			}
		}
	}

	if needReload {
		schemaResponse = r.reloadSchema(ctx, schemaChangeHandling, connectionID, resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		// validate request one more time after reload schema
		err, _ = data.ValidateSchemaElements(schemaResponse, *client, ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to create Connection Schema Resource.",
				fmt.Sprintf("Schema configuration is not aligned with source schema. Details:\n %v;", err),
			)
			return
		}
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
			"Unable to Create Connection Schema Resource.",
			fmt.Sprintf("Error while applying schema config patch. %v;", err),
		)
		return
	}

	if config.HasUpdates() {
		// applying patch
		svc := config.PrepareRequest(client.NewConnectionSchemaUpdateService())
		svc.ConnectionID(connectionID)
		// update schema_change_handling if needed
		if schemaChangeHandling != "" && schemaChangeHandling != schemaResponse.Data.SchemaChangeHandling {
			svc.SchemaChangeHandling(schemaChangeHandling)
		}
		// we should not parse response here because it will contain only applied diffs, not the whole configuration
		applyResponse, err := svc.Do(ctx)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Connection Schema Resource.",
				fmt.Sprintf("Error while applying schema config patch. %v; code: %v; message: %v", err, applyResponse.Code, applyResponse.Message),
			)
			return
		}
	} else {
		// we update only schema_change_handling if needed
		if schemaChangeHandling != "" && schemaChangeHandling != schemaResponse.Data.SchemaChangeHandling {
			svc := client.NewConnectionSchemaUpdateService().ConnectionID(connectionID)
			svc.SchemaChangeHandling(schemaChangeHandling)
			schResponse, err := svc.Do(ctx)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to Create Connection Schema Resource.",
					fmt.Sprintf("Error while applying schema change handling policy. %v; code: %v; message: %v", err, schResponse.Code, schResponse.Message),
				)
				return
			}
		}
	}

	// We need to re-read schema
	schemaResponse, err = client.NewConnectionSchemaDetails().ConnectionID(connectionID).Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Connection Schema Resource.",
			fmt.Sprintf("Error while reading schema after schema change handling apply. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message),
		)
		return
	}

	// after applying changes it may come that columns weren't saved in table configs, but after switching schema_change_handling - new columns apper in enabled tables.
	// we have to additionally disable them if table has non empty columns configuration

	configAfterApply := configSchema.SchemaConfig{}
	configAfterApply.ReadFromResponse(schemaResponse)

	// apply local config, managing upstream config according to schema change handling policy
	err = configAfterApply.Override(&localConfig, schemaChangeHandling)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Connection Schema Resource.",
			fmt.Sprintf("Error while applying schema config patch. %v.", err),
		)
		return
	}
	if configAfterApply.HasUpdates() {
		svc := configAfterApply.PrepareRequest(client.NewConnectionSchemaUpdateService())
		svc.ConnectionID(connectionID)
		// we should not parse response here because it will contain only applied diffs, not the whole configuration
		applyResponse, err := svc.Do(ctx)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Connection Schema Resource.",
				fmt.Sprintf("Error while applying schema config patch. %v; code: %v; message: %v", err, applyResponse.Code, applyResponse.Message),
			)
			return
		}
	}

	// We need to re-read schema
	schemaResponse, err = client.NewConnectionSchemaDetails().ConnectionID(connectionID).Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Connection Schema Resource.",
			fmt.Sprintf("Error while reading schema after schema change handling apply. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message),
		)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Connection Schema Resource.",
			fmt.Sprintf("Some elements missing in upstream schema. Details: %v", err),
		)
		return
	}

	// read data from response and merge with existing config
	data.ReadFromResponse(schemaResponse, &resp.Diagnostics)
	data.Id = types.StringValue(connectionID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionSchema) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	client := r.GetClient()

	var data model.ConnectionSchemaResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	connectionID := data.ConnectionId.ValueString()

	schemaResponse, err := client.NewConnectionSchemaDetails().ConnectionID(connectionID).Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Connection Schema Resource.",
			fmt.Sprintf("Error while retrieving existing schema. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message),
		)
		return
	}
	data.ReadFromResponse(schemaResponse, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionSchema) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	client := r.GetClient()
	if client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var plan, state model.ConnectionSchemaResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.IsValid() || !state.IsValid() {
		resp.Diagnostics.AddError(
			"Unable to Update Connection Schema Resource.",
			"You can use solely one field to define schema settings.",
		)
		return
	}

	connectionID := state.ConnectionId.ValueString()

	schemaResponse, err := client.NewConnectionSchemaDetails().ConnectionID(connectionID).Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Connection Schema Resource.",
			fmt.Sprintf("Error while retrieving existing schema settings. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message),
		)
		return
	}

	if plan.ValidationLevel.ValueString() != "NONE" {
		// Before applying changes we should validate existing state and planned changes and decide if we need to reload schema
		err, _ := plan.ValidateSchemaElements(schemaResponse, *client, ctx)
		if err != nil {
			schemaResponse = r.reloadSchema(ctx, plan.SchemaChangeHandling.ValueString(), connectionID, resp.Diagnostics)
		}
		err, _ = plan.ValidateSchemaElements(schemaResponse, *client, ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to update Connection Schema Resource.",
				fmt.Sprintf("Schema configuration is not aligned with source schema. Details:\n %v;", err),
			)
			return
		}
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
			"Unable to Create Connection Schema Resource.",
			fmt.Sprintf("Error while applying schema config patch. %v.", err),
		)
		return
	}

	if config.HasUpdates() {
		// applying patch
		svc := config.PrepareRequest(client.NewConnectionSchemaUpdateService())
		svc.ConnectionID(connectionID)
		// update schema_change_handling as well if needed
		if plan.SchemaChangeHandling.String() != "" && plan.SchemaChangeHandling != state.SchemaChangeHandling {
			svc.SchemaChangeHandling(plan.SchemaChangeHandling.ValueString())
		}
		// we should not parse response here because it will contain only applied diffs, not the whole configuration
		applyResponse, err := svc.Do(ctx)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Connection Schema Resource.",
				fmt.Sprintf("Error while applying schema config patch. %v; code: %v; message: %v", err, applyResponse.Code, applyResponse.Message),
			)
			return
		}

	} else {
		// update schema_change_handling if needed
		if plan.SchemaChangeHandling.String() != "" && plan.SchemaChangeHandling != state.SchemaChangeHandling {
			svc := client.NewConnectionSchemaUpdateService().ConnectionID(connectionID)
			svc.SchemaChangeHandling(plan.SchemaChangeHandling.ValueString())
			schResponse, err := svc.Do(ctx)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to Update Connection Schema Resource.",
					fmt.Sprintf("Error while updating schema change handling policy. %v; code: %v; message: %v", err, schResponse.Code, schResponse.Message),
				)
				return
			}
		}
	}

	// re-read schema after apply changes
	schemaResponse, err = client.NewConnectionSchemaDetails().ConnectionID(connectionID).Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Connection Schema Resource.",
			fmt.Sprintf("Error while reading upstream schema. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message),
		)
		return
	}
	// read data from response and merge with existing config

	configAfterApply := configSchema.SchemaConfig{}
	configAfterApply.ReadFromResponse(schemaResponse)

	// apply local config, managing upstream config according to schema change handling policy
	err = configAfterApply.Override(&localConfig, plan.SchemaChangeHandling.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Connection Schema Resource.",
			fmt.Sprintf("Error while applying schema config patch. %v.", err),
		)
		return
	}

	if configAfterApply.HasUpdates() {
		svc := configAfterApply.PrepareRequest(client.NewConnectionSchemaUpdateService())
		svc.ConnectionID(connectionID)
		applyResponse, err := svc.Do(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Connection Schema Resource.",
				fmt.Sprintf("Error while applying schema config patch. %v; code: %v; message: %v", err, applyResponse.Code, applyResponse.Message),
			)
			return
		}
	}

	// re-read schema after apply changes
	schemaResponse, err = client.NewConnectionSchemaDetails().ConnectionID(connectionID).Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Connection Schema Resource.",
			fmt.Sprintf("Error while reading upstream schema. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message),
		)
		return
	}

	plan.ReadFromResponse(schemaResponse, &resp.Diagnostics)
	plan.Id = types.StringValue(connectionID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *connectionSchema) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Nothing to do
}
