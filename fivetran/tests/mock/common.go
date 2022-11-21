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

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var client *fivetranSdk.Client
var mockClient *mock.HttpClient
var testProviders map[string]*schema.Provider

var (
	TEST_KEY    = "test_key"
	TEST_SECRET = "test_secret"
)

func init() {
	client = fivetranSdk.New(TEST_KEY, TEST_SECRET)
	mockClient = mock.NewHttpClient()
	client.BaseURL("https://api.fivetran.com/v1")
	client.SetHttpClient(mockClient)

	provider := fivetran.Provider()
	provider.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return client, diag.Diagnostics{}
	}

	testProviders = map[string]*schema.Provider{
		"fivetran-provider": provider,
	}

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

func assertEqual(t *testing.T, actual interface{}, expected interface{}) {
	t.Helper()

	if !reflect.DeepEqual(expected, actual) {
		printError(t, actual, expected)
	}
}

func assertEmpty(t *testing.T, actual interface{}) {
	t.Helper()

	if !isEmpty(actual) {
		printError(t, actual, "empty value")
	}
}

func assertNotEmpty(t *testing.T, actual interface{}) {
	t.Helper()

	if isEmpty(actual) {
		printError(t, actual, "none-empty value")
	}
}

func createMapFromJsonString(t *testing.T, schemaJson string) map[string]interface{} {
	result := map[string]interface{}{}
	err := json.Unmarshal([]byte(schemaJson), &result)
	if err != nil {
		t.Errorf("requestBodyToJson, cannot parse request body: %s", err)
	}
	return result
}
