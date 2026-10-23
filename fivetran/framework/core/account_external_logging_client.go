package core

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/common"
	"github.com/fivetran/go-fivetran/metadata"
)

// The public API serves the account-level external logging endpoints, but the go-fivetran SDK
// (latest release v1.3.13) has no services for them nor for /metadata/log-service-types, so this
// file calls them through the SDK's generic HttpService. Replace it with the SDK services once
// they are released.
const accountExternalLoggingPath = "/external-logging/account"

type AccountExternalLoggingData struct {
	Id         string                     `json:"id"`
	Service    string                     `json:"service"`
	Enabled    bool                       `json:"enabled"`
	Config     map[string]interface{}     `json:"config"`
	SetupTests []common.SetupTestResponse `json:"setup_tests"`
}

type AccountExternalLoggingResponse struct {
	Code    string                     `json:"code"`
	Message string                     `json:"message"`
	Data    AccountExternalLoggingData `json:"data"`
}

type AccountExternalLoggingSetupTestsResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		SetupTests []common.SetupTestResponse `json:"setup_tests"`
	} `json:"data"`
}

type AccountExternalLoggingCreateRequest struct {
	Service       string                 `json:"service"`
	Enabled       *bool                  `json:"enabled,omitempty"`
	Config        map[string]interface{} `json:"config,omitempty"`
	RunSetupTests *bool                  `json:"run_setup_tests,omitempty"`
}

type AccountExternalLoggingUpdateRequest struct {
	Enabled       *bool                  `json:"enabled,omitempty"`
	Config        map[string]interface{} `json:"config,omitempty"`
	RunSetupTests *bool                  `json:"run_setup_tests,omitempty"`
}

func CreateAccountExternalLogging(ctx context.Context, client *fivetran.Client, userAgentSuffix string, req AccountExternalLoggingCreateRequest) (AccountExternalLoggingResponse, error) {
	var response AccountExternalLoggingResponse
	err := client.NewHttpService().DoWithUserAgentSuffix(ctx, userAgentSuffix, "POST", accountExternalLoggingPath, req, nil, 201, &response)
	return response, err
}

func ReadAccountExternalLogging(ctx context.Context, client *fivetran.Client, userAgentSuffix string) (AccountExternalLoggingResponse, error) {
	var response AccountExternalLoggingResponse
	err := client.NewHttpService().DoWithUserAgentSuffix(ctx, userAgentSuffix, "GET", accountExternalLoggingPath, nil, nil, 200, &response)
	return response, err
}

func UpdateAccountExternalLogging(ctx context.Context, client *fivetran.Client, userAgentSuffix string, req AccountExternalLoggingUpdateRequest) (AccountExternalLoggingResponse, error) {
	var response AccountExternalLoggingResponse
	err := client.NewHttpService().DoWithUserAgentSuffix(ctx, userAgentSuffix, "PATCH", accountExternalLoggingPath, req, nil, 200, &response)
	return response, err
}

func DeleteAccountExternalLogging(ctx context.Context, client *fivetran.Client, userAgentSuffix string) (common.CommonResponse, error) {
	var response common.CommonResponse
	err := client.NewHttpService().DoWithUserAgentSuffix(ctx, userAgentSuffix, "DELETE", accountExternalLoggingPath, nil, nil, 200, &response)
	return response, err
}

func RunAccountExternalLoggingSetupTests(ctx context.Context, client *fivetran.Client, userAgentSuffix string) (AccountExternalLoggingSetupTestsResponse, error) {
	var response AccountExternalLoggingSetupTestsResponse
	err := client.NewHttpService().DoWithUserAgentSuffix(ctx, userAgentSuffix, "POST", accountExternalLoggingPath+"/test", nil, nil, 200, &response)
	return response, err
}

// ReadLogServiceMetadata fetches the config metadata of a log service type. The response has the same
// shape as connector metadata (without `auth`), so the SDK's metadata types are reused.
func ReadLogServiceMetadata(ctx context.Context, client *fivetran.Client, userAgentSuffix, service string) (metadata.ConnectorMetadataResponse, error) {
	var response metadata.ConnectorMetadataResponse
	err := client.NewHttpService().DoWithUserAgentSuffix(ctx, userAgentSuffix, "GET", fmt.Sprintf("/metadata/log-service-types/%v", service), nil, nil, 200, &response)
	return response, err
}
