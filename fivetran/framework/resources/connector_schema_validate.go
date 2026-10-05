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
		// The primary key check resolves the connector by group_id + connector_name when possible.
		r.validatePrimaryKeyConstraints(ctx, &data, nil, resp)
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

// validatePrimaryKeyConstraints warns at plan time when configured is_primary_key values can't be applied.
// schemaResponse is optional, it's fetched when nil.
func (r *connectorSchema) validatePrimaryKeyConstraints(
	ctx context.Context,
	data *model.ConnectorSchemaResourceModel,
	schemaResponse *connections.ConnectionSchemaDetailsResponse,
	resp *resource.ValidateConfigResponse) {
	client, err := r.connectorSchemaClient()
	if err != nil {
		return
	}
	// ValidateConfig only sees the configuration, so data has no computed values
	connectorId, service, changed := primaryKeyChangesAfterSync(ctx, client, data, data.ConfiguredPrimaryKeys(), schemaResponse)
	if len(changed) > 0 {
		resp.Diagnostics.AddWarning(
			"Primary Key Configuration After Sync",
			primaryKeyChangeMessage(connectorId, service, changed)+
				" Apply will fail until the configuration matches the current values or the connector is recreated.",
		)
	}
}

// primaryKeyChangesAfterSync returns the connector ID, its service and the columns whose configured is_primary_key differs
// from upstream, when the connector is a file connector that has already synced: primary keys can only be set before the first sync.
// configured must come from the configuration, not the plan: the plan also contains computed values kept from state.
// The connector ID is resolved by group_id + connector_name when connector_id isn't set.
// Any lookup failure returns no changes, the check is skipped in this case.
// schemaResponse is optional, it's fetched when nil.
func primaryKeyChangesAfterSync(
	ctx context.Context,
	client *fivetran.Client,
	data *model.ConnectorSchemaResourceModel,
	configured map[string]map[string]map[string]bool,
	schemaResponse *connections.ConnectionSchemaDetailsResponse) (string, string, []string) {
	if len(configured) == 0 || client == nil {
		return "", "", nil
	}

	connectorId := ""
	if !data.ConnectorId.IsNull() && !data.ConnectorId.IsUnknown() {
		connectorId = data.ConnectorId.ValueString()
	}
	if connectorId == "" {
		if data.GroupId.IsNull() || data.GroupId.IsUnknown() || data.ConnectorName.IsNull() || data.ConnectorName.IsUnknown() {
			return "", "", nil
		}
		foundConnectorId, err := findConnectorIdByGroupAndSchemaName(ctx, client, data)
		if err != nil {
			return "", "", nil
		}
		connectorId = foundConnectorId
	}

	details, err := client.NewConnectionDetails().ConnectionID(connectorId).DoCustom(ctx)
	if err != nil {
		return connectorId, "", nil
	}

	service := details.Data.Service
	if !canChangePrimaryKey(service) || details.Data.SucceededAt.IsZero() {
		return connectorId, service, nil
	}

	if schemaResponse == nil {
		response, err := client.NewConnectionSchemaDetails().ConnectionID(connectorId).Do(ctx)
		if err != nil {
			return connectorId, service, nil
		}
		schemaResponse = &response
	}

	return connectorId, service, changedPrimaryKeys(configured, upstreamPrimaryKeys(ctx, client, connectorId, configured, *schemaResponse))
}

// upstreamPrimaryKeys returns known upstream is_primary_key values for configured columns as schema -> table -> column -> value.
// The schemas response doesn't always contain columns, so columns of configured tables are fetched when missing.
// Columns with no upstream value are omitted.
func upstreamPrimaryKeys(
	ctx context.Context,
	client *fivetran.Client,
	connectorId string,
	configured map[string]map[string]map[string]bool,
	response connections.ConnectionSchemaDetailsResponse) map[string]map[string]map[string]bool {
	result := map[string]map[string]map[string]bool{}
	for schemaName, tables := range configured {
		upstreamSchema, ok := response.Data.Schemas[schemaName]
		if !ok || upstreamSchema == nil {
			continue
		}
		for tableName, columns := range tables {
			upstreamTable, ok := upstreamSchema.Tables[tableName]
			if !ok || upstreamTable == nil {
				continue
			}
			upstreamColumns := upstreamTable.Columns
			for columnName := range columns {
				if _, ok := upstreamColumns[columnName]; !ok {
					columnsResponse, err := client.NewConnectionColumnConfigListService().
						ConnectionId(connectorId).Schema(schemaName).Table(tableName).Do(ctx)
					if err == nil {
						upstreamColumns = columnsResponse.Data.Columns
					}
					break
				}
			}
			for columnName := range columns {
				if upstreamColumn, ok := upstreamColumns[columnName]; ok && upstreamColumn != nil && upstreamColumn.IsPrimaryKey != nil {
					if result[schemaName] == nil {
						result[schemaName] = map[string]map[string]bool{}
					}
					if result[schemaName][tableName] == nil {
						result[schemaName][tableName] = map[string]bool{}
					}
					result[schemaName][tableName][columnName] = *upstreamColumn.IsPrimaryKey
				}
			}
		}
	}
	return result
}

func primaryKeyChangeMessage(connectorId, service string, changed []string) string {
	return fmt.Sprintf(
		"Connector `%v` (service: %v) has already synced, so `is_primary_key` can no longer be changed for:\n  %v\n"+
			"Primary keys for file connectors can only be set before the first sync. "+
			"Replacing `fivetran_connector_schema_config` doesn't help, as it doesn't recreate the connector. "+
			"To apply this change, recreate the connector itself (for example `terraform apply -replace=<fivetran_connector resource address>`), "+
			"which triggers a new initial sync.",
		connectorId, service, strings.Join(changed, "\n  "),
	)
}

// changedPrimaryKeys returns sorted "schema.table.column" entries whose configured is_primary_key differs from upstream.
// Columns with no known upstream value are skipped, there's nothing to compare with.
func changedPrimaryKeys(configured, upstream map[string]map[string]map[string]bool) []string {
	result := []string{}
	for schemaName, tables := range configured {
		for tableName, columns := range tables {
			for columnName, isPk := range columns {
				upstreamIsPk, ok := upstream[schemaName][tableName][columnName]
				if ok && upstreamIsPk != isPk {
					result = append(result, fmt.Sprintf("%v.%v.%v (current: %v, configured: %v)",
						schemaName, tableName, columnName, upstreamIsPk, isPk))
				}
			}
		}
	}
	sort.Strings(result)
	return result
}
