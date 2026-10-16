package resources_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	fivetranSdk "github.com/fivetran/go-fivetran"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/common"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/resources"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

const unconfirmedWarningSummary = "Destination config values were not returned by the API"

// destinationWarningTestSetup builds the destination resource with a client backed by the shared mock HTTP client.
func destinationWarningTestSetup(t *testing.T) (resource.Resource, schema.Schema) {
	t.Helper()
	common.LoadConfigFieldsMap()
	common.LoadAuthFieldsMap()
	common.LoadDestinationFieldsMap()
	common.LoadExternalLoggingFieldsMap()

	client := fivetranSdk.New("test_key", "test_secret")
	client.BaseURL("https://api.fivetran.com/v1")
	client.SetHttpClient(tfmock.MockClient())

	r := resources.Destination()
	configureResp := &resource.ConfigureResponse{}
	r.(resource.ResourceWithConfigure).Configure(context.Background(),
		resource.ConfigureRequest{ProviderData: client}, configureResp)
	if configureResp.Diagnostics.HasError() {
		t.Fatalf("configure failed: %v", configureResp.Diagnostics)
	}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)
	return r, schemaResp.Schema
}

// destinationWarningTestConfig builds the config object. Attributes that are not in values are null.
func destinationWarningTestConfig(t *testing.T, s schema.Schema, values map[string]attr.Value) types.Object {
	t.Helper()
	ctx := context.Background()
	cfgType := s.Blocks["config"].Type().(basetypes.ObjectType)

	attrs := make(map[string]attr.Value, len(cfgType.AttrTypes))
	for name, attrType := range cfgType.AttrTypes {
		if v, ok := values[name]; ok {
			attrs[name] = v
			continue
		}
		nullValue, err := attrType.ValueFromTerraform(ctx, tftypes.NewValue(attrType.TerraformType(ctx), nil))
		if err != nil {
			t.Fatalf("cannot build null for %s: %v", name, err)
		}
		attrs[name] = nullValue
	}

	obj, diags := types.ObjectValue(cfgType.AttrTypes, attrs)
	if diags.HasError() {
		t.Fatalf("cannot build config object: %v", diags)
	}
	return obj
}

// destinationWarningTestModel builds a raw Terraform value from a model, using the resource schema.
func destinationWarningTestModel(t *testing.T, s schema.Schema, data model.DestinationResourceModel) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()

	timeoutsAttrTypes := s.Blocks["timeouts"].Type().(attr.TypeWithAttributeTypes).AttributeTypes()
	data.Timeouts = timeoutsNull(timeoutsAttrTypes)

	plan := tfsdk.Plan{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)}
	if diags := plan.Set(ctx, &data); diags.HasError() {
		t.Fatalf("cannot build plan: %v", diags)
	}
	return plan
}

func destinationWarningTestState(s schema.Schema) tfsdk.State {
	ctx := context.Background()
	return tfsdk.State{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)}
}

// Responses for a redshift destination. The API returns no cluster_id, and the password is returned only in the second one.
const redshiftWarningResponse = `
{
	"id":"destination_id",
	"group_id":"group_id",
	"service":"redshift",
	"region":"GCP_US_EAST4",
	"time_zone_offset":"0",
	"setup_status":"connected",
	"daylight_saving_time_enabled":false,
	"networking_method":"Directly",
	"config":{
		"host":"host",
		"port":"5439",
		"database":"database",
		"user":"user",
		"connection_type":"Directly"
	}
}
`

const redshiftWarningResponseWithPassword = `
{
	"id":"destination_id",
	"group_id":"group_id",
	"service":"redshift",
	"region":"GCP_US_EAST4",
	"time_zone_offset":"0",
	"setup_status":"connected",
	"daylight_saving_time_enabled":false,
	"networking_method":"Directly",
	"config":{
		"host":"host",
		"port":"5439",
		"database":"database",
		"user":"user",
		"password":"******",
		"connection_type":"Directly"
	}
}
`

func timeoutsNull(attrTypes map[string]attr.Type) timeouts.Value {
	return timeouts.Value{Object: types.ObjectNull(attrTypes)}
}

// assertUnconfirmedWarningPaths checks there is exactly one unconfirmed-config warning, and that it names all paths.
func assertUnconfirmedWarningPaths(t *testing.T, diags diag.Diagnostics, paths ...string) {
	t.Helper()

	var matching []diag.Diagnostic
	for _, w := range diags.Warnings() {
		if w.Summary() == unconfirmedWarningSummary {
			matching = append(matching, w)
		}
	}
	if len(matching) != 1 {
		t.Fatalf("expected exactly one unconfirmed-config warning, got %d; diagnostics: %v", len(matching), diags)
	}

	for _, p := range paths {
		if !strings.Contains(matching[0].Detail(), p) {
			t.Errorf("warning must name %q, detail: %s", p, matching[0].Detail())
		}
	}
}

// assertNoSecretInDiagnostics checks that no diagnostic summary or detail contains the secret value.
func assertNoSecretInDiagnostics(t *testing.T, diags diag.Diagnostics, secret string) {
	t.Helper()

	for _, d := range diags {
		if strings.Contains(d.Summary(), secret) || strings.Contains(d.Detail(), secret) {
			t.Errorf("diagnostic leaks a configured value: %s: %s", d.Summary(), d.Detail())
		}
	}
}

func assertWarningDoesNotNamePath(t *testing.T, diags diag.Diagnostics, path string) {
	t.Helper()

	for _, w := range diags.Warnings() {
		if w.Summary() == unconfirmedWarningSummary && strings.Contains(w.Detail(), path) {
			t.Errorf("warning must not name %q, detail: %s", path, w.Detail())
		}
	}
}

func assertNoUnconfirmedWarning(t *testing.T, diags diag.Diagnostics) {
	t.Helper()

	for _, w := range diags.Warnings() {
		if w.Summary() == unconfirmedWarningSummary {
			t.Errorf("unexpected unconfirmed-config warning: %s", w.Detail())
		}
	}
}

func assertStateConfigString(t *testing.T, state tfsdk.State, name, expected string) {
	t.Helper()

	var value types.String
	if diags := state.GetAttribute(context.Background(), path.Root("config").AtName(name), &value); diags.HasError() {
		t.Fatalf("cannot read config.%s from state: %v", name, diags)
	}
	if value.ValueString() != expected || value.IsNull() {
		t.Errorf("config.%s in state: got %v, want %q", name, value, expected)
	}
}

func TestDestinationCreateWarnsForUnreturnedConfigWithoutSecrets(t *testing.T) {
	ctx := context.Background()
	r, s := destinationWarningTestSetup(t)
	tfmock.MockClient().Reset()

	tfmock.MockClient().When(http.MethodPost, "/v1/destinations").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := tfmock.RequestBodyToJson(t, req)
			config := body["config"].(map[string]interface{})
			tfmock.AssertKeyExistsAndHasValue(t, config, "cluster_id", "cluster_id")
			tfmock.AssertKeyExistsAndHasValue(t, config, "password", "password-secret-value")

			// cluster_id and password are accepted but not returned. The sensitive password is kept
			// by the read logic as before, so only cluster_id is unconfirmed.
			return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success",
				tfmock.CreateMapFromJsonString(t, redshiftWarningResponse)), nil
		},
	)

	plan := destinationWarningTestModel(t, s, model.DestinationResourceModel{
		GroupId:        types.StringValue("group_id"),
		Service:        types.StringValue("redshift"),
		Region:         types.StringValue("GCP_US_EAST4"),
		TimeZoneOffset: types.StringValue("0"),
		RunSetupTests:  types.BoolValue(false),
		Config: destinationWarningTestConfig(t, s, map[string]attr.Value{
			"host":       types.StringValue("host"),
			"port":       types.Int64Value(5439),
			"database":   types.StringValue("database"),
			"user":       types.StringValue("user"),
			"password":   types.StringValue("password-secret-value"),
			"cluster_id": types.StringValue("cluster_id"),
		}),
	})

	resp := &resource.CreateResponse{State: destinationWarningTestState(s)}
	r.Create(ctx, resource.CreateRequest{Plan: plan, Config: tfsdk.Config{Schema: s, Raw: plan.Raw}}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("create must not fail: %v", resp.Diagnostics)
	}

	assertUnconfirmedWarningPaths(t, resp.Diagnostics, "config.cluster_id")
	assertWarningDoesNotNamePath(t, resp.Diagnostics, "config.password")
	assertNoSecretInDiagnostics(t, resp.Diagnostics, "password-secret-value")
	assertStateConfigString(t, resp.State, "cluster_id", "cluster_id")
	assertStateConfigString(t, resp.State, "password", "password-secret-value")
}

func TestDestinationUpdateWarnsForUnreturnedConfigWithoutSecrets(t *testing.T) {
	ctx := context.Background()
	r, s := destinationWarningTestSetup(t)
	tfmock.MockClient().Reset()

	tfmock.MockClient().When(http.MethodPatch, "/v1/destinations/destination_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := tfmock.RequestBodyToJson(t, req)
			config := body["config"].(map[string]interface{})
			tfmock.AssertKeyExistsAndHasValue(t, config, "cluster_id", "cluster_id")

			// cluster_id is accepted but not returned. password is still returned, so only cluster_id is unconfirmed.
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success",
				tfmock.CreateMapFromJsonString(t, redshiftWarningResponseWithPassword)), nil
		},
	)

	stateData := model.DestinationResourceModel{
		Id:             types.StringValue("destination_id"),
		GroupId:        types.StringValue("group_id"),
		Service:        types.StringValue("redshift"),
		Region:         types.StringValue("GCP_US_EAST4"),
		TimeZoneOffset: types.StringValue("0"),
		RunSetupTests:  types.BoolValue(false),
		Config: destinationWarningTestConfig(t, s, map[string]attr.Value{
			"host":     types.StringValue("host"),
			"port":     types.Int64Value(5439),
			"database": types.StringValue("database"),
			"user":     types.StringValue("user"),
			"password": types.StringValue("password-secret-value"),
		}),
	}
	state := destinationWarningTestModel(t, s, stateData)

	planData := stateData
	planData.Config = destinationWarningTestConfig(t, s, map[string]attr.Value{
		"host":       types.StringValue("host"),
		"port":       types.Int64Value(5439),
		"database":   types.StringValue("database"),
		"user":       types.StringValue("user"),
		"password":   types.StringValue("password-secret-value"),
		"cluster_id": types.StringValue("cluster_id"),
	})
	plan := destinationWarningTestModel(t, s, planData)

	resp := &resource.UpdateResponse{State: destinationWarningTestState(s)}
	r.Update(ctx, resource.UpdateRequest{
		Plan:   plan,
		State:  tfsdk.State{Schema: s, Raw: state.Raw},
		Config: tfsdk.Config{Schema: s, Raw: plan.Raw},
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("update must not fail: %v", resp.Diagnostics)
	}

	assertUnconfirmedWarningPaths(t, resp.Diagnostics, "config.cluster_id")
	assertNoSecretInDiagnostics(t, resp.Diagnostics, "password-secret-value")
	assertStateConfigString(t, resp.State, "cluster_id", "cluster_id")
}

// An update that sends no config must not report config fields that the response omits.
func TestDestinationUpdateWithoutConfigChangesDoesNotWarn(t *testing.T) {
	ctx := context.Background()
	r, s := destinationWarningTestSetup(t)
	tfmock.MockClient().Reset()

	patchHandler := tfmock.MockClient().When(http.MethodPatch, "/v1/destinations/destination_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := tfmock.RequestBodyToJson(t, req)
			if body["config"] != nil {
				t.Errorf("config must not be sent, body: %v", body)
			}

			// cluster_id is not returned, but it was not sent either
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success",
				tfmock.CreateMapFromJsonString(t, strings.Replace(redshiftWarningResponse, `"time_zone_offset":"0"`, `"time_zone_offset":"+3"`, 1))), nil
		},
	)

	stateData := model.DestinationResourceModel{
		Id:             types.StringValue("destination_id"),
		GroupId:        types.StringValue("group_id"),
		Service:        types.StringValue("redshift"),
		Region:         types.StringValue("GCP_US_EAST4"),
		TimeZoneOffset: types.StringValue("0"),
		RunSetupTests:  types.BoolValue(false),
		Config: destinationWarningTestConfig(t, s, map[string]attr.Value{
			"host":       types.StringValue("host"),
			"port":       types.Int64Value(5439),
			"database":   types.StringValue("database"),
			"user":       types.StringValue("user"),
			"cluster_id": types.StringValue("cluster_id"),
		}),
	}
	state := destinationWarningTestModel(t, s, stateData)

	planData := stateData
	planData.TimeZoneOffset = types.StringValue("+3")
	plan := destinationWarningTestModel(t, s, planData)

	resp := &resource.UpdateResponse{State: destinationWarningTestState(s)}
	r.Update(ctx, resource.UpdateRequest{
		Plan:   plan,
		State:  tfsdk.State{Schema: s, Raw: state.Raw},
		Config: tfsdk.Config{Schema: s, Raw: plan.Raw},
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("update must not fail: %v", resp.Diagnostics)
	}
	tfmock.AssertEqual(t, patchHandler.Interactions, 1)

	assertNoUnconfirmedWarning(t, resp.Diagnostics)
	assertStateConfigString(t, resp.State, "cluster_id", "cluster_id")
}
