package schema

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/connections"
)

const multipleTableColumnsBatchSize = 100

func getMultipleTableColumnsConfig(
	ctx context.Context,
	client fivetran.Client,
	connectionID,
	schemaName string,
	tables []string,
) (connections.MultipleTableColumnsConfigResponse, error) {
	return client.
		NewFetchSourceColumnsService().
		ConnectionId(connectionID).
		Schema(schemaName).
		Tables(tables).
		Do(ctx)
}

// FetchTablesColumns returns columns config of the given tables of a schema as table -> columns.
// Tables are requested in batches; per-table requests are used only if the batch endpoint is not implemented (NotImplemented_*).
// NotFound_* (resource-not-found), rate limit, auth, timeout, and other errors propagate
// immediately without fallback to avoid masking the real issue or amplifying failures.
func FetchTablesColumns(
	ctx context.Context,
	client fivetran.Client,
	connectorId,
	sName string,
	tables []string,
) (map[string]map[string]*connections.ConnectionSchemaConfigColumnResponse, error) {
	result := map[string]map[string]*connections.ConnectionSchemaConfigColumnResponse{}
	var batchError error
	for start := 0; start < len(tables); start += multipleTableColumnsBatchSize {
		end := start + multipleTableColumnsBatchSize
		if end > len(tables) {
			end = len(tables)
		}

		response, err := getMultipleTableColumnsConfig(ctx, client, connectorId, sName, tables[start:end])
		if err != nil {
			batchError = err
			if isEndpointUnavailableError(response.Code) {
				break
			}
			return nil, fmt.Errorf("Error while retrieving columns config for schema `%s`. Error: %v; Code: `%v`.",
				sName, err, response.Code)
		}
		for tableName, tableData := range response.Data.Tables {
			result[tableName] = tableData.Columns
		}
	}

	if batchError != nil {
		for _, tName := range tables {
			response, err := client.NewConnectionColumnConfigListService().
				ConnectionId(connectorId).
				Schema(sName).
				Table(tName).
				Do(ctx)
			if err != nil {
				return nil, fmt.Errorf("Error while retrieving columns config for table `%s` of schema `%s`. Error: %v; Code: `%v`.",
					tName, sName, err, response.Code)
			}
			result[tName] = response.Data.Columns
		}
	}
	return result, nil
}
