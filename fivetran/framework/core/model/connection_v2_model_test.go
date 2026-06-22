package model_test

import (
	"context"
	"testing"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestConnectionV2ResourceModelTfsdkShape(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	config, diags := types.ObjectValue(
		map[string]attr.Type{
			"schema": types.StringType,
			"table":  types.StringType,
		},
		map[string]attr.Value{
			"schema": types.StringValue("app"),
			"table":  types.StringValue("events"),
		},
	)
	if diags.HasError() {
		t.Fatalf("building config: %v", diags)
	}

	input := model.ConnectionV2ResourceModel{
		Id:                      types.StringValue("connection_id"),
		Name:                    types.StringValue("app"),
		ConnectedBy:             types.StringValue("user_id"),
		CreatedAt:               types.StringValue("2026-06-18T10:00:00Z"),
		GroupId:                 types.StringValue("group_id"),
		Service:                 types.StringValue("postgres"),
		Config:                  types.DynamicValue(config),
		Auth:                    types.DynamicNull(),
		SucceededAt:             types.StringValue("2026-06-18T11:00:00Z"),
		FailedAt:                types.StringNull(),
		ServiceVersion:          types.StringValue("1"),
		SyncFrequency:           types.Int64Value(60),
		ScheduleType:            types.StringValue("auto"),
		PauseAfterTrial:         types.BoolValue(false),
		DailySyncTime:           types.StringNull(),
		ProxyAgentId:            types.StringNull(),
		NetworkingMethod:        types.StringValue("Directly"),
		HybridDeploymentAgentId: types.StringNull(),
		PrivateLinkId:           types.StringNull(),
		DataDelaySensitivity:    types.StringValue("NORMAL"),
		DataDelayThreshold:      types.Int64Value(0),
		RunSetupTests:           types.BoolValue(false),
		TrustCertificates:       types.BoolValue(false),
		TrustFingerprints:       types.BoolValue(false),
		Status:                  types.ObjectNull(model.ConnectionV2StatusAttrTypes()),
	}

	var object types.Object
	diags = tfsdk.ValueFrom(ctx, input, types.ObjectType{AttrTypes: model.ConnectionV2ResourceModelAttrTypes()}, &object)
	if diags.HasError() {
		t.Fatalf("ValueFrom diagnostics: %v", diags)
	}

	var output model.ConnectionV2ResourceModel
	diags = tfsdk.ValueAs(ctx, object, &output)
	if diags.HasError() {
		t.Fatalf("ValueAs diagnostics: %v", diags)
	}

	if output.Service.ValueString() != "postgres" {
		t.Fatalf("unexpected service: got %q", output.Service.ValueString())
	}
	if output.Config.IsNull() || output.Config.IsUnknown() {
		t.Fatal("expected config dynamic value to round-trip")
	}
	if !output.Auth.IsNull() {
		t.Fatal("expected null auth dynamic value to round-trip")
	}
	if output.Status.IsUnknown() {
		t.Fatal("expected status object to keep a known null value")
	}
}
