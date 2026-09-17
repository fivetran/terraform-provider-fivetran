package schema

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/common"
	"github.com/fivetran/go-fivetran/connections"
	httputils "github.com/fivetran/go-fivetran/http_utils"
)

const multipleTableColumnsBatchSize = 100

type multipleTableColumnsConfigResponse struct {
	common.CommonResponse
	Data struct {
		Tables map[string]tableColumnsConfigResponse `json:"tables"`
	} `json:"data"`
}

type tableColumnsConfigResponse struct {
	Columns map[string]*connections.ConnectionSchemaConfigColumnResponse `json:"columns"`
}

func getMultipleTableColumnsConfig(
	ctx context.Context,
	client fivetran.Client,
	connectionID,
	schemaName string,
	tables []string,
) (multipleTableColumnsConfigResponse, error) {
	var response multipleTableColumnsConfigResponse
	httpService := client.NewHttpService()

	requestURL := fmt.Sprintf(
		"%s/connections/%s/schemas/%s/fetch-source-columns",
		httpService.BaseUrl,
		connectionID,
		schemaName,
	)

	requestBody := map[string]interface{}{
		"tables": tables,
	}

	reqBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return response, err
	}

	respBody, respStatus, err := (&httputils.Request{
		Method:           "POST",
		Url:              requestURL,
		Headers:          httpService.CommonHeaders,
		Body:             reqBodyJSON,
		Client:           httpService.Client,
		HandleRateLimits: httpService.HandleRateLimits,
		MaxRetryAttempts: httpService.MaxRetryAttempts,
	}).Do(ctx)
	if err != nil {
		return response, err
	}

	if err := json.Unmarshal(respBody, &response); err != nil {
		return response, err
	}

	if respStatus != 200 {
		return response, fmt.Errorf("status code: %v; expected: 200", respStatus)
	}

	return response, nil
}
