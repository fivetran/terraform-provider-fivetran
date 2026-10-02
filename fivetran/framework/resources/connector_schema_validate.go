package resources

import (
	"context"
	"errors"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	configSchema "github.com/fivetran/terraform-provider-fivetran/modules/connector/schema"
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
	needReload := false
	if err != nil {
		if schemaResponse.Code != "NotFound_SchemaConfig" {
			resp.Diagnostics.AddWarning(
				"Unable to Validate Connector Schema Configuration",
				fmt.Sprintf("Unable to retrieve the current schema to validate this configuration at plan time. "+
					"Validation will still run at apply time. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message),
			)
			return
		}
		// No schema captured yet for this connection — reload before validating,
		// same as Create already does at apply time.
		needReload = true
	} else if validateErr, needReloadSchema := data.ValidateSchemaElements(schemaResponse, false, *client, ctx); validateErr != nil {
		// Mismatch against the current schema doesn't necessarily mean the config is
		// wrong — the schema on record may simply be stale (e.g. a table was added at
		// the source after the last reload). Reload and re-check before failing, unless
		// the validation module says reloading wouldn't help (e.g. a genuinely
		// misnamed column) — match Create/Update by failing immediately in that case.
		if !needReloadSchema {
			resp.Diagnostics.AddError(
				"Invalid Connector Schema Resource Configuration.",
				fmt.Sprintf("Schema configuration is not aligned with source schema. Details:\n %v;", validateErr),
			)
			return
		}
		needReload = true
	}

	if needReload {
		schemaResponse = r.reloadSchema(ctx, data.ConnectorId.ValueString(), &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		// Match Create/Update: force column (re-)validation after a reload when
		// schema_change_handling is BLOCK_ALL, so newly-unblocked columns are checked
		// here too, not just on whichever of plan/apply happens to trigger the reload.
		forceValidateColumns := data.SchemaChangeHandling.ValueString() == configSchema.BLOCK_ALL
		if validateErr, _ := data.ValidateSchemaElements(schemaResponse, forceValidateColumns, *client, ctx); validateErr != nil {
			resp.Diagnostics.AddError(
				"Invalid Connector Schema Resource Configuration.",
				fmt.Sprintf("Schema configuration is not aligned with source schema. Details:\n %v;", validateErr),
			)
		}
	}

	// Validate primary key constraints (issue #8)
	r.validatePrimaryKeyConstraints(ctx, &data, resp)
}

func (r *connectorSchema) connectorSchemaClient() (*fivetran.Client, error) {
	client := r.GetClient()
	if client == nil {
		return nil, errUnconfiguredClient
	}
	return client, nil
}

// validatePrimaryKeyConstraints checks if primary key configuration changes are valid.
// Primary keys can only be set before the connector has synced; after syncing, changes require resource replacement.
func (r *connectorSchema) validatePrimaryKeyConstraints(ctx context.Context, data *model.ConnectorSchemaResourceModel, resp *resource.ValidateConfigResponse) {
	if data.ConnectorId.IsNull() || data.ConnectorId.IsUnknown() || data.ConnectorId.ValueString() == "" {
		return // Can't validate without connector ID
	}

	client, err := r.connectorSchemaClient()
	if err != nil {
		return // Skip validation if client unavailable
	}

	// Check if connector has synced
	hasSynced, err := r.hasSynced(ctx, data.ConnectorId.ValueString())
	if err != nil {
		return // Skip validation on error (will be caught at apply time)
	}

	if !hasSynced {
		return // No constraint before first sync
	}

	// Get connector details to check connector type
	connDetails, err := client.NewConnectionDetails().ConnectionID(data.ConnectorId.ValueString()).DoCustom(ctx)
	if err != nil {
		return // Skip validation on error
	}

	connectorType := connDetails.Data.Service
	if !canChangePrimaryKey(connectorType) {
		return // This connector type doesn't support primary key config anyway
	}

	// At this point: connector has synced AND is a file connector that supports primary keys
	// Check if user is trying to configure is_primary_key (which would require replacement)
	resp.Diagnostics.AddWarning(
		"Primary Key Configuration After Sync",
		fmt.Sprintf(
			"You are configuring is_primary_key on a connector that has already synced (connector type: %s). "+
				"Applying this configuration will require destroying and recreating the resource, which will trigger a new sync cycle. "+
				"Consider using `terraform apply -replace` if this is intentional.",
			connectorType,
		),
	)
}
