package resources

import (
	"context"
	"sync"
	"testing"

	"github.com/fivetran/go-fivetran/metadata"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestAccountExternalLoggingV2ModifyPlanOnCreateConfig(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	emptyObject, diags := core.MapToDynamic(ctx, map[string]interface{}{})
	if diags.HasError() {
		t.Fatalf("empty config diagnostics: %v", diags)
	}
	apiKeyOnly, diags := core.MapToDynamic(ctx, map[string]interface{}{"api_key": "secret"})
	if diags.HasError() {
		t.Fatalf("config diagnostics: %v", diags)
	}

	tests := []struct {
		name       string
		config     types.Dynamic
		wantErrors int
	}{
		{name: "unknown dynamic is deferred", config: types.DynamicUnknown()},
		{name: "unknown underlying object is deferred", config: types.DynamicValue(types.ObjectUnknown(map[string]attr.Type{}))},
		{name: "known empty object reports missing required field", config: emptyObject, wantErrors: 1},
		{name: "known config with required field is accepted", config: apiKeyOnly},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cache := &sync.Map{}
			cache.Store(logServiceMetadataCachePrefix+"datadog_log", &metadata.ConnectorMetadata{
				Config: metadata.Property{
					Required: []string{"api_key"},
					Properties: map[string]*metadata.Property{
						"api_key": {Type: "string"},
					},
				},
			})

			r := &accountExternalLoggingV2{}
			var configureResp resource.ConfigureResponse
			r.Configure(ctx, resource.ConfigureRequest{
				ProviderData: &core.ProviderResourceData{MetadataCache: cache},
			}, &configureResp)
			if configureResp.Diagnostics.HasError() {
				t.Fatalf("configure diagnostics: %v", configureResp.Diagnostics)
			}

			schema := fivetranSchema.AccountExternalLoggingV2ResourceSchema()
			plan := model.AccountExternalLoggingV2ResourceModel{
				Id:            types.StringUnknown(),
				Service:       types.StringValue("datadog_log"),
				Enabled:       types.BoolValue(true),
				RunSetupTests: types.BoolValue(false),
				Config:        tt.config,
			}
			var object types.Object
			if d := tfsdk.ValueFrom(ctx, plan, schema.Type(), &object); d.HasError() {
				t.Fatalf("ValueFrom diagnostics: %v", d)
			}
			raw, err := object.ToTerraformValue(ctx)
			if err != nil {
				t.Fatalf("converting plan to Terraform value: %v", err)
			}

			req := resource.ModifyPlanRequest{
				Plan:  tfsdk.Plan{Raw: raw, Schema: schema},
				State: tfsdk.State{Raw: tftypes.NewValue(schema.Type().TerraformType(ctx), nil), Schema: schema},
			}
			var resp resource.ModifyPlanResponse
			r.ModifyPlan(ctx, req, &resp)

			assertErrorCount(t, resp.Diagnostics, tt.wantErrors)
		})
	}
}
