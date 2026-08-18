package model

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestConnectionV2ResourceModel_FieldShape(t *testing.T) {
	t.Parallel()

	codeMsgAttrTypes := map[string]attr.Type{
		"code":    types.StringType,
		"message": types.StringType,
	}
	emptySet, _ := types.SetValue(types.ObjectType{AttrTypes: codeMsgAttrTypes}, []attr.Value{})
	statusVal, _ := types.ObjectValue(StatusAttrTypes, map[string]attr.Value{
		"setup_state":        types.StringValue("connected"),
		"sync_state":         types.StringValue("scheduled"),
		"update_state":       types.StringValue("on_schedule"),
		"is_historical_sync": types.BoolValue(false),
		"tasks":              emptySet,
		"warnings":           emptySet,
	})

	inner, _ := types.ObjectValue(
		map[string]attr.Type{"host": types.StringType},
		map[string]attr.Value{"host": types.StringValue("db.example.com")},
	)

	m := ConnectionV2ResourceModel{
		Id:                      types.StringValue("conn_123"),
		Service:                 types.StringValue("postgres"),
		GroupId:                 types.StringValue("grp_456"),
		DestinationSchema:       types.StringValue("my_schema"),
		Config:                  types.DynamicValue(inner),
		Auth:                    types.DynamicNull(),
		Name:                    types.StringValue("my_schema"),
		ConnectedBy:             types.StringValue("user_789"),
		CreatedAt:               types.StringValue("2026-01-01T00:00:00Z"),
		SucceededAt:             types.StringNull(),
		FailedAt:                types.StringNull(),
		ServiceVersion:          types.StringValue("1"),
		Status:                  statusVal,
		Paused:                  types.BoolValue(false),
		SyncFrequency:           types.Int64Value(360),
		ScheduleType:            types.StringValue("auto"),
		DailySyncTime:           types.StringNull(),
		PauseAfterTrial:         types.BoolValue(false),
		NetworkingMethod:        types.StringValue("Directly"),
		ProxyAgentId:            types.StringNull(),
		PrivateLinkId:           types.StringNull(),
		HybridDeploymentAgentId: types.StringNull(),
		DataDelaySensitivity:    types.StringValue("NORMAL"),
		DataDelayThreshold:      types.Int64Value(0),
		RunSetupTests:           types.BoolValue(false),
		TrustCertificates:       types.BoolValue(false),
		TrustFingerprints:       types.BoolValue(false),
	}

	if m.Id.ValueString() != "conn_123" {
		t.Errorf("Id: got %v, want conn_123", m.Id.ValueString())
	}
	if m.Service.ValueString() != "postgres" {
		t.Errorf("Service: got %v, want postgres", m.Service.ValueString())
	}
	if m.GroupId.ValueString() != "grp_456" {
		t.Errorf("GroupId: got %v, want grp_456", m.GroupId.ValueString())
	}
	if m.DestinationSchema.ValueString() != "my_schema" {
		t.Errorf("DestinationSchema: got %v, want my_schema", m.DestinationSchema.ValueString())
	}
	if m.Config.IsNull() {
		t.Error("Config: should not be null")
	}
	if !m.Auth.IsNull() {
		t.Error("Auth: should be null")
	}
	if m.SyncFrequency.ValueInt64() != 360 {
		t.Errorf("SyncFrequency: got %v, want 360", m.SyncFrequency.ValueInt64())
	}
	if m.Status.IsNull() {
		t.Error("Status: should not be null")
	}
	if m.DataDelaySensitivity.ValueString() != "NORMAL" {
		t.Errorf("DataDelaySensitivity: got %v, want NORMAL", m.DataDelaySensitivity.ValueString())
	}
}
