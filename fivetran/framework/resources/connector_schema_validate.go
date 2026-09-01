package resources

import (
	"context"
	"errors"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.ResourceWithValidateConfig = &connectorSchema{}

func (r *connectorSchema) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data model.ConnectorSchemaResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.IsValid() {
		resp.Diagnostics.AddError(
			"Invalid Connector Schema Resource Configuration.",
			"You can use solely one field to define schema settings.",
		)
		return
	}

	if data.ValidationLevel.ValueString() == "NONE" {
		resp.Diagnostics.AddWarning(
			"Schema Validation Disabled",
			"validation_level is NONE — table and column names in this configuration will not be checked "+
				"against the actual source schema. If they don't match, they may be silently ignored or cause "+
				"unexpected sync behavior (for example, a column you intend to hash may not be hashed if the "+
				"name doesn't match the source).",
		)
		return
	}

	if data.ConnectorId.IsNull() || data.ConnectorId.IsUnknown() || data.ConnectorId.ValueString() == "" {
		// Connector isn't known yet at plan time (e.g. referenced from another resource
		// being created in the same apply) — nothing to validate against yet.
		return
	}

	client, err := r.connectorSchemaClient()
	if err != nil {
		if errors.Is(err, errUnconfiguredClient) {
			return
		}
		resp.Diagnostics.AddWarning(
			"Unable to Validate Connector Schema Configuration",
			fmt.Sprintf("Unable to access the provider client to validate this configuration at plan time. "+
				"Validation will still run at apply time. Original error: %v", err),
		)
		return
	}

	schemaResponse, err := client.NewConnectionSchemaDetails().ConnectionID(data.ConnectorId.ValueString()).Do(ctx)
	if err != nil {
		if schemaResponse.Code == "NotFound_SchemaConfig" {
			// No schema captured yet for this connection — nothing to validate against
			// until it's reloaded, which only happens at apply time.
			return
		}
		resp.Diagnostics.AddWarning(
			"Unable to Validate Connector Schema Configuration",
			fmt.Sprintf("Unable to retrieve the current schema to validate this configuration at plan time. "+
				"Validation will still run at apply time. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message),
		)
		return
	}

	if validateErr, _ := data.ValidateSchemaElements(schemaResponse, false, *client, ctx); validateErr != nil {
		resp.Diagnostics.AddError(
			"Invalid Connector Schema Resource Configuration.",
			fmt.Sprintf("Schema configuration is not aligned with source schema. Details:\n %v;", validateErr),
		)
	}
}

func (r *connectorSchema) connectorSchemaClient() (*fivetran.Client, error) {
	client := r.GetClient()
	if client == nil {
		return nil, errUnconfiguredClient
	}
	return client, nil
}
