package schema

import (
	"context"

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
