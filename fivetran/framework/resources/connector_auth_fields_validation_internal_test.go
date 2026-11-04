package resources

import (
	"testing"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/common"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestRejectAuthFieldsInConfig(t *testing.T) {
	common.LoadConfigFieldsMap()
	common.LoadAuthFieldsMap()

	cases := []struct {
		name      string
		service   string
		field     string
		value     attr.Value
		wantError bool
	}{
		{
			name:      "non-sensitive auth-only field is rejected",
			service:   "google_play",
			field:     "authentication_method",
			value:     types.StringValue("FIVETRAN_SERVICE_ACCOUNT"),
			wantError: true,
		},
		{
			name:      "unknown auth-only field is rejected",
			service:   "google_play",
			field:     "authentication_method",
			value:     types.StringUnknown(),
			wantError: true,
		},
		{
			name:    "null auth-only field is accepted",
			service: "google_play",
			field:   "authentication_method",
			value:   types.StringNull(),
		},
		{
			name:    "sensitive auth-only field is accepted",
			service: "asana",
			field:   "refresh_token",
			value:   types.StringValue("token"),
		},
		{
			name:    "field expected in both config and auth is accepted",
			service: "share_point",
			field:   "tenant_id",
			value:   types.StringValue("tenant"),
		},
		{
			name:    "same field for a service that expects it in config is accepted",
			service: "adobe_analytics",
			field:   "authentication_method",
			value:   types.StringValue("OAUTH"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			config := types.ObjectValueMust(
				map[string]attr.Type{tc.field: types.StringType},
				map[string]attr.Value{tc.field: tc.value},
			)
			var diags diag.Diagnostics
			rejectAuthFieldsInConfig(config, types.StringValue(tc.service), path.Root("config"), &diags)
			if diags.HasError() != tc.wantError {
				t.Fatalf("expected error: %v, got diagnostics: %v", tc.wantError, diags)
			}
		})
	}
}
