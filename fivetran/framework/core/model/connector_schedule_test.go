package model

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestReadScheduleFromResponsePreservesExistingScheduleWhenResponseHasNoSchedule(t *testing.T) {
	existing := types.ObjectValueMust(connectorScheduleBlockAttrTypes, map[string]attr.Value{
		"schedule_type": types.StringValue("INTERVAL"),
		"interval":      types.Int64Value(60),
		"time_of_day":   types.StringNull(),
		"days_of_week":  types.SetNull(types.StringType),
		"cron":          types.StringNull(),
	})

	got := readScheduleFromResponse(nil, existing)
	if got.IsNull() || got.IsUnknown() {
		t.Fatalf("got null/unknown schedule, want existing schedule")
	}

	if got.Attributes()["schedule_type"].(types.String).ValueString() != "INTERVAL" {
		t.Fatalf("got schedule_type %q, want INTERVAL", got.Attributes()["schedule_type"].(types.String).ValueString())
	}
	if got.Attributes()["interval"].(types.Int64).ValueInt64() != 60 {
		t.Fatalf("got interval %v, want 60", got.Attributes()["interval"].(types.Int64).ValueInt64())
	}
}

func TestReadScheduleFromResponseReturnsNullWhenResponseAndExistingHaveNoSchedule(t *testing.T) {
	got := readScheduleFromResponse(nil, types.ObjectNull(connectorScheduleBlockAttrTypes))
	if !got.IsNull() {
		t.Fatalf("got non-null schedule, want null")
	}
}

func TestReadScheduleFromResponseReturnsNullForFlexibleExistingScheduleWhenResponseHasNoSchedule(t *testing.T) {
	existing := types.ObjectValueMust(connectorScheduleBlockAttrTypes, map[string]attr.Value{
		"schedule_type": types.StringValue("CRON"),
		"interval":      types.Int64Null(),
		"time_of_day":   types.StringNull(),
		"days_of_week":  types.SetNull(types.StringType),
		"cron":          types.StringValue("0 9 * * 1-5"),
	})

	got := readScheduleFromResponse(nil, existing)
	if !got.IsNull() {
		t.Fatalf("got non-null schedule, want null")
	}
}
