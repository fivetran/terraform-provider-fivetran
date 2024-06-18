package resources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	configSchema "github.com/fivetran/terraform-provider-fivetran/modules/connector/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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

	if !data.IsValid() {
		resp.Diagnostics.AddError(
			"Unable to Create Connector Schema Resource.",
			"You can use solely one field to define schema settings.",
		)
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	connectorID := data.ConnectorId.ValueString()
	schemaChangeHandling := data.SchemaChangeHandling.ValueString()

	client := r.GetClient()

	schemaResponse, err := client.NewConnectorSchemaDetails().ConnectorID(connectorID).Do(ctx)

	if err != nil {
		if schemaResponse.Code != "NotFound_SchemaConfig" {
			resp.Diagnostics.AddError(
				"Unable to Create Connector Schema Resource.",
				fmt.Sprintf("Error wile retrieving existing schema. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message),
			)
			return
		} else {
			// Reload schema: we can't update schema if connector doesn't have it yet.
			excludeMode := "PRESERVE"

			// If we want to disable everything by default - we can do it in schema reload
			if schemaChangeHandling == configSchema.BLOCK_ALL {
				excludeMode = "EXCLUDE"
			}

			schemaResponse, err = client.NewConnectorSchemaReload().ExcludeMode(excludeMode).ConnectorID(connectorID).Do(ctx)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to Create Connector Schema Resource.",
					fmt.Sprintf("Error wile schema reload. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message),
				)
				return
			}
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
			"Unable to Create Connector Schema Resource.",
			fmt.Sprintf("Error wile applying schema config patch. %v; Please report this issue to the provider developers.", err),
		)
		return
	}

	if config.HasUpdates() {
		// applying patch
		svc := config.PrepareRequest(client.NewConnectorSchemaUpdateService())
		svc.ConnectorID(connectorID)
		// update schema_change_handling if needed
		if schemaChangeHandling != schemaResponse.Data.SchemaChangeHandling {
			svc.SchemaChangeHandling(schemaChangeHandling)
		}
		// we should not parse response here because it will contain only applied diffs, not the whole configuration
		applyResponse, err := svc.Do(ctx)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Connector Schema Resource.",
				fmt.Sprintf("Error wile applying schema config patch. %v; code: %v; message: %v", err, applyResponse.Code, applyResponse.Message),
			)
			return
		}
	} else {
		// we update only schema_change_handling if needed
		if schemaChangeHandling != schemaResponse.Data.SchemaChangeHandling {
			svc := client.NewConnectorSchemaUpdateService().ConnectorID(connectorID)
			svc.SchemaChangeHandling(schemaChangeHandling)
			schResponse, err := svc.Do(ctx)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to Create Connector Schema Resource.",
					fmt.Sprintf("Error wile applying schema change handling policy. %v; code: %v; message: %v", err, schResponse.Code, schResponse.Message),
				)
				return
			}
		}
	}

	// We need to re-read schema
	schemaResponse, err = client.NewConnectorSchemaDetails().ConnectorID(connectorID).Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Connector Schema Resource.",
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
			"Unable to Create Connector Schema Resource.",
			fmt.Sprintf("Error wile applying schema config patch. %v; Please report this issue to the provider developers.", err),
		)
		return
	}
	if configAfterApply.HasUpdates() {
		svc := configAfterApply.PrepareRequest(client.NewConnectorSchemaUpdateService())
		svc.ConnectorID(connectorID)
		// we should not parse response here because it will contain only applied diffs, not the whole configuration
		applyResponse, err := svc.Do(ctx)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Connector Schema Resource.",
				fmt.Sprintf("Error wile applying schema config patch. %v; code: %v; message: %v", err, applyResponse.Code, applyResponse.Message),
			)
			return
		}
	}

	// We need to re-read schema
	schemaResponse, err = client.NewConnectorSchemaDetails().ConnectorID(connectorID).Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Connector Schema Resource.",
			fmt.Sprintf("Error while reading schema after schema change handling apply. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message),
		)
		return
	}

	// read data from response and merge with existing config
	data.ReadFromResponse(schemaResponse)

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

	connectorID := data.ConnectorId.ValueString()

	schemaResponse, err := client.NewConnectorSchemaDetails().ConnectorID(connectorID).Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Connector Schema Resource.",
			fmt.Sprintf("Error wile retrieving existing schema. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message),
		)
		return
	}

	data.ReadFromResponse(schemaResponse)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectorSchema) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	client := r.GetClient()

	var plan, state model.ConnectorSchemaResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if !plan.IsValid() || !state.IsValid() {
		resp.Diagnostics.AddError(
			"Unable to Update Connector Schema Resource.",
			"You can use solely one field to define schema settings.",
		)
		return
	}

	connectorID := state.ConnectorId.ValueString()

	schemaResponse, err := client.NewConnectorSchemaDetails().ConnectorID(connectorID).Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Connector Schema Resource.",
			fmt.Sprintf("Error while reading upstream schema. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message),
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
			fmt.Sprintf("Error wile applying schema config patch. %v; Please report this issue to the provider developers.", err),
		)
		return
	}

	if config.HasUpdates() {
		// applying patch
		svc := config.PrepareRequest(client.NewConnectorSchemaUpdateService())
		svc.ConnectorID(connectorID)
		// update schema_change_handling as well if needed
		if plan.SchemaChangeHandling != state.SchemaChangeHandling {
			svc.SchemaChangeHandling(plan.SchemaChangeHandling.ValueString())
		}
		// we should not parse response here because it will contain only applied diffs, not the whole configuration
		applyResponse, err := svc.Do(ctx)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Connector Schema Resource.",
				fmt.Sprintf("Error wile applying schema config patch. %v; code: %v; message: %v", err, applyResponse.Code, applyResponse.Message),
			)
			return
		}

	} else {
		// update schema_change_handling if needed
		if plan.SchemaChangeHandling != state.SchemaChangeHandling {
			svc := client.NewConnectorSchemaUpdateService().ConnectorID(connectorID)
			svc.SchemaChangeHandling(plan.SchemaChangeHandling.ValueString())
			schResponse, err := svc.Do(ctx)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to Update Connector Schema Resource.",
					fmt.Sprintf("Error wile updating schema change handling policy. %v; code: %v; message: %v", err, schResponse.Code, schResponse.Message),
				)
				return
			}
		}
	}

	// re-read schema after apply changes
	schemaResponse, err = client.NewConnectorSchemaDetails().ConnectorID(connectorID).Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Connector Schema Resource.",
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
			"Unable to Update Connector Schema Resource.",
			fmt.Sprintf("Error wile applying schema config patch. %v; Please report this issue to the provider developers.", err),
		)
		return
	}

	if configAfterApply.HasUpdates() {
		svc := configAfterApply.PrepareRequest(client.NewConnectorSchemaUpdateService())
		svc.ConnectorID(connectorID)
		applyResponse, err := svc.Do(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Connector Schema Resource.",
				fmt.Sprintf("Error wile applying schema config patch. %v; code: %v; message: %v", err, applyResponse.Code, applyResponse.Message),
			)
			return
		}
	}

	// re-read schema after apply changes
	schemaResponse, err = client.NewConnectorSchemaDetails().ConnectorID(connectorID).Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Connector Schema Resource.",
			fmt.Sprintf("Error while reading upstream schema. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message),
		)
		return
	}

	plan.ReadFromResponse(schemaResponse)
	plan.Id = types.StringValue(connectorID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *connectorSchema) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Nothing to do
}
