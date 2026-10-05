package resources

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/connections"
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
		// primary key constraints don't depend on name validation
		r.validatePrimaryKeyConstraints(ctx, &data, nil, resp)
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

	r.validatePrimaryKeyConstraints(ctx, &data, &schemaResponse, resp)
}

func (r *connectorSchema) connectorSchemaClient() (*fivetran.Client, error) {
	client := r.GetClient()
	if client == nil {
		return nil, errUnconfiguredClient
	}
	return client, nil
}

// validatePrimaryKeyConstraints warns when configured is_primary_key values differ from upstream on a file connector
// that has already synced: primary keys can only be set before the first sync.
// schemaResponse is optional, it's fetched when nil.
func (r *connectorSchema) validatePrimaryKeyConstraints(
	ctx context.Context,
	data *model.ConnectorSchemaResourceModel,
	schemaResponse *connections.ConnectionSchemaDetailsResponse,
	resp *resource.ValidateConfigResponse) {
	configured := data.ConfiguredPrimaryKeys()
	if len(configured) == 0 {
		return
	}

	if data.ConnectorId.IsNull() || data.ConnectorId.IsUnknown() || data.ConnectorId.ValueString() == "" {
		return
	}
	connectorId := data.ConnectorId.ValueString()

	client, err := r.connectorSchemaClient()
	if err != nil {
		return
	}

	// any lookup failure skips this check, it's advisory only
	details, err := client.NewConnectionDetails().ConnectionID(connectorId).DoCustom(ctx)
	if err != nil {
		return
	}

	service := details.Data.Service
	if !canChangePrimaryKey(service) || details.Data.SucceededAt.IsZero() {
		return
	}

	if schemaResponse == nil {
		response, err := client.NewConnectionSchemaDetails().ConnectionID(connectorId).Do(ctx)
		if err != nil {
			return
		}
		schemaResponse = &response
	}

	changed := changedPrimaryKeys(configured, *schemaResponse)
	if len(changed) == 0 {
		return
	}

	resp.Diagnostics.AddWarning(
		"Primary Key Configuration After Sync",
		fmt.Sprintf(
			"Connector `%v` (service: %v) has already synced, so `is_primary_key` can no longer be changed for:\n  %v\n"+
				"Primary keys for file connectors can only be set before the first sync. "+
				"Replacing `fivetran_connector_schema_config` doesn't help, as it doesn't recreate the connector. "+
				"To apply this change, recreate the connector itself (for example `terraform apply -replace=<fivetran_connector resource address>`), "+
				"which triggers a new initial sync.",
			connectorId, service, strings.Join(changed, "\n  "),
		),
	)
}

// changedPrimaryKeys returns sorted "schema.table.column" entries whose configured is_primary_key differs from upstream.
// Columns unknown upstream are skipped, there's nothing to compare with.
func changedPrimaryKeys(configured map[string]map[string]map[string]bool, upstream connections.ConnectionSchemaDetailsResponse) []string {
	result := []string{}
	for schemaName, tables := range configured {
		upstreamSchema, ok := upstream.Data.Schemas[schemaName]
		if !ok || upstreamSchema == nil {
			continue
		}
		for tableName, columns := range tables {
			upstreamTable, ok := upstreamSchema.Tables[tableName]
			if !ok || upstreamTable == nil {
				continue
			}
			for columnName, isPk := range columns {
				upstreamColumn, ok := upstreamTable.Columns[columnName]
				if !ok || upstreamColumn == nil || upstreamColumn.IsPrimaryKey == nil {
					continue
				}
				if *upstreamColumn.IsPrimaryKey != isPk {
					result = append(result, fmt.Sprintf("%v.%v.%v (current: %v, configured: %v)",
						schemaName, tableName, columnName, *upstreamColumn.IsPrimaryKey, isPk))
				}
			}
		}
	}
	sort.Strings(result)
	return result
}
