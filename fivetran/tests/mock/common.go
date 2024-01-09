package mock

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"testing"

	fivetranSdk "github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/tests/mock"

	"github.com/fivetran/terraform-provider-fivetran/fivetran"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var client *fivetranSdk.Client
var mockClient *mock.HttpClient

// var testProviders map[string]*schema.Provider
var testProviderSdk *schema.Provider
var testProvioderFramework provider.Provider

func MockClient() *mock.HttpClient {
	return mockClient
}

var (
	TEST_KEY    = "test_key"
	TEST_SECRET = "test_secret"
)

var ProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"fivetran-provider": func() (tfprotov6.ProviderServer, error) {
		ctx := context.Background()

		upgradedSdkServer, err := tf5to6server.UpgradeServer(
			ctx,
			testProviderSdk.GRPCProvider, // Example terraform-plugin-sdk provider
		)

		if err != nil {
			return nil, err
		}

		providers := []func() tfprotov6.ProviderServer{
			providerserver.NewProtocol6(testProvioderFramework),
			func() tfprotov6.ProviderServer {
				return upgradedSdkServer
			},
		}

		muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)

		if err != nil {
			return nil, err
		}

		return muxServer.ProviderServer(), nil
	},
}

func init() {
	client = fivetranSdk.New(TEST_KEY, TEST_SECRET)
	mockClient = mock.NewHttpClient()
	client.BaseURL("https://api.fivetran.com/v1")
	client.SetHttpClient(mockClient)

	testProviderSdk = fivetran.Provider()
	testProviderSdk.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return client, diag.Diagnostics{}
	}

	testProvioderFramework = framework.FivetranProviderMock(mockClient)

	if os.Getenv("TF_ACC") == "" {
		// These are the mock tests, so we can freely set the TF_ACC environment variable
		os.Setenv("TF_ACC", "True")
	}

	if os.Getenv("FIVETRAN_APIKEY") == "" {
		os.Setenv("FIVETRAN_APIKEY", TEST_KEY)
		os.Setenv("FIVETRAN_APISECRET", TEST_SECRET)
	}
}

func requestBodyToJson(t *testing.T, req *http.Request) map[string]interface{} {
	t.Helper()

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		t.Errorf("requestBodyToJson, cannot read request body: %s", err)
	}
	req.Body.Close()
	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	result := map[string]interface{}{}
	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		t.Errorf("requestBodyToJson, cannot parse request body: %s", err)
	}

	return result
}

func RequestBodyToJson(t *testing.T, req *http.Request) map[string]interface{} {
	return requestBodyToJson(t, req)
}

func fivetranResponse(t *testing.T, req *http.Request, statusCode string, code int, message string,
	data map[string]interface{}) *http.Response {

	t.Helper()

	respBody := map[string]interface{}{
		"code": statusCode,
	}

	if message != "" {
		respBody["message"] = message
	}

	if data != nil {
		respBody["data"] = data
	}

	respBodyJson, err := json.Marshal(respBody)
	if err != nil {
		t.Errorf("fivetranSuccessResponse, cannot encode JSON: %s", err)
	}

	response := mock.NewResponse(req, code, string(respBodyJson))
	return response
}

func FivetranSuccessResponse(t *testing.T, req *http.Request, code int, message string,
	data map[string]interface{}) *http.Response {
	return fivetranSuccessResponse(t, req, code, message, data)
}

func fivetranSuccessResponse(t *testing.T, req *http.Request, code int, message string,
	data map[string]interface{}) *http.Response {

	return fivetranResponse(t, req, "Success", code, message, data)
}

func printError(t *testing.T, actual interface{}, expected interface{}) {
	t.Helper()
	t.Errorf("Expected: %s"+
		"\n     but: <%s>\n",
		fmt.Sprintf("value equal to <%v>", expected),
		fmt.Sprintf("%v", actual),
	)
}

func printErrorWithMessage(t *testing.T, actual, expected interface{}, message string) {
	t.Helper()
	t.Errorf("%s \n Expected: %s"+
		"\n     but: <%s>\n",
		message,
		fmt.Sprintf("value equal to <%v>", expected),
		fmt.Sprintf("%v", actual),
	)
}

func isEmpty(actual interface{}) bool {
	if actual == nil {
		return true
	} else if actualValue, ok := actual.(string); ok {
		return actualValue == ""
	} else if reflect.ValueOf(actual).Len() == 0 {
		return true
	}

	return false
}

func AssertEqual(t *testing.T, actual interface{}, expected interface{}) {
	assertEqual(t, actual, expected)
}

func assertEqual(t *testing.T, actual interface{}, expected interface{}) {
	t.Helper()

	if !reflect.DeepEqual(expected, actual) {
		printError(t, actual, expected)
	}
}

func AssertEmpty(t *testing.T, actual interface{}) {
	assertEmpty(t, actual)
}

func assertEmpty(t *testing.T, actual interface{}) {
	t.Helper()

	if !isEmpty(actual) {
		printError(t, actual, "empty value")
	}
}

func AssertNotEmpty(t *testing.T, actual interface{}) {
	assertNotEmpty(t, actual)
}

func assertNotEmpty(t *testing.T, actual interface{}) {
	t.Helper()

	if isEmpty(actual) {
		printError(t, actual, "none-empty value")
	}
}

func AssertKeyDoesNotExist(t *testing.T, source map[string]interface{}, key string) {
	t.Helper()

	if _, ok := source[key]; ok {
		printError(t, key, "no such key in given map")
	}
}

func assertKeyExists(t *testing.T, source map[string]interface{}, key string) interface{} {
	t.Helper()

	if v, ok := source[key]; !ok {
		printError(t, key, "key represented in given map")
		return nil
	} else {
		return v
	}
}

func AssertKeyExists(t *testing.T, source map[string]interface{}, key string) interface{} {
	return assertKeyExists(t, source, key)
}

func assertArrayItems(t *testing.T, source []interface{}, expected []interface{}) {
	t.Helper()

	if len(source) != len(expected) {
		printErrorWithMessage(t, len(source), len(expected), "Array size mismatch")
		return
	}
	for _, a := range source {
		if !contains(expected, a) {
			printErrorWithMessage(t, a, "", "Expected value not found in provided array")
			return
		}
	}
}

func contains(s []interface{}, e interface{}) bool {
	for _, a := range s {
		if reflect.DeepEqual(a, e) {
			return true
		}
	}
	return false
}

func assertKeyExistsAndHasValue(t *testing.T, source map[string]interface{}, key string, value interface{}) {
	t.Helper()

	if v, ok := source[key]; !ok || v != value {
		if !ok {
			printError(t, key, "key not found in source")
		} else {
			printError(t, v, value)
		}
	}
}
func AssertKeyExistsAndHasValue(t *testing.T, source map[string]interface{}, key string, value interface{}) {
	assertKeyExistsAndHasValue(t, source, key, value)
}

func CreateMapFromJsonString(t *testing.T, schemaJson string) map[string]interface{} {
	return createMapFromJsonString(t, schemaJson)
}

func createMapFromJsonString(t *testing.T, schemaJson string) map[string]interface{} {
	result := map[string]interface{}{}
	err := json.Unmarshal([]byte(schemaJson), &result)
	if err != nil {
		t.Errorf("requestBodyToJson, cannot parse request body: %s", err)
	}
	return result
}
